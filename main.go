package main

import (
	"go-people-api/db"
	"go-people-api/handlers"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	if err := db.Init(); err != nil {
		log.Fatal("DB connection failed:", err)
	}

	r := gin.Default()

	r.POST("/people", handlers.CreatePerson)
	r.GET("/people", handlers.GetPeople)
	r.DELETE("/people/:id", handlers.DeletePerson)
	r.PUT("/people/:id", handlers.UpdatePerson)

	r.Run(":8086")
}
