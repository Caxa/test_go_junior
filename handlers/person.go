package handlers

import (
	"go-people-api/db"
	"go-people-api/log"
	"go-people-api/models"
	"go-people-api/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/*
// Объявляем обогащающую функцию как переменную (может быть заменена в тестах)
var enrichFunc = func(name string) (*models.Person, error) {
	enriched, err := services.Enrich(name)
	if err != nil {
		return nil, err
	}
	return &models.Person{
		Gender:      enriched.Gender,
		Age:         enriched.Age,
		Nationality: enriched.Nationality,
	}, nil
}*/

// CreatePerson godoc
// @Summary Создать нового человека
// @Description Обогащает данные через внешние API и сохраняет в БД
// @Tags people
// @Accept json
// @Produce json
// @Param person body models.Person true "Информация о человеке"
// @Success 200 {object} models.Person
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people [post]
func CreatePerson(c *gin.Context) {
	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Logger.Warn("Invalid input for CreatePerson: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Logger.Infof("Creating person: %+v", input)

	enriched, err := services.Enrich(input.Name)
	if err != nil {
		log.Logger.Error("Enrichment failed: ", err)
	}

	query := `
		INSERT INTO people (name, surname, patronymic, gender, age, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err = db.DB.QueryRow(query, input.Name, input.Surname, input.Patronymic, enriched.Gender, enriched.Age, enriched.Nationality).
		Scan(&input.ID)
	if err != nil {
		log.Logger.Error("Insert error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert"})
		return
	}

	input.Gender = enriched.Gender
	input.Age = enriched.Age
	input.Nationality = enriched.Nationality
	log.Logger.Infof("Person created with ID %d", input.ID)
	c.JSON(http.StatusOK, input)
}

// GetPeople godoc
// @Summary Получить список людей
// @Description Возвращает список людей с возможностью фильтрации по имени
// @Tags people
// @Accept json
// @Produce json
// @Param name query string false "Имя для фильтрации"
// @Param limit query int false "Ограничение количества"
// @Param offset query int false "Смещение"
// @Success 200 {array} models.Person
// @Failure 500 {object} map[string]string
// @Router /people [get]
func GetPeople(c *gin.Context) {
	name := c.Query("name")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	log.Logger.Infof("Fetching people: name=%s, limit=%d, offset=%d", name, limit, offset)

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
		log.Logger.Error("Query error in GetPeople: ", err)
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

	log.Logger.Infof("Returned %d people", len(people))
	c.JSON(http.StatusOK, people)
}

// DeletePerson godoc
// @Summary Удалить человека
// @Description Удаляет человека по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [delete]
func DeletePerson(c *gin.Context) {
	id := c.Param("id")
	log.Logger.Infof("Deleting person with ID: %s", id)

	_, err := db.DB.Exec("DELETE FROM people WHERE id=$1", id)
	if err != nil {
		log.Logger.Error("Delete failed: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

// UpdatePerson godoc
// @Summary Обновить данные человека
// @Description Обновляет имя, фамилию и отчество по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.Person true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /people/{id} [put]
func UpdatePerson(c *gin.Context) {
	id := c.Param("id")
	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Logger.Warn("Invalid input for update: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Logger.Infof("Updating person %s with %+v", id, input)

	_, err := db.DB.Exec(`
		UPDATE people SET name=$1, surname=$2, patronymic=$3 WHERE id=$4
	`, input.Name, input.Surname, input.Patronymic, id)

	if err != nil {
		log.Logger.Error("Update failed: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": id})
}
