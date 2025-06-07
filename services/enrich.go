package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	log "go-people-api/log"
	"go-people-api/models"
)

// Enricher определяет интерфейс для сервиса обогащения данных
type Enricher interface {
	Enrich(ctx context.Context, name string) (*models.Person, error)
}

type EnrichmentService struct {
	client *http.Client
}

// NewEnrichmentService создает новый экземпляр сервиса обогащения
func NewEnrichmentService() *EnrichmentService {
	return &EnrichmentService{
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

// fetchAPI выполняет запрос к внешнему API и парсит JSON ответ
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
		return nil, errors.New("api returned non-200 status code")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Enrich обогащает данные о человеке через внешние API
func (s *EnrichmentService) Enrich(ctx context.Context, name string) (*models.Person, error) {
	logger := log.WithContext(ctx)
	logger.Debugf("Starting enrichment for name: %s", name)

	person := &models.Person{}
	var errs []error

	// Получаем возраст
	if ageAPI := os.Getenv("AGE_API"); ageAPI != "" {
		if age, err := s.getAge(ctx, ageAPI, name); err == nil {
			person.Age = age
			logger.Debugf("Age fetched: %d", age)
		} else {
			errs = append(errs, err)
			logger.Warnf("Failed to fetch age: %v", err)
		}
	}

	// Получаем пол
	if genderAPI := os.Getenv("GENDER_API"); genderAPI != "" {
		if gender, err := s.getGender(ctx, genderAPI, name); err == nil {
			person.Gender = gender
			logger.Debugf("Gender fetched: %s", gender)
		} else {
			errs = append(errs, err)
			logger.Warnf("Failed to fetch gender: %v", err)
		}
	}

	// Получаем национальность
	if nationalityAPI := os.Getenv("NATIONALITY_API"); nationalityAPI != "" {
		if nationality, err := s.getNationality(ctx, nationalityAPI, name); err == nil {
			person.Nationality = nationality
			logger.Debugf("Nationality fetched: %s", nationality)
		} else {
			errs = append(errs, err)
			logger.Warnf("Failed to fetch nationality: %v", err)
		}
	}

	if len(errs) > 0 {
		return person, errors.Join(errs...)
	}

	logger.Infof("Enrichment complete for %s: %+v", name, person)
	return person, nil
}

func (s *EnrichmentService) getAge(ctx context.Context, apiURL, name string) (int, error) {
	res, err := s.fetchAPI(ctx, apiURL+"?name="+name)
	if err != nil {
		return 0, err
	}

	age, ok := res["age"].(float64)
	if !ok {
		return 0, errors.New("invalid age format in response")
	}

	return int(age), nil
}

func (s *EnrichmentService) getGender(ctx context.Context, apiURL, name string) (string, error) {
	res, err := s.fetchAPI(ctx, apiURL+"?name="+name)
	if err != nil {
		return "", err
	}

	gender, ok := res["gender"].(string)
	if !ok {
		return "", errors.New("invalid gender format in response")
	}

	return gender, nil
}

func (s *EnrichmentService) getNationality(ctx context.Context, apiURL, name string) (string, error) {
	res, err := s.fetchAPI(ctx, apiURL+"?name="+name)
	if err != nil {
		return "", err
	}

	countries, ok := res["country"].([]interface{})
	if !ok || len(countries) == 0 {
		return "", errors.New("no country data in response")
	}

	country, ok := countries[0].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid country format in response")
	}

	id, ok := country["country_id"].(string)
	if !ok {
		return "", errors.New("invalid country_id format in response")
	}

	return id, nil
}
