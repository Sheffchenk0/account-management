package app

import (
	"account-manager/config"
	v1 "account-manager/internal/controller/http/v1"
	"account-manager/internal/infrastructure/rabbitmq"
	repoPG "account-manager/internal/repo/postgres"
	"account-manager/internal/service"
	"account-manager/internal/worker"
	"account-manager/pkg/postgres"
	rmqPkg "account-manager/pkg/rabbitmq"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// TODO: move in cfg
const (
	exchangeName           = "account_events"
	outboxWorkerIntervalMs = 100
	outboxWorkerBatchSize  = 10
	readTimeOutS           = 10
	writeTimeOutS          = 10
)

func Run(cfg *config.Config) error {
	pgPool, err := postgres.New(cfg.PG.URL)
	if err != nil {
		return fmt.Errorf("postgres init: %w", err)
	}
	defer pgPool.Close()

	rmqConn, err := rmqPkg.New(cfg.RMQ.Url)
	if err != nil {
		return fmt.Errorf("rabbitmq init: %w", err)
	}
	defer rmqConn.Close()

	txMgr := postgres.NewManager(pgPool.Pool)

	accountRepo := repoPG.NewAccountRepo(txMgr)
	outboxRepo := repoPG.NewOutboxRepo(txMgr)
	transactionRepo := repoPG.NewTransactionRepo(txMgr)

	rmqPublisher, err := rabbitmq.NewPublisher(rmqConn, exchangeName)
	if err != nil {
		return fmt.Errorf("publisher init %w:", err)
	}

	transactionService := service.NewTransactionService(
		txMgr,
		accountRepo,
		transactionRepo,
		outboxRepo,
	)

	accountService := service.NewAccountService(txMgr, accountRepo)

	outboxProcessor := worker.NewOutboxProcessor(
		txMgr,
		outboxRepo,
		rmqPublisher,
		outboxWorkerIntervalMs*time.Millisecond,
		outboxWorkerBatchSize,
	)

	handler := gin.Default()
	handler.Use(gin.Recovery())

	v1Group := handler.Group("/v1")
	{
		transactionHandler := v1.NewTransactionHandler(transactionService)
		transactionHandler.RegisterRoutes(v1Group)

		accountHandler := v1.NewAccountHandler(accountService)
		accountHandler.RegisterRoutes(v1Group)
	}

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      handler,
		ReadTimeout:  readTimeOutS * time.Second,
		WriteTimeout: writeTimeOutS * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		outboxProcessor.Start(ctx)
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// TODO: add logs
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// TODO: add logs
	}

	return nil
}
