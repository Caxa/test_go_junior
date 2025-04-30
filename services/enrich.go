package services

import (
	"encoding/json"
	"net/http"
)

type Enriched struct {
	Age         int
	Gender      string
	Nationality string
}

func Enrich(name string) (*Enriched, error) {
	e := &Enriched{}

	if resp, err := http.Get("https://api.agify.io/?name=" + name); err == nil {
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if age, ok := res["age"].(float64); ok {
			e.Age = int(age)
		}
	}

	if resp, err := http.Get("https://api.genderize.io/?name=" + name); err == nil {
		defer resp.Body.Close()
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if gender, ok := res["gender"].(string); ok {
			e.Gender = gender
		}
	}

	if resp, err := http.Get("https://api.nationalize.io/?name=" + name); err == nil {
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
	}

	return e, nil
}
