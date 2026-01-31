package controllers

import (
	"net/http"

	"divvydoo/backend/internal/services"
	"divvydoo/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type BalanceController struct {
	balanceService *services.BalanceService
}

func NewBalanceController(balanceService *services.BalanceService) *BalanceController {
	return &BalanceController{balanceService: balanceService}
}

func (c *BalanceController) GetUserBalances(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if the requesting user is accessing their own balances
	requestingUserID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if requestingUserID.(string) != userID {
		utils.RespondWithError(ctx, http.StatusForbidden, "Access denied")
		return
	}

	balances, err := c.balanceService.GetUserBalances(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, balances)
}

func (c *BalanceController) GetGroupBalances(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	balances, err := c.balanceService.GetGroupBalances(ctx.Request.Context(), groupID)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, balances)
}
