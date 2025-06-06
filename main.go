package main

import (
	"os"

	"go-people-api/db"
	"go-people-api/handlers"
	log "go-people-api/log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-people-api/docs"
)

func main() {

	log.Init()

	log.Logger.Info("Loading environment variables…")
	if err := godotenv.Load(); err != nil {
		log.Logger.Warn("No .env file found or failed to load")
	}

	log.Logger.Info("Initializing database…")
	if err := db.Init(); err != nil {
		log.Logger.Fatal("DB connection failed: ", err)
	}

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/people", handlers.CreatePerson)
	r.GET("/people", handlers.GetPeople)
	r.GET("/people/:id", handlers.GetPersonByID)
	r.PUT("/people/:id", handlers.UpdatePerson)
	r.PATCH("/people/:id", handlers.PatchPerson)
	r.DELETE("/people/:id", handlers.DeletePerson)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}
	log.Logger.Info("Server running on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Logger.Fatal("Server stopped: ", err)
	}
}
