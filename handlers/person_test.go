package handlers

import (
	"bytes"
	"encoding/json"
	"go-people-api/db"
	"go-people-api/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Init()

	router := gin.Default()
	router.POST("/people", CreatePerson)

	person := models.Person{
		Name:       "Ivan",
		Surname:    "Petrov",
		Patronymic: "Sergeevich",
	}
	body, _ := json.Marshal(person)

	req, _ := http.NewRequest("POST", "/people", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.Code)
	}

	var created models.Person
	json.Unmarshal(resp.Body.Bytes(), &created)

	if created.ID == 0 || created.Age == 0 || created.Gender == "" {
		t.Error("Returned object was not enriched properly")
	}
}
