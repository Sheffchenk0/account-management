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
	ID string `json:"id" binding:"required,uuid"`
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

// Create creates a new account
// @Summary Create account
// @Description Create a new account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body createAccountRequest true "Account creation request"
// @Success 201 {object} map[string]interface{} "Account created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	const op = "AccountHandler.Create"

	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account := entity.Account{
		ID: uuid.MustParse(req.ID),
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

// GetByID retrieves an account by ID
// @Summary Get account by ID
// @Description Retrieve account details by account ID
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} map[string]interface{} "Account details"
// @Failure 400 {object} map[string]interface{} "Invalid account ID format"
// @Failure 404 {object} map[string]interface{} "Account not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /accounts/{id} [get]
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
	case errors.Is(err, entity.ErrAccountNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
