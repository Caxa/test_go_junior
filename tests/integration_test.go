package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Инициализация тестовой среды
	code := m.Run()
	os.Exit(code)
}

func TestAPIEndpoints(t *testing.T) {
	router := setupRouter() // Ваша функция инициализации роутера

	t.Run("Create Person", func(t *testing.T) {
		payload := []byte(`{"name": "Dmitriy", "surname": "Ushakov"}`)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/people", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["id"])
	})

	t.Run("Get Person", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/people/1", nil)

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update Person", func(t *testing.T) {
		payload := []byte(`{"age": 35}`)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/api/v1/people/1", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Delete Person", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/v1/people/1", nil)

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
