package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "go-people-api/log"

	"github.com/sirupsen/logrus"
)

// Переопределяем httpGet для мокинга
func init() {
	// Инициализация логгера для теста
	log.Logger = logrus.New()
	log.Logger.SetLevel(logrus.DebugLevel)
}

func TestEnrich(t *testing.T) {
	// Моки ответов для трёх API
	agifyResp := `{"age": 30}`
	genderizeResp := `{"gender": "male"}`
	nationalizeResp := `{"country": [{"country_id": "US"}]}`

	// Создаём тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.String(), "agify"):
			fmt.Fprint(w, agifyResp)
		case strings.Contains(r.URL.String(), "genderize"):
			fmt.Fprint(w, genderizeResp)
		case strings.Contains(r.URL.String(), "nationalize"):
			fmt.Fprint(w, nationalizeResp)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Подмена httpGet функцией, которая возвращает ответы с тестового сервера
	httpGet = func(url string) (*http.Response, error) {
		switch {
		case strings.Contains(url, "agify.io"):
			return http.Get(server.URL + "/agify")
		case strings.Contains(url, "genderize.io"):
			return http.Get(server.URL + "/genderize")
		case strings.Contains(url, "nationalize.io"):
			return http.Get(server.URL + "/nationalize")
		default:
			return nil, fmt.Errorf("unknown url: %s", url)
		}
	}

	// Запуск теста
	result, err := Enrich("John")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Age != 30 {
		t.Errorf("Expected Age=30, got %d", result.Age)
	}
	if result.Gender != "male" {
		t.Errorf("Expected Gender=male, got %s", result.Gender)
	}
	if result.Nationality != "US" {
		t.Errorf("Expected Nationality=US, got %s", result.Nationality)
	}
}
