package services

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "go-people-api/log"
	"go-people-api/models"
)

func fetchJSON(url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func Enrich(name string) (*models.Person, error) {
	log.Logger.Debug("Starting enrichment for name: ", name)
	person := &models.Person{}

	// Age
	if res, err := fetchJSON(os.Getenv("AGE_API") + "?name=" + name); err == nil {
		if age, ok := res["age"].(float64); ok {
			person.Age = int(age)
		}
		log.Logger.Debug("Age fetched: ", person.Age)
	} else {
		log.Logger.Warn("Failed to fetch age: ", err)
	}

	// Gender
	if res, err := fetchJSON(os.Getenv("GENDER_API") + "?name=" + name); err == nil {
		if gender, ok := res["gender"].(string); ok {
			person.Gender = gender
		}
		log.Logger.Debug("Gender fetched: ", person.Gender)
	} else {
		log.Logger.Warn("Failed to fetch gender: ", err)
	}

	// Nationality
	if res, err := fetchJSON(os.Getenv("NATIONALITY_API") + "?name=" + name); err == nil {
		if countries, ok := res["country"].([]interface{}); ok && len(countries) > 0 {
			if c, ok := countries[0].(map[string]interface{}); ok {
				if id, ok := c["country_id"].(string); ok {
					person.Nationality = id
				}
			}
		}
		log.Logger.Debug("Nationality fetched: ", person.Nationality)
	} else {
		log.Logger.Warn("Failed to fetch nationality: ", err)
	}

	log.Logger.Info("Enrichment complete for ", name, ": ", *person)
	return person, nil
}
