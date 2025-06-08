package handlers

import (
	"context"
	"database/sql"
	"errors"
	"go-people-api/db"
	"go-people-api/log"
	"go-people-api/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type PersonService interface {
	Enrich(ctx context.Context, name string) (*models.Person, error)
}

var (
	personService PersonService
)

func SetPersonService(service PersonService) {
	personService = service
}

func CreatePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid input")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	if input.Name == "" || input.Surname == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Name and surname are required",
		})
		return
	}

	enriched, err := personService.Enrich(ctx, input.Name)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("Partial enrichment failure")

	}

	result := mergePersonData(&input, enriched)

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	query := `
		INSERT INTO people 
		(name, surname, patronymic, gender, age, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var gender *string
	if result.Gender != "" {
		gender = &result.Gender
	}

	var nationality *string
	if result.Nationality != "" {
		nationality = &result.Nationality
	}

	err = dbConn.QueryRowContext(ctx, query,
		result.Name,
		result.Surname,
		result.Patronymic,
		gender,
		result.Age,
		nationality,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		handleDatabaseError(c, ctx, err, "Failed to create person record")
		return
	}

	c.JSON(http.StatusCreated, result)
}

func mergePersonData(input, enriched *models.Person) *models.Person {
	if enriched == nil {
		return input
	}

	result := *input
	if enriched.Age > 0 && input.Age == 0 {
		result.Age = enriched.Age
	}
	if enriched.Gender != "" && input.Gender == "" {
		result.Gender = enriched.Gender
	}
	if enriched.Nationality != "" && input.Nationality == "" {
		result.Nationality = enriched.Nationality
	}
	return &result
}

func GetPeople(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var filter models.PersonFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid filter params")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid filter parameters",
			Details: err.Error(),
		})
		return
	}

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	query, args := buildFilterQuery(filter)
	rows, err := dbConn.QueryContext(ctx, query, args...)
	if err != nil {
		handleDatabaseError(c, ctx, err, "Failed to fetch people")
		return
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		var gender, nationality sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Surname, &p.Patronymic,
			&p.Age, &gender, &nationality, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			log.WithContext(ctx).WithError(err).Error("DB scan failed")
			continue
		}
		if gender.Valid {
			p.Gender = gender.String
		}
		if nationality.Valid {
			p.Nationality = nationality.String
		}
		people = append(people, p)
	}

	c.JSON(http.StatusOK, people)
}

func GetPersonByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	var person models.Person
	var gender, nationality sql.NullString
	query := `SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at 
	          FROM people WHERE id = $1`
	err = dbConn.QueryRowContext(ctx, query, id).Scan(
		&person.ID, &person.Name, &person.Surname, &person.Patronymic,
		&person.Age, &gender, &nationality, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Person not found",
			})
			return
		}
		handleDatabaseError(c, ctx, err, "Failed to fetch person")
		return
	}

	if gender.Valid {
		person.Gender = gender.String
	}
	if nationality.Valid {
		person.Nationality = nationality.String
	}

	c.JSON(http.StatusOK, person)
}

func UpdatePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid input")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	query := `
		UPDATE people 
		SET name = $1, surname = $2, patronymic = $3, age = $4, 
		    gender = $5, nationality = $6
		WHERE id = $7
		RETURNING updated_at
	`

	var gender *string
	if input.Gender != "" {
		gender = &input.Gender
	}

	var nationality *string
	if input.Nationality != "" {
		nationality = &input.Nationality
	}

	var updatedAt time.Time
	err = dbConn.QueryRowContext(ctx, query,
		input.Name, input.Surname, input.Patronymic, input.Age,
		gender, nationality, id,
	).Scan(&updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Person not found",
			})
			return
		}
		handleDatabaseError(c, ctx, err, "Failed to update person")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"updated_at": updatedAt,
	})
}

func PatchPerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	var input models.UpdatePersonRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid input")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
			Details: err.Error(),
		})
		return
	}

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	query, args := buildPartialUpdateQuery(id, input)
	var updatedAt time.Time
	err = dbConn.QueryRowContext(ctx, query, args...).Scan(&updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Person not found",
			})
			return
		}
		handleDatabaseError(c, ctx, err, "Failed to partially update person")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"updated_at": updatedAt,
	})
}

func DeletePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Warn("Invalid ID format")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	dbConn, err := db.GetDB()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("Failed to get DB connection")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	result, err := dbConn.ExecContext(ctx, "DELETE FROM people WHERE id = $1", id)
	if err != nil {
		handleDatabaseError(c, ctx, err, "Failed to delete person")
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

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"rows_affected": rowsAffected,
	})
}

func handleDatabaseError(c *gin.Context, ctx context.Context, err error, message string) {
	log.WithContext(ctx).WithError(err).Error("Database operation failed")

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "database_error",
			Message: message,
			Details: pgErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error:   "database_error",
		Message: message,
	})
}

func buildFilterQuery(filter models.PersonFilter) (string, []interface{}) {
	query := `SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at 
              FROM people WHERE 1=1`
	var args []interface{}
	argPos := 1

	if filter.Name != "" {
		query += " AND name ILIKE $" + strconv.Itoa(argPos)
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}
	if filter.Surname != "" {
		query += " AND surname ILIKE $" + strconv.Itoa(argPos)
		args = append(args, "%"+filter.Surname+"%")
		argPos++
	}
	if filter.AgeFrom != nil {
		query += " AND age >= $" + strconv.Itoa(argPos)
		args = append(args, *filter.AgeFrom)
		argPos++
	}
	if filter.AgeTo != nil {
		query += " AND age <= $" + strconv.Itoa(argPos)
		args = append(args, *filter.AgeTo)
		argPos++
	}
	if filter.Gender != "" {
		query += " AND gender = $" + strconv.Itoa(argPos)
		args = append(args, filter.Gender)
		argPos++
	}
	if filter.Nationality != "" {
		query += " AND nationality = $" + strconv.Itoa(argPos)
		args = append(args, filter.Nationality)
		argPos++
	}

	query += " ORDER BY created_at DESC"
	return query, args
}

func buildPartialUpdateQuery(id int, input models.UpdatePersonRequest) (string, []interface{}) {
	query := "UPDATE people SET "
	var args []interface{}
	argPos := 1
	fields := 0

	if input.Name != nil {
		if fields > 0 {
			query += ", "
		}
		query += "name = $" + strconv.Itoa(argPos)
		args = append(args, *input.Name)
		argPos++
		fields++
	}
	if input.Surname != nil {
		if fields > 0 {
			query += ", "
		}
		query += "surname = $" + strconv.Itoa(argPos)
		args = append(args, *input.Surname)
		argPos++
		fields++
	}
	if input.Patronymic != nil {
		if fields > 0 {
			query += ", "
		}
		query += "patronymic = $" + strconv.Itoa(argPos)
		args = append(args, *input.Patronymic)
		argPos++
		fields++
	}
	if input.Age != nil {
		if fields > 0 {
			query += ", "
		}
		query += "age = $" + strconv.Itoa(argPos)
		args = append(args, *input.Age)
		argPos++
		fields++
	}
	if input.Gender != nil {
		if fields > 0 {
			query += ", "
		}
		query += "gender = $" + strconv.Itoa(argPos)
		args = append(args, *input.Gender)
		argPos++
		fields++
	}
	if input.Nationality != nil {
		if fields > 0 {
			query += ", "
		}
		query += "nationality = $" + strconv.Itoa(argPos)
		args = append(args, *input.Nationality)
		argPos++
		fields++
	}

	if fields == 0 {
		return "", nil
	}

	query += ", updated_at = NOW() WHERE id = $" + strconv.Itoa(argPos) + " RETURNING updated_at"
	args = append(args, id)

	return query, args
}
