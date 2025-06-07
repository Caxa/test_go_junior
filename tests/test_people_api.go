package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPeopleAPI(t *testing.T) {
	// 1. Тест создания пользователя
	t.Run("Create Person", func(t *testing.T) {
		payload := []byte(`{"name": "Алексей", "surname": "Петров", "patronymic": "Сергеевич"}`)
		req, _ := http.NewRequest("POST", "http://localhost:8086/api/v1/people", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		// Здесь должен быть вызов вашего обработчика
		// Например: router.ServeHTTP(resp, req)

		if status := resp.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		// Сохраняем ID для последующих тестов
		if id, ok := result["id"].(float64); ok {
			personID = int(id)
		}
	})

	// 2. Тест получения списка
	t.Run("Get People List", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://localhost:8086/api/v1/people", nil)
		resp := httptest.NewRecorder()

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	// 3. Тест получения по ID (используем сохраненный ID)
	t.Run("Get Person By ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://localhost:8086/api/v1/people/"+strconv.Itoa(personID), nil)
		resp := httptest.NewRecorder()

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	// 4. Тест полного обновления
	t.Run("Update Person", func(t *testing.T) {
		payload := []byte(`{"name": "Обновленное", "surname": "Имя", "age": 40, "gender": "male", "nationality": "RU"}`)
		req, _ := http.NewRequest("PUT", "http://localhost:8086/api/v1/people/"+strconv.Itoa(personID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	// 5. Тест частичного обновления
	t.Run("Patch Person", func(t *testing.T) {
		payload := []byte(`{"age": 35}`)
		req, _ := http.NewRequest("PATCH", "http://localhost:8086/api/v1/people/"+strconv.Itoa(personID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	// 6. Тест удаления
	t.Run("Delete Person", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "http://localhost:8086/api/v1/people/"+strconv.Itoa(personID), nil)
		resp := httptest.NewRecorder()

		if status := resp.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})
}

var personID int // Глобальная переменная для хранения ID созданного пользователя
