package services

import (
	"context"
	"encoding/json"
	log "go-people-api/log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAPI(t *testing.T, response interface{}) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("could not encode mock response: %v", err)
		}
	})
	return httptest.NewServer(handler)
}

func TestEnrichmentService_Enrich_PartialFailure(t *testing.T) {
	ageServer := mockAPI(t, map[string]interface{}{"age": 25})
	defer ageServer.Close()

	genderServer := mockAPI(t, map[string]interface{}{"something": "wrong"})
	defer genderServer.Close()

	nationalityServer := mockAPI(t, map[string]interface{}{
		"country": []interface{}{
			map[string]interface{}{"country_id": "CA"},
		},
	})
	defer nationalityServer.Close()

	service := NewEnrichmentService(ageServer.URL, genderServer.URL, nationalityServer.URL)
	person, err := service.Enrich(context.Background(), "Alice")

	if err == nil {
		t.Errorf("expected partial enrichment error, got nil")
	}
	if person.Age != 25 {
		t.Errorf("expected age 25, got %d", person.Age)
	}
	if person.Gender != "" {
		t.Errorf("expected gender to be empty, got %s", person.Gender)
	}
	if person.Nationality != "CA" {
		t.Errorf("expected nationality 'CA', got %s", person.Nationality)
	}
}

func TestEnrichmentService_Enrich_NoAPIConfigured(t *testing.T) {
	service := NewEnrichmentService("", "", "")
	person, err := service.Enrich(context.Background(), "Test")

	if err == nil {
		t.Errorf("expected error due to missing API configs, got nil")
	}
	if person == nil {
		t.Errorf("expected partial result, got nil")
	}
}
func TestEnrichmentService_Enrich_Success(t *testing.T) {
	log.Init()

	ageServer := mockAPI(t, map[string]interface{}{"age": 30})
	defer ageServer.Close()

	genderServer := mockAPI(t, map[string]interface{}{"gender": "male"})
	defer genderServer.Close()

	nationalityServer := mockAPI(t, map[string]interface{}{
		"country": []interface{}{
			map[string]interface{}{"country_id": "US"},
		},
	})
	defer nationalityServer.Close()

	service := NewEnrichmentService(ageServer.URL, genderServer.URL, nationalityServer.URL)

	ctx := context.Background()
	enriched, err := service.Enrich(ctx, "John")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if enriched.Age != 30 {
		t.Errorf("expected age 30, got %d", enriched.Age)
	}
	if enriched.Gender != "male" {
		t.Errorf("expected gender 'male', got %s", enriched.Gender)
	}
	if enriched.Nationality != "US" {
		t.Errorf("expected nationality 'US', got %s", enriched.Nationality)
	}
}
