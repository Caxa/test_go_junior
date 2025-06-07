package unit

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-people-api/handlers"
	"go-people-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (m *MockService) Enrich(ctx context.Context, name string) (*models.Person, error) {
	return &models.Person{
		Age:         30,
		Gender:      "male",
		Nationality: "RU",
	}, nil
}

func TestCreatePersonHandler(t *testing.T) {
	// Настройка моков
	handlers.SetPersonService(&MockService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{"name": "Test", "surname": "User"}`)
	c.Request, _ = http.NewRequest("POST", "/api/v1/people", body)
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreatePerson(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}
