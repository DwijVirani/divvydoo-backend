package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"divvydoo/backend/internal/config"
	"divvydoo/backend/internal/controllers"
	"divvydoo/backend/internal/middleware"
	"divvydoo/backend/internal/repositories"
	"divvydoo/backend/internal/services"
	"divvydoo/backend/pkg/auth"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Verify MongoDB connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(cfg.MongoDBName)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	groupRepo := repositories.NewGroupRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	balanceRepo := repositories.NewBalanceRepository(db)
	settlementRepo := repositories.NewSettlementRepository(db)

	// Initialize services
	authService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)
	userService := services.NewUserService(userRepo)
	groupService := services.NewGroupService(groupRepo, userRepo)
	expenseService := services.NewExpenseService(expenseRepo, balanceRepo, groupRepo, userRepo)
	balanceService := services.NewBalanceService(balanceRepo, expenseRepo, userRepo)
	settlementService := services.NewSettlementService(settlementRepo, balanceRepo, userRepo)

	// Initialize controllers
	authMiddleware := middleware.NewAuthMiddleware(authService)
	userController := controllers.NewUserController(userService, authService)
	groupController := controllers.NewGroupController(groupService)
	expenseController := controllers.NewExpenseController(expenseService)
	balanceController := controllers.NewBalanceController(balanceService)
	settlementController := controllers.NewSettlementController(settlementService)

	// Set up Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestSizeLimit(cfg.MaxRequestSize))
	router.Use(middleware.RateLimit(cfg.RateLimitPerSecond))

	// Public routes
	public := router.Group("/v1")
	{
		public.POST("/login", userController.Login)
		public.POST("/users", userController.CreateUser)
	}

	// Authenticated routes
	private := router.Group("/v1")
	private.Use(authMiddleware.Authenticate())
	{
		// User routes
		private.GET("/user-lookup", userController.LookupUser)
		private.GET("/users/:id", userController.GetUser)
		private.PUT("/users/:id", userController.UpdateUser)

		// Group routes
		private.POST("/groups", groupController.CreateGroup)
		private.GET("/groups/:id", groupController.GetGroup)
		private.GET("/groups/:id/members", groupController.GetMembers)
		private.POST("/groups/:id/members", groupController.AddMember)

		// Expense routes
		private.POST("/expenses", expenseController.CreateExpense)
		private.GET("/expenses/:id", expenseController.GetExpense)
		private.GET("/groups/:id/expenses", expenseController.ListGroupExpenses)
		private.GET("/users/:id/expenses", expenseController.ListUserExpenses)

		// Balance routes
		private.GET("/users/:id/balances", balanceController.GetUserBalances)
		private.GET("/groups/:id/balances", balanceController.GetGroupBalances)

		// Settlement routes
		private.POST("/settlements", settlementController.CreateSettlement)
		private.GET("/settlements/:id", settlementController.GetSettlement)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.ServerPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
