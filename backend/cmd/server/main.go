package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/task-manager/backend/internal/config"
	"github.com/task-manager/backend/internal/handler"
	"github.com/task-manager/backend/internal/middleware"
	"github.com/task-manager/backend/internal/repository"
	"github.com/task-manager/backend/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("mysql", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	taskService := service.NewTaskService(taskRepo)
	aiService := service.NewAIService(cfg.GeminiKey)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)
	aiHandler := handler.NewAIHandler(aiService, taskService)

	// Set up Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Auth routes (protected)
			protected.GET("/auth/me", authHandler.Me)

			// Task routes
			tasks := protected.Group("/tasks")
			{
				tasks.GET("", taskHandler.ListTasks)
				tasks.POST("", taskHandler.CreateTask)
				tasks.GET("/categories", taskHandler.GetCategories)
				tasks.GET("/:id", taskHandler.GetTask)
				tasks.PUT("/:id", taskHandler.UpdateTask)
				tasks.DELETE("/:id", taskHandler.DeleteTask)
			}

			// AI routes
			ai := protected.Group("/ai")
			{
				ai.POST("/generate", aiHandler.GenerateTasks)
				ai.POST("/breakdown/:id", aiHandler.BreakdownTask)
				ai.POST("/suggest-priority", aiHandler.SuggestPriority)
				ai.POST("/estimate-time", aiHandler.EstimateTime)
			}
		}
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
