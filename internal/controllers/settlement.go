package controllers

import (
	"net/http"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/services"
	"divvydoo/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type SettlementController struct {
	settlementService *services.SettlementService
}

func NewSettlementController(settlementService *services.SettlementService) *SettlementController {
	return &SettlementController{settlementService: settlementService}
}

func (c *SettlementController) CreateSettlement(ctx *gin.Context) {
	var req models.SettlementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// The authenticated user should be the one making the payment
	if req.FromUserID != userID.(string) {
		utils.RespondWithError(ctx, http.StatusForbidden, "Can only create settlements for yourself")
		return
	}

	settlement, err := c.settlementService.CreateSettlement(ctx.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusCreated, settlement)
}

func (c *SettlementController) GetSettlement(ctx *gin.Context) {
	settlementID := ctx.Param("id")
	if settlementID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Settlement ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	settlement, err := c.settlementService.GetSettlement(ctx.Request.Context(), settlementID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, settlement)
}

type CompleteSettlementRequest struct {
	TransactionID *string `json:"transaction_id,omitempty"`
}

func (c *SettlementController) CompleteSettlement(ctx *gin.Context) {
	settlementID := ctx.Param("id")
	if settlementID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Settlement ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CompleteSettlementRequest
	ctx.ShouldBindJSON(&req) // Optional body

	err := c.settlementService.CompleteSettlement(ctx.Request.Context(), settlementID, userID.(string), req.TransactionID)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{"message": "Settlement completed successfully"})
}

func (c *SettlementController) CancelSettlement(ctx *gin.Context) {
	settlementID := ctx.Param("id")
	if settlementID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Settlement ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := c.settlementService.CancelSettlement(ctx.Request.Context(), settlementID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{"message": "Settlement cancelled successfully"})
}

func (c *SettlementController) GetPendingSettlements(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	settlements, err := c.settlementService.GetPendingSettlements(ctx.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, settlements)
}
