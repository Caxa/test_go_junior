package main

import (
	"fmt"
	"os"
	"time"

	"go-people-api/db"
	"go-people-api/handlers"
	"go-people-api/log"
	"go-people-api/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-people-api/docs"
)

func main() {
	log.Init()

	if err := loadEnvWithTimeout(5 * time.Second); err != nil {
		log.Logger.Fatal("Failed to load environment variables: ", err)
	}

	if err := initDBWithRetry(5, 3*time.Second); err != nil {
		log.Logger.Fatal("Failed to initialize database: ", err)
	}
	log.Logger.Info("Successfully connected to database")

	checkExternalAPIs()

	enrichmentService := services.NewEnrichmentService(
		os.Getenv("AGE_API"),
		os.Getenv("GENDER_API"),
		os.Getenv("NATIONALITY_API"),
	)
	handlers.SetPersonService(enrichmentService)

	r := setupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Logger.Info("Server starting on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Logger.Fatal("Server failed: ", err)
	}
}

func loadEnvWithTimeout(timeout time.Duration) error {
	errChan := make(chan error)
	go func() {
		if err := godotenv.Load(); err != nil {
			errChan <- fmt.Errorf("failed to load .env file: %w", err)
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timeout while loading .env file")
	}
}

func initDBWithRetry(maxRetries int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := db.Init(); err == nil {
			return nil
		}

	}
	return fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastErr)
}

func checkExternalAPIs() {
	requiredAPIs := map[string]string{
		"AGE_API":         "Age API",
		"GENDER_API":      "Gender API",
		"NATIONALITY_API": "Nationality API",
	}

	for env, name := range requiredAPIs {
		if os.Getenv(env) == "" {
			log.Logger.Warnf("%s endpoint not configured (%s environment variable is empty)", name, env)
		} else {
			log.Logger.Infof("%s endpoint: %s", name, os.Getenv(env))
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	if os.Getenv("GIN_MODE") != "release" {
		r.Use(gin.Logger())
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.POST("/people", handlers.CreatePerson)
		api.GET("/people", handlers.GetPeople)
		api.GET("/people/:id", handlers.GetPersonByID)
		api.PUT("/people/:id", handlers.UpdatePerson)
		api.PATCH("/people/:id", handlers.PatchPerson)
		api.DELETE("/people/:id", handlers.DeletePerson)
	}

	return r
}
