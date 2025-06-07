package unit

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-people-api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockDB struct{}

func (m *MockDB) GetDB() (*sql.DB, error) {
	// Реализация мока БД
}

func TestHandlers(t *testing.T) {
	// Настройка моков
	handlers.SetPersonService(&MockEnrichmentService{})
	handlers.SetDB(&MockDB{})

	t.Run("Create Person Handler", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Подготовка запроса
		body := bytes.NewBufferString(`{"name": "Test", "surname": "User"}`)
		c.Request, _ = http.NewRequest("POST", "/api/v1/people", body)
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.CreatePerson(c)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
