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

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Microsecond)
		defer cancel()
		time.Sleep(1 * time.Millisecond)

		_, err := service.Enrich(ctx, "Anna")
		require.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
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

	t.Run("invalid JSON", func(t *testing.T) {
		invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `invalid json`)
		}))
		defer invalidServer.Close()

		_, err := service.fetchAPI(ctx, invalidServer.URL)
		require.Error(t, err)
	})
}
