package controllers

import (
	"net/http"

	"divvydoo/backend/internal/services"
	"divvydoo/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type GroupController struct {
	groupService *services.GroupService
}

func NewGroupController(groupService *services.GroupService) *GroupController {
	return &GroupController{groupService: groupService}
}

func (c *GroupController) CreateGroup(ctx *gin.Context) {
	var req services.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	group, err := c.groupService.CreateGroup(ctx.Request.Context(), userID.(string), req)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusCreated, group)
}

func (c *GroupController) GetGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	group, err := c.groupService.GetGroup(ctx.Request.Context(), groupID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, group)
}

func (c *GroupController) AddMember(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	var req services.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := c.groupService.AddMember(ctx.Request.Context(), groupID, userID.(string), req)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{"message": "Member added successfully"})
}

func (c *GroupController) RemoveMember(ctx *gin.Context) {
	groupID := ctx.Param("id")
	memberID := ctx.Param("memberId")
	if groupID == "" || memberID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID and Member ID are required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := c.groupService.RemoveMember(ctx.Request.Context(), groupID, userID.(string), memberID)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func (c *GroupController) LeaveGroup(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	err := c.groupService.LeaveGroup(ctx.Request.Context(), groupID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{"message": "Left group successfully"})
}

func (c *GroupController) GetMembers(ctx *gin.Context) {
	groupID := ctx.Param("id")
	if groupID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Group ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	members, err := c.groupService.GetMembers(ctx.Request.Context(), groupID, userID.(string))
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, members)
}
