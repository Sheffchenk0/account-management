package v1

import (
	"account-manager/internal/entity"
	"account-manager/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createTransactionRequest struct {
	ID        string `json:"id" binding:"required,uuid"`
	AccountID string `json:"account_id" binding:"required,uuid"`
	Amount    int64  `json:"amount" binding:"required,gt=0"`
	Type      string `json:"type" binding:"required,oneof=credit debit"`
}

type TransactionHandler struct {
	service *service.TransactionService
}

func NewTransactionHandler(service *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service}
}

func (h *TransactionHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/transactions", h.Create)
}

// Create creates a new transaction
// @Summary Create transaction
// @Description Create a new credit or debit transaction for an account
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body createTransactionRequest true "Transaction creation request"
// @Success 201 {object} map[string]interface{} "Transaction created"
// @Success 200 {object} map[string]interface{} "Transaction already processed"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Account not found"
// @Failure 422 {object} map[string]interface{} "Insufficient funds"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /transactions [post]
func (h *TransactionHandler) Create(c *gin.Context) {
	const op = "TransactionHandler.Create"

	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction := entity.Transaction{
		ID:        uuid.MustParse(req.ID),
		AccountID: uuid.MustParse(req.AccountID),
		Amount:    req.Amount,
		Type:      req.Type,
	}

	err := h.service.CreateTransaction(c.Request.Context(), transaction)
	if err != nil {
		h.handleError(c, err, op)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
		"id":     req.ID,
	})
}

func (h *TransactionHandler) handleError(c *gin.Context, err error, op string) {
	if errors.Is(err, entity.ErrDuplicateTransaction) {
		c.JSON(http.StatusOK, gin.H{
			"status": "already_processed",
			"info":   "transaction with this id already exists",
		})
		return
	}

	switch {
	case errors.Is(err, entity.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "insufficient funds"})
	case errors.Is(err, entity.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
	case errors.Is(err, entity.ErrInvalidTransactionType):
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid transaction type"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
