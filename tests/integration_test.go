package tests

import (
	v1 "account-manager/internal/controller/http/v1"
	"account-manager/internal/entity"
	repoPG "account-manager/internal/repo/postgres"
	"account-manager/internal/service"
	pgPkg "account-manager/pkg/postgres"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCreateTransaction_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	migrationSQL, err := os.ReadFile("../migrations/0001_init_schema_up.sql")
	require.NoError(t, err)

	pgPool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pgPool.Close()

	cleanSQL := extractCleanSQL(string(migrationSQL))
	_, err = pgPool.Exec(ctx, cleanSQL)
	require.NoError(t, err)

	txMgr := pgPkg.NewManager(pgPool)

	accountRepo := repoPG.NewAccountRepo(txMgr)
	outboxRepo := repoPG.NewOutboxRepo(txMgr)
	transactionRepo := repoPG.NewTransactionRepo(txMgr)

	transactionService := service.NewTransactionService(
		txMgr,
		accountRepo,
		transactionRepo,
		outboxRepo,
	)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			t.Logf("Gin errors: %v", c.Errors)
		}
	})
	v1Group := router.Group("/v1")

	transactionHandler := v1.NewTransactionHandler(transactionService)
	transactionHandler.RegisterRoutes(v1Group)

	accountID := uuid.New()
	err = createTestAccount(ctx, pgPool, accountID, 1000)
	require.NoError(t, err)

	var checkBalance int64
	err = pgPool.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&checkBalance)
	require.NoError(t, err)
	require.Equal(t, int64(1000), checkBalance, "account balance should be 1000 after creation")

	t.Run("Success Credit", func(t *testing.T) {
		var existingBalance int64
		err := pgPool.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&existingBalance)
		require.NoError(t, err)
		t.Logf("Existing account balance before transaction: %d", existingBalance)

		transactionID := uuid.New()
		payload := map[string]interface{}{
			"id":         transactionID.String(),
			"account_id": accountID.String(),
			"amount":     500,
			"type":       "credit",
		}
		body, _ := json.Marshal(payload)
		t.Logf("Sending payload: %s", string(body))

		req, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		if w.Code != http.StatusCreated {
			t.Logf("Response body: %s", w.Body.String())
		}

		var updatedAccount entity.Account
		err = pgPool.QueryRow(ctx,
			"SELECT id, balance, created_at FROM accounts WHERE id = $1",
			accountID,
		).Scan(&updatedAccount.ID, &updatedAccount.Balance, &updatedAccount.CreatedAt)
		assert.NoError(t, err)
		assert.Equal(t, int64(1500), updatedAccount.Balance)

		var savedTransaction entity.Transaction
		err = pgPool.QueryRow(ctx,
			"SELECT id, account_id, amount, type FROM transactions WHERE id = $1",
			transactionID,
		).Scan(&savedTransaction.ID, &savedTransaction.AccountID, &savedTransaction.Amount, &savedTransaction.Type)
		assert.NoError(t, err)
		assert.Equal(t, transactionID, savedTransaction.ID)
		assert.Equal(t, accountID, savedTransaction.AccountID)
		assert.Equal(t, int64(500), savedTransaction.Amount)
		assert.Equal(t, "credit", savedTransaction.Type)
	})
}

func createTestAccount(ctx context.Context, pool *pgxpool.Pool, accountID uuid.UUID, balance int64) error {
	_, err := pool.Exec(ctx,
		"INSERT INTO accounts (id, balance) VALUES ($1, $2)",
		accountID, balance,
	)
	return err
}

func extractCleanSQL(migrationSQL string) string {
	var result string
	lines := []string{}
	inStatement := false

	for _, line := range splitLines(migrationSQL) {
		trimmed := line
		if trimmed == "-- +goose StatementBegin" {
			inStatement = true
			continue
		}
		if trimmed == "-- +goose StatementEnd" {
			inStatement = false
			continue
		}
		if trimmed == "-- +goose Up" {
			continue
		}
		if trimmed == "-- +goose Down" {
			break
		}
		if inStatement {
			lines = append(lines, line)
		}
	}

	for _, line := range lines {
		result += line + "\n"
	}

	return result
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
