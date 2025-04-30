package handlers

import (
	"go-people-api/db"
	"go-people-api/models"
	"go-people-api/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreatePerson(c *gin.Context) {
	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enriched, _ := services.Enrich(input.Name)

	query := `
		INSERT INTO people (name, surname, patronymic, gender, age, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := db.DB.QueryRow(query, input.Name, input.Surname, input.Patronymic, enriched.Gender, enriched.Age, enriched.Nationality).
		Scan(&input.ID)
	if err != nil {
		log.Println("Insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert"})
		return
	}

	input.Gender = enriched.Gender
	input.Age = enriched.Age
	input.Nationality = enriched.Nationality
	c.JSON(http.StatusOK, input)
}

func GetPeople(c *gin.Context) {
	name := c.Query("name")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	query := "SELECT id, name, surname, patronymic, gender, age, nationality FROM people"
	var args []interface{}
	if name != "" {
		query += " WHERE name ILIKE $1"
		args = append(args, "%"+name+"%")
	}
	query += " ORDER BY id LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Patronymic, &p.Gender, &p.Age, &p.Nationality)
		people = append(people, p)
	}

	c.JSON(http.StatusOK, people)
}

func DeletePerson(c *gin.Context) {
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM people WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func UpdatePerson(c *gin.Context) {
	id := c.Param("id")
	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.DB.Exec(`
		UPDATE people SET name=$1, surname=$2, patronymic=$3 WHERE id=$4
	`, input.Name, input.Surname, input.Patronymic, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": id})
}
