package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-people-api/handlers"
	"go-people-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIEndpoints(t *testing.T) {
	// Инициализация тестового роутера
	router := gin.Default()
	handlers.SetPersonService(&MockEnrichmentService{})

	router.POST("/api/v1/people", handlers.CreatePerson)
	router.GET("/api/v1/people", handlers.GetPeople)
	router.GET("/api/v1/people/:id", handlers.GetPersonByID)
	router.PUT("/api/v1/people/:id", handlers.UpdatePerson)
	router.PATCH("/api/v1/people/:id", handlers.PatchPerson)
	router.DELETE("/api/v1/people/:id", handlers.DeletePerson)

	t.Run("Create and Get Person", func(t *testing.T) {
		// Создание
		createPayload := []byte(`{"name": "Dmitriy", "surname": "Ushakov"}`)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/people", bytes.NewBuffer(createPayload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var createdPerson models.Person
		json.Unmarshal(w.Body.Bytes(), &createdPerson)

		// Получение
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/v1/people/"+string(createdPerson.ID), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var fetchedPerson models.Person
		json.Unmarshal(w.Body.Bytes(), &fetchedPerson)
		assert.Equal(t, createdPerson.ID, fetchedPerson.ID)
	})
}
