package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrichmentServices(t *testing.T) {
	// Мок-сервер для API обогащения
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

	t.Run("Age Enrichment", func(t *testing.T) {
		age, err := getAge(context.Background(), "Dmitriy", server.URL+"/agify")
		assert.NoError(t, err)
		assert.Equal(t, 35, age)
	})

	t.Run("Gender Enrichment", func(t *testing.T) {
		gender, err := getGender(context.Background(), "Dmitriy", server.URL+"/genderize")
		assert.NoError(t, err)
		assert.Equal(t, "male", gender)
	})

	t.Run("Nationality Enrichment", func(t *testing.T) {
		nationality, err := getNationality(context.Background(), "Dmitriy", server.URL+"/nationalize")
		assert.NoError(t, err)
		assert.Equal(t, "RU", nationality)
	})
}
