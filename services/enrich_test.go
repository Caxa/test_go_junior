package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichmentService(t *testing.T) {
	// Настройка тестового сервера
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/agify"):
			fmt.Fprint(w, `{"age": 35}`)
		case strings.Contains(r.URL.Path, "/genderize"):
			fmt.Fprint(w, `{"gender": "female"}`)
		case strings.Contains(r.URL.Path, "/nationalize"):
			fmt.Fprint(w, `{"country": [{"country_id": "RU"}]}`)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Установка переменных окружения для теста
	t.Setenv("AGE_API", server.URL+"/agify")
	t.Setenv("GENDER_API", server.URL+"/genderize")
	t.Setenv("NATIONALITY_API", server.URL+"/nationalize")

	ctx := context.Background()
	service := NewEnrichmentService()

	t.Run("successful enrichment", func(t *testing.T) {
		person, err := service.Enrich(ctx, "Anna")
		require.NoError(t, err)
		assert.Equal(t, 35, person.Age)
		assert.Equal(t, "female", person.Gender)
		assert.Equal(t, "RU", person.Nationality)
	})

	t.Run("partial enrichment with one failed API", func(t *testing.T) {
		t.Setenv("GENDER_API", server.URL+"/invalid")

		person, err := service.Enrich(ctx, "Anna")
		require.Error(t, err)
		assert.Equal(t, 35, person.Age)
		assert.Equal(t, "", person.Gender) // Gender не получен
		assert.Equal(t, "RU", person.Nationality)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond) // Гарантируем истечение таймаута

		_, err := service.Enrich(ctx, "Anna")
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("invalid API responses", func(t *testing.T) {
		invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"invalid": "data"}`)
		}))
		defer invalidServer.Close()

		t.Setenv("AGE_API", invalidServer.URL)
		t.Setenv("GENDER_API", invalidServer.URL)
		t.Setenv("NATIONALITY_API", invalidServer.URL)

		_, err := service.Enrich(ctx, "Anna")
		require.Error(t, err)
	})
}

func TestFetchAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"key": "value"}`)
	}))
	defer server.Close()

	service := NewEnrichmentService()
	ctx := context.Background()

	t.Run("successful request", func(t *testing.T) {
		res, err := service.fetchAPI(ctx, server.URL)
		require.NoError(t, err)
		assert.Equal(t, "value", res["key"])
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := service.fetchAPI(ctx, "http://invalid.url")
		require.Error(t, err)
	})

	t.Run("non-200 status", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		_, err := service.fetchAPI(ctx, errorServer.URL)
		require.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `invalid json`)
		}))
		defer invalidJSONServer.Close()

		_, err := service.fetchAPI(ctx, invalidJSONServer.URL)
		require.Error(t, err)
	})
}
