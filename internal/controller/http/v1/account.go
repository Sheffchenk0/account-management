package v1

import (
	"account-manager/internal/entity"
	"account-manager/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createAccountRequest struct {
	ID       string `json:"id" binding:"required,uuid"`
	ClientID string `json:"client_id" binding:"required,uuid"`
}

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service}
}

func (h *AccountHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/accounts", h.Create)
	r.GET("/accounts/:id", h.GetByID)
}

func (h *AccountHandler) Create(c *gin.Context) {
	const op = "AccountHandler.Create"

	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account := entity.Account{
		ID:       uuid.MustParse(req.ID),
		ClientID: uuid.MustParse(req.ClientID),
	}

	createdAccount, err := h.service.CreateAccount(c.Request.Context(), account)
	if err != nil {
		h.handleError(c, err, op)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
		"id":     createdAccount.ID.String(),
	})
}

func (h *AccountHandler) GetByID(c *gin.Context) {
	const op = "AccountHandler.GetByID"

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id format"})
		return
	}

	account, err := h.service.GetAccount(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err, op)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      account.ID.String(),
		"balance": account.Balance,
	})
}

func (h *AccountHandler) handleError(c *gin.Context, err error, op string) {
	switch {
	case errors.Is(err, entity.ErrClientNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "client not found"})
	case errors.Is(err, entity.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
