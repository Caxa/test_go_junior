package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-people-api/db"
	"go-people-api/log"
	"go-people-api/models"
	"go-people-api/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Объявляем зависимости как интерфейсы для тестирования
type PersonService interface {
	Enrich(ctx context.Context, name string) (*models.Person, error)
}

type Database interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

var (
	personService PersonService = services.NewEnrichmentService()
	database      Database      = db.DB
)

// CreatePerson godoc
// @Summary Создать нового человека
// @Description Обогащает данные через внешние API и сохраняет в БД
// @Tags people
// @Accept json
// @Produce json
// @Param person body models.Person true "Информация о человеке"
// @Success 201 {object} models.Person
// @Failure 400 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people [post]
func CreatePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Logger.WithError(err).Warn("Invalid input for CreatePerson")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Обогащение данных
	enriched, err := personService.Enrich(ctx, input.Name)
	if err != nil {
		log.Logger.WithError(err).Error("Enrichment failed")
		c.JSON(http.StatusFailedDependency, models.ErrorResponse{
			Error:   "enrichment_error",
			Message: "Failed to enrich person data",
		})
		return
	}

	query := `
		INSERT INTO people (name, surname, patronymic, gender, age, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err = database.QueryRowContext(ctx, query,
		input.Name,
		input.Surname,
		input.Patronymic,
		enriched.Gender,
		enriched.Age,
		enriched.Nationality,
	).Scan(&input.ID, &createdAt, &updatedAt)

	if err != nil {
		log.Logger.WithError(err).Error("Database insert failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create person",
		})
		return
	}

	input.Gender = enriched.Gender
	input.Age = enriched.Age
	input.Nationality = enriched.Nationality
	input.CreatedAt = createdAt
	input.UpdatedAt = updatedAt

	log.Logger.WithField("person_id", input.ID).Info("Person created successfully")
	c.JSON(http.StatusCreated, input)
}

// GetPeople godoc
// @Summary Получить список людей
// @Description Возвращает список людей с пагинацией и фильтрацией
// @Tags people
// @Accept json
// @Produce json
// @Param name query string false "Фильтр по имени"
// @Param surname query string false "Фильтр по фамилии"
// @Param gender query string false "Фильтр по полу"
// @Param age_from query int false "Минимальный возраст"
// @Param age_to query int false "Максимальный возраст"
// @Param nationality query string false "Фильтр по национальности"
// @Param limit query int false "Лимит (по умолчанию 10)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} models.Person
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people [get]
func GetPeople(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Парсинг параметров запроса
	filter := models.PersonFilter{
		Name:        c.Query("name"),
		Surname:     c.Query("surname"),
		Gender:      c.Query("gender"),
		Nationality: c.Query("nationality"),
	}

	if ageFrom := c.Query("age_from"); ageFrom != "" {
		if val, err := strconv.Atoi(ageFrom); err == nil {
			filter.AgeFrom = &val
		}
	}

	if ageTo := c.Query("age_to"); ageTo != "" {
		if val, err := strconv.Atoi(ageTo); err == nil {
			filter.AgeTo = &val
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Построение SQL запроса
	query, args := buildFilterQuery(filter, limit, offset)

	rows, err := database.QueryContext(ctx, query, args...)
	if err != nil {
		log.Logger.WithError(err).Error("Database query failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch people",
		})
		return
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Surname,
			&p.Patronymic,
			&p.Gender,
			&p.Age,
			&p.Nationality,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			log.Logger.WithError(err).Warn("Failed to scan person row")
			continue
		}
		people = append(people, p)
	}

	if err = rows.Err(); err != nil {
		log.Logger.WithError(err).Error("Rows iteration error")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to process results",
		})
		return
	}

	c.JSON(http.StatusOK, people)
}

// buildFilterQuery строит SQL запрос на основе фильтров
func buildFilterQuery(filter models.PersonFilter, limit, offset int) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}
	var conditions []string

	query.WriteString(`
		SELECT id, name, surname, patronymic, gender, age, nationality, created_at, updated_at
		FROM people
	`)

	argPos := 1

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	if filter.Surname != "" {
		conditions = append(conditions, fmt.Sprintf("surname ILIKE $%d", argPos))
		args = append(args, "%"+filter.Surname+"%")
		argPos++
	}

	if filter.Gender != "" {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argPos))
		args = append(args, filter.Gender)
		argPos++
	}

	if filter.Nationality != "" {
		conditions = append(conditions, fmt.Sprintf("nationality = $%d", argPos))
		args = append(args, filter.Nationality)
		argPos++
	}

	if filter.AgeFrom != nil {
		conditions = append(conditions, fmt.Sprintf("age >= $%d", argPos))
		args = append(args, *filter.AgeFrom)
		argPos++
	}

	if filter.AgeTo != nil {
		conditions = append(conditions, fmt.Sprintf("age <= $%d", argPos))
		args = append(args, *filter.AgeTo)
		argPos++
	}

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, offset)

	return query.String(), args
}

// GetPersonByID godoc
// @Summary Получить человека по ID
// @Description Возвращает полную информацию о человеке
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Success 200 {object} models.Person
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people/{id} [get]
func GetPersonByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	var p models.Person
	err = database.QueryRowContext(ctx, `
		SELECT id, name, surname, patronymic, gender, age, nationality, created_at, updated_at
		FROM people WHERE id = $1
	`, id).Scan(
		&p.ID,
		&p.Name,
		&p.Surname,
		&p.Patronymic,
		&p.Gender,
		&p.Age,
		&p.Nationality,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Person not found",
			})
			return
		}

		log.Logger.WithError(err).Error("Database query failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch person",
		})
		return
	}

	c.JSON(http.StatusOK, p)
}

// UpdatePerson godoc
// @Summary Полное обновление данных человека
// @Description Обновляет все поля человека по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.Person true "Данные для обновления"
// @Success 200 {object} models.Person
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people/{id} [put]
func UpdatePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Logger.WithError(err).Warn("Invalid input for UpdatePerson")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Проверка существования записи
	var exists bool
	err = database.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM people WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Person not found",
		})
		return
	}

	query := `
		UPDATE people 
		SET name = $1, surname = $2, patronymic = $3, gender = $4, age = $5, nationality = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	var updatedAt time.Time
	err = database.QueryRowContext(ctx, query,
		input.Name,
		input.Surname,
		input.Patronymic,
		input.Gender,
		input.Age,
		input.Nationality,
		id,
	).Scan(&updatedAt)

	if err != nil {
		log.Logger.WithError(err).Error("Database update failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update person",
		})
		return
	}

	input.ID = id
	input.UpdatedAt = updatedAt

	c.JSON(http.StatusOK, input)
}

