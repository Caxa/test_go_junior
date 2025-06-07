package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-people-api/handlers"
	"go-people-api/models"
	"go-people-api/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIEndpoints(t *testing.T) {
	// Инициализация тестового роутера
	router := gin.Default()
	mockService := &services.EnrichmentService{}
	handlers.SetPersonService(mockService)

	router.POST("/api/v1/people", handlers.CreatePerson)
	router.GET("/api/v1/people/:id", handlers.GetPersonByID)

	t.Run("Create Person Success", func(t *testing.T) {
		payload := []byte(`{"name": "Test", "surname": "User"}`)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/people", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var person models.Person
		json.Unmarshal(w.Body.Bytes(), &person)
		assert.NotZero(t, person.ID)
	})
}
