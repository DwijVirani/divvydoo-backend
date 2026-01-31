package controllers

import (
	"net/http"

	"divvydoo/backend/internal/services"
	"divvydoo/backend/internal/utils"
	"divvydoo/backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
	authService auth.JWTService
}

func NewUserController(userService *services.UserService, authService auth.JWTService) *UserController {
	return &UserController{
		userService: userService,
		authService: authService,
	}
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	var req services.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := c.userService.CreateUser(ctx.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusCreated, user)
}

func (c *UserController) Login(ctx *gin.Context) {
	var req services.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := c.userService.ValidateCredentials(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := c.authService.GenerateToken(user.UserID, user.Email)
	if err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := services.LoginResponse{
		Token: token,
		User:  user,
	}

	utils.RespondWithJSON(ctx, http.StatusOK, response)
}

func (c *UserController) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if the requesting user is accessing their own data
	requestingUserID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Users can only access their own profile
	if requestingUserID.(string) != userID {
		utils.RespondWithError(ctx, http.StatusForbidden, "Access denied")
		return
	}

	user, err := c.userService.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if the requesting user is updating their own data
	requestingUserID, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if requestingUserID.(string) != userID {
		utils.RespondWithError(ctx, http.StatusForbidden, "Access denied")
		return
	}

	var req services.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := c.userService.UpdateUser(ctx.Request.Context(), userID, req)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	utils.RespondWithJSON(ctx, http.StatusOK, user)
}

func (c *UserController) LookupUser(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Query parameter 'q' is required (email or phone)")
		return
	}

	// Verify user is authenticated
	_, exists := ctx.Get("userID")
	if !exists {
		utils.RespondWithError(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := c.userService.LookupUser(ctx.Request.Context(), query)
	if err != nil {
		utils.RespondWithError(ctx, utils.GetStatusCode(err), err.Error())
		return
	}

	print("User", user)
	// Return limited user info for privacy
	utils.RespondWithJSON(ctx, http.StatusOK, gin.H{
		"user_id": user.UserID,
		"name":    user.Name,
		"email":   user.Email,
	})
}
