package main

import (
	"go-people-api/db"
	"go-people-api/handlers"
	log "go-people-api/log" // добавлено

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-people-api/docs"
)

func main() {
	log.Init()

	log.Logger.Info("Loading environment variables...")
	godotenv.Load()

	log.Logger.Info("Initializing database...")
	if err := db.Init(); err != nil {
		log.Logger.Fatal("DB connection failed: ", err)
	}

	r := gin.Default()

	log.Logger.Info("Setting up routes...")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/people", handlers.CreatePerson)
	r.GET("/people", handlers.GetPeople)
	r.DELETE("/people/:id", handlers.DeletePerson)
	r.PUT("/people/:id", handlers.UpdatePerson)

	log.Logger.Info("Server running on port 8086")
	r.Run(":8086")
}
