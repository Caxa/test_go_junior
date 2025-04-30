package services

import (
	"testing"
)

func TestEnrich(t *testing.T) {
	enriched, err := Enrich("Dmitriy")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if enriched.Age <= 0 {
		t.Errorf("Expected age > 0, got %d", enriched.Age)
	}
	if enriched.Gender == "" {
		t.Error("Gender should not be empty")
	}
	if enriched.Nationality == "" {
		t.Error("Nationality should not be empty")
	}
}
