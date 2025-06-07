package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-people-api/services"

	"github.com/stretchr/testify/assert"
)

func TestEnrichmentService(t *testing.T) {
	// Мок-сервер для API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agify":
			w.Write([]byte(`{"age": 35}`))
		case "/genderize":
			w.Write([]byte(`{"gender": "male"}`))
		case "/nationalize":
			w.Write([]byte(`{"country": [{"country_id": "RU"}]}`))
		}
	}))
	defer server.Close()

	service := services.NewEnrichmentService(
		server.URL+"/agify",
		server.URL+"/genderize",
		server.URL+"/nationalize",
	)

	t.Run("Successful Enrichment", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		person, err := service.Enrich(ctx, "Dmitriy")
		assert.NoError(t, err)
		assert.Equal(t, 35, person.Age)
		assert.Equal(t, "male", person.Gender)
		assert.Equal(t, "RU", person.Nationality)
	})
}
