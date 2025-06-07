package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "go-people-api/log"
	"go-people-api/models"
)

type EnrichmentService struct {
	client         *http.Client
	ageAPI         string
	genderAPI      string
	nationalityAPI string
}

func NewEnrichmentService(ageAPI, genderAPI, nationalityAPI string) *EnrichmentService {
	return &EnrichmentService{
		client: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 5,
			},
		},
		ageAPI:         ageAPI,
		genderAPI:      genderAPI,
		nationalityAPI: nationalityAPI,
	}
}

func (s *EnrichmentService) Enrich(ctx context.Context, name string) (*models.Person, error) {
	logger := log.WithContext(ctx)
	logger.Infof("Starting enrichment for: %s", name)

	person := &models.Person{}
	errChan := make(chan error, 3)
	resultChan := make(chan struct {
		field string
		value interface{}
	}, 3)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Параллельное выполнение запросов
	go s.fetchAge(ctx, name, resultChan, errChan)
	go s.fetchGender(ctx, name, resultChan, errChan)
	go s.fetchNationality(ctx, name, resultChan, errChan)

	// Обработка результатов
	var errs []error
	for i := 0; i < 3; i++ {
		select {
		case res := <-resultChan:
			switch res.field {
			case "age":
				if age, ok := res.value.(int); ok {
					person.Age = age
				}
			case "gender":
				if gender, ok := res.value.(string); ok {
					person.Gender = gender
				}
			case "nationality":
				if nationality, ok := res.value.(string); ok {
					person.Nationality = nationality
				}
			}
		case err := <-errChan:
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		logger.Warnf("Partial enrichment errors: %v", errs)
		return person, fmt.Errorf("partial enrichment failure (%d errors)", len(errs))
	}

	logger.Infof("Successfully enriched: %+v", person)
	return person, nil
}

func (s *EnrichmentService) fetchAge(ctx context.Context, name string, resultChan chan<- struct {
	field string
	value interface{}
}, errChan chan<- error) {
	if s.ageAPI == "" {
		errChan <- errors.New("age API not configured")
		return
	}

	res, err := s.fetchAPI(ctx, s.ageAPI+"?name="+name)
	if err != nil {
		errChan <- fmt.Errorf("age API request failed: %w", err)
		return
	}

	age, ok := res["age"]
	if !ok {
		errChan <- errors.New("age field missing in response")
		return
	}

	ageFloat, ok := age.(float64)
	if !ok {
		errChan <- fmt.Errorf("invalid age type: %T", age)
		return
	}

	resultChan <- struct {
		field string
		value interface{}
	}{"age", int(ageFloat)}
}

func (s *EnrichmentService) fetchGender(ctx context.Context, name string, resultChan chan<- struct {
	field string
	value interface{}
}, errChan chan<- error) {
	if s.genderAPI == "" {
		errChan <- errors.New("gender API not configured")
		return
	}

	res, err := s.fetchAPI(ctx, s.genderAPI+"?name="+name)
	if err != nil {
		errChan <- fmt.Errorf("gender API request failed: %w", err)
		return
	}

	gender, ok := res["gender"]
	if !ok {
		errChan <- errors.New("gender field missing in response")
		return
	}

	genderStr, ok := gender.(string)
	if !ok {
		errChan <- fmt.Errorf("invalid gender type: %T", gender)
		return
	}

	resultChan <- struct {
		field string
		value interface{}
	}{"gender", genderStr}
}

func (s *EnrichmentService) fetchNationality(ctx context.Context, name string, resultChan chan<- struct {
	field string
	value interface{}
}, errChan chan<- error) {
	if s.nationalityAPI == "" {
		errChan <- errors.New("nationality API not configured")
		return
	}

	res, err := s.fetchAPI(ctx, s.nationalityAPI+"?name="+name)
	if err != nil {
		errChan <- fmt.Errorf("nationality API request failed: %w", err)
		return
	}

	countries, ok := res["country"].([]interface{})
	if !ok || len(countries) == 0 {
		errChan <- errors.New("no country data in response")
		return
	}

	country, ok := countries[0].(map[string]interface{})
	if !ok {
		errChan <- errors.New("invalid country format in response")
		return
	}

	id, ok := country["country_id"].(string)
	if !ok {
		errChan <- errors.New("invalid country_id format in response")
		return
	}

	resultChan <- struct {
		field string
		value interface{}
	}{"nationality", id}
}

func (s *EnrichmentService) fetchAPI(ctx context.Context, url string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
