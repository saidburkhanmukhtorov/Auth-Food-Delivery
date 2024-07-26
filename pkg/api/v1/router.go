package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
	_ "github.com/time_capsule/Auth-Servic-Timecapsule/docs"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/auth"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/redis"
	"github.com/time_capsule/Auth-Servic-Timecapsule/pkg/api/middleware"
	"github.com/time_capsule/Auth-Servic-Timecapsule/pkg/api/v1/handlers"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /v1

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description					Description for what is this security definition being used
func SetupRouter(db *pgxpool.Pool, redisClient *redis.Client, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Swagger setup
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Apply global middleware
	router.Use(middleware.Logger())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, redisClient, cfg)
	userHandler := handlers.NewUserHandler(db)

	// API version 1 group
	v1 := router.Group("")
	{
		// Authentication routes
		authR := v1.Group("/auth")
		{
			authR.POST("/register", authHandler.Register)
			authR.POST("/verify-otp", authHandler.VerifyOTP) // Route for OTP verification
			authR.POST("/login", authHandler.Login)
			authR.GET("/validate", auth.AuthMiddleware(cfg), authHandler.Validate)
			authR.POST("/forgot-password", authHandler.ForgotPassword)
			authR.POST("/reset-password", authHandler.ResetPassword)
			authR.POST("/approve-user", authHandler.ApproveUser)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.Use(auth.AuthMiddleware(cfg)) // Protect user routes with auth middleware
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:userId", auth.AuthorizationMiddleware(), userHandler.GetUserByID)
			users.PUT("/:userId", auth.AuthorizationMiddleware(), userHandler.UpdateUser)
			users.DELETE("/:userId", userHandler.DeleteUser)
		}
	}

	return router
}