// PatchPerson godoc
// @Summary Частичное обновление данных человека
// @Description Обновляет указанные поля человека по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Param person body models.UpdatePersonRequest true "Поля для обновления"
// @Success 200 {object} models.Person
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people/{id} [patch]
func PatchPerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	var input models.UpdatePersonRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Logger.WithError(err).Warn("Invalid input for PatchPerson")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	// Проверка существования записи
	var exists bool
	err = database.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM people WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Person not found",
		})
		return
	}

	// Построение динамического запроса
	query, args := buildPartialUpdateQuery(id, input)

	var p models.Person
	err = database.QueryRowContext(ctx, query, args...).Scan(
		&p.ID,
		&p.Name,
		&p.Surname,
		&p.Patronymic,
		&p.Gender,
		&p.Age,
		&p.Nationality,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		log.Logger.WithError(err).Error("Database partial update failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to partially update person",
		})
		return
	}

	c.JSON(http.StatusOK, p)
}

// buildPartialUpdateQuery строит запрос для частичного обновления
func buildPartialUpdateQuery(id int, input models.UpdatePersonRequest) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}
	var updates []string

	argPos := 1

	if input.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *input.Name)
		argPos++
	}

	if input.Surname != nil {
		updates = append(updates, fmt.Sprintf("surname = $%d", argPos))
		args = append(args, *input.Surname)
		argPos++
	}

	if input.Patronymic != nil {
		updates = append(updates, fmt.Sprintf("patronymic = $%d", argPos))
		args = append(args, *input.Patronymic)
		argPos++
	}

	if input.Gender != nil {
		updates = append(updates, fmt.Sprintf("gender = $%d", argPos))
		args = append(args, *input.Gender)
		argPos++
	}

	if input.Age != nil {
		updates = append(updates, fmt.Sprintf("age = $%d", argPos))
		args = append(args, *input.Age)
		argPos++
	}

	if input.Nationality != nil {
		updates = append(updates, fmt.Sprintf("nationality = $%d", argPos))
		args = append(args, *input.Nationality)
		argPos++
	}

	if len(updates) == 0 {
		updates = append(updates, "updated_at = updated_at") // Ничего не обновляем, но триггер сработает
	} else {
		updates = append(updates, "updated_at = NOW()")
	}

	query.WriteString(`
		UPDATE people 
		SET ` + strings.Join(updates, ", ") + `
		WHERE id = $` + strconv.Itoa(argPos) + `
		RETURNING id, name, surname, patronymic, gender, age, nationality, created_at, updated_at
	`)
	args = append(args, id)

	return query.String(), args
}

// DeletePerson godoc
// @Summary Удалить человека
// @Description Удаляет человека по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID человека"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /people/{id} [delete]
func DeletePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	result, err := database.ExecContext(ctx, "DELETE FROM people WHERE id = $1", id)
	if err != nil {
		log.Logger.WithError(err).Error("Database delete failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete person",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Person not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
