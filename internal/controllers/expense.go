package controllers

import (
	"net/http"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/services"
	"divvydoo/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ExpenseController struct {
	expenseService *services.ExpenseService
}

func NewExpenseController(expenseService *services.ExpenseService) *ExpenseController {
	return &ExpenseController{expenseService: expenseService}
}

func (c *ExpenseController) CreateExpense(ctx *gin.Context) {
	var expense models.Expense
	if err := ctx.ShouldBindJSON(&expense); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set creator ID from authenticated user
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}
	expense.CreatorID = userID.(string)

	createdExpense, err := c.expenseService.CreateExpense(ctx.Request.Context(), expense)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusCreated, createdExpense)
}

func (c *ExpenseController) GetExpense(ctx *gin.Context) {
	expenseID := ctx.Param("id")
	if expenseID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Expense ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	expense, err := c.expenseService.GetExpense(ctx.Request.Context(), expenseID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, expense)
}

func (c *ExpenseController) ListGroupExpenses(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	// Default pagination
	limit := int64(20)
	offset := int64(0)

	expenses, err := c.expenseService.GetGroupExpenses(ctx.Request.Context(), groupID, limit, offset)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, expenses)
}

func (c *ExpenseController) ListUserExpenses(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if the requesting user is accessing their own expenses
	requestingUserID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if requestingUserID.(string) != userID {
		utils.RespondWithError(ctx, http.StatusForbidden, "Access denied")
		return
	}

	// Default pagination
	limit := int64(20)
	offset := int64(0)

	expenses, err := c.expenseService.GetUserExpenses(ctx.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, expenses)
}
