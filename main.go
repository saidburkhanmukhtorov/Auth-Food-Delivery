package main

import (
	"log"

	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
	_ "github.com/time_capsule/Auth-Servic-Timecapsule/docs"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/db"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/redis"
	v1 "github.com/time_capsule/Auth-Servic-Timecapsule/pkg/api/v1"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @BasePath  /v1
// @description					Description for what is this security definition being used
func main() {
	cfg := config.Load()

	// Initialize database connection
	dbPool, err := db.Connect(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	// Initialize Redis client
	redisClient, err := redis.Connect(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// Set up API routes
	router := v1.SetupRouter(dbPool, redisClient, &cfg)

	if err := router.Run(cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}

	// Graceful shutdown

	log.Println("Server exiting")
}
