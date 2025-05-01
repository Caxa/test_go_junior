package services

import (
	"encoding/json"
	"net/http"

	log "go-people-api/log"
	"go-people-api/models"
)

var httpGet = http.Get

func Enrich(name string) (*models.Person, error) {
	log.Logger.Debug("Starting enrichment for name: ", name)
	e := &models.Person{}

	// Age
	if resp, err := httpGet("https://api.agify.io/?name=" + name); err == nil {
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if age, ok := res["age"].(float64); ok {
			e.Age = int(age)
		}
		log.Logger.Debug("Age fetched: ", e.Age)
	} else {
		log.Logger.Warn("Failed to fetch age: ", err)
	}

	// Gender
	if resp, err := httpGet("https://api.genderize.io/?name=" + name); err == nil {
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if gender, ok := res["gender"].(string); ok {
			e.Gender = gender
		}
		log.Logger.Debug("Gender fetched: ", e.Gender)
	} else {
		log.Logger.Warn("Failed to fetch gender: ", err)
	}

	// Nationality
	if resp, err := httpGet("https://api.nationalize.io/?name=" + name); err == nil {
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if countries, ok := res["country"].([]interface{}); ok && len(countries) > 0 {
			if c, ok := countries[0].(map[string]interface{}); ok {
				if id, ok := c["country_id"].(string); ok {
					e.Nationality = id
				}
			}
		}
		log.Logger.Debug("Nationality fetched: ", e.Nationality)
	} else {
		log.Logger.Warn("Failed to fetch nationality: ", err)
	}

	log.Logger.Info("Enrichment complete for ", name, ": ", *e)
	return e, nil
}
