package handlers

import (
	"context"
	"database/sql"
	"go-people-api/models"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRow кастомная реализация sql.Row для тестов
type MockRow struct {
	scanErr error
	values  []interface{}
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.scanErr != nil {
		return m.scanErr
	}

	for i := range dest {
		if i >= len(m.values) {
			continue
		}

		switch d := dest[i].(type) {
		case *int:
			if v, ok := m.values[i].(int); ok {
				*d = v
			}
		case *string:
			if v, ok := m.values[i].(string); ok {
				*d = v
			}
		case **string:
			if v, ok := m.values[i].(string); ok {
				*d = &v
			}
		case *time.Time:
			if v, ok := m.values[i].(time.Time); ok {
				*d = v
			}
		case *sql.NullString:
			if v, ok := m.values[i].(sql.NullString); ok {
				*d = v
			}
		}
	}
	return nil
}

// MockPersonService мок сервиса обогащения данных
type MockPersonService struct {
	mock.Mock
}

func (m *MockPersonService) Enrich(ctx context.Context, name string) (*models.Person, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*models.Person), args.Error(1)
}

// MockDBWrapper мок обертки для DB
type MockDBWrapper struct {
	mock.Mock
}

func (m *MockDBWrapper) GetDB() (*sql.DB, error) {
	args := m.Called()
	return args.Get(0).(*sql.DB), args.Error(1)
}

// MockSQLDB мок sql.DB
type MockSQLDB struct {
	mock.Mock
}

func (m *MockSQLDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockSQLDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

func (m *MockSQLDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {

	return &sql.Row{}
}

type DBWrapper interface {
	GetDB() (*sql.DB, error)
}

// TestCreatePerson тест для CreatePerson
func TestCreatePerson(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		input          string
		mockEnrich     func(*MockPersonService)
		mockDB         func(*MockDBWrapper, *MockSQLDB)
		expectedStatus int
	}{
		{
			name:  "successful creation",
			input: `{"name": "Иван", "surname": "Иванов"}`,
			mockEnrich: func(m *MockPersonService) {
				m.On("Enrich", mock.Anything, "Иван").Return(&models.Person{
					Gender:      "male",
					Age:         30,
					Nationality: "RU",
				}, nil)
			},
			mockDB: func(dbWrapper *MockDBWrapper, db *MockSQLDB) {
				dbWrapper.On("GetDB").Return(db, nil)
				db.On("QueryRowContext", mock.Anything, mock.Anything, mock.Anything).Return(&MockRow{
					values: []interface{}{1, now, now},
				})
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockPersonService{}
			dbWrapper := &MockDBWrapper{}
			dbMock := &MockSQLDB{}

			tt.mockEnrich(service)
			tt.mockDB(dbWrapper, dbMock)

			h := &Handler{
				PersonService: service,
				DB:            dbWrapper,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/people", strings.NewReader(tt.input))
			c.Request.Header.Set("Content-Type", "application/json")

			h.CreatePerson(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			service.AssertExpectations(t)
			dbWrapper.AssertExpectations(t)
			dbMock.AssertExpectations(t)
		})
	}
}

// TestGetPeople тесты для GetPeople
func TestGetPeople(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockDB         func(*MockDBWrapper, *MockSQLDB)
		expectedStatus int
	}{
		{
			name:  "successful fetch",
			query: "",
			mockDB: func(dbWrapper *MockDBWrapper, db *MockSQLDB) {
				dbWrapper.On("GetDB").Return(db, nil)
				rows := &sql.Rows{}
				db.On("QueryContext", mock.Anything,
					mock.Anything,
					mock.Anything).Return(rows, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbWrapper := &MockDBWrapper{}
			dbMock := &MockSQLDB{}
			tt.mockDB(dbWrapper, dbMock)

			h := &Handler{
				DB: dbWrapper,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/people"+tt.query, nil)

			h.GetPeople(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			dbWrapper.AssertExpectations(t)
			dbMock.AssertExpectations(t)
		})
	}
}

// TestGetPersonByID тесты для GetPersonByID
func TestGetPersonByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockDB         func(*MockDBWrapper, *MockSQLDB)
		expectedStatus int
	}{
		{
			name: "successful fetch",
			id:   "1",
			mockDB: func(dbWrapper *MockDBWrapper, db *MockSQLDB) {
				dbWrapper.On("GetDB").Return(db, nil)
				row := &sql.Row{}
				db.On("QueryRowContext", mock.Anything,
					mock.Anything,
					mock.Anything).Return(row)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbWrapper := &MockDBWrapper{}
			dbMock := &MockSQLDB{}
			tt.mockDB(dbWrapper, dbMock)

			h := &Handler{
				DB: dbWrapper,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/people/"+tt.id, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			h.GetPersonByID(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			dbWrapper.AssertExpectations(t)
			dbMock.AssertExpectations(t)
		})
	}
}

// Handler структура для инъекции зависимостей
type Handler struct {
	PersonService PersonService
	DB            *MockDBWrapper
}

// CreatePerson метод с инъекцией зависимостей
func (h *Handler) CreatePerson(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	var input models.Person
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid input data",
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

	enriched, err := h.PersonService.Enrich(ctx, input.Name)
	if err != nil {
		// Продолжаем без обогащения
	}

	result := mergePersonData(&input, enriched)

	dbConn, err := h.DB.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	query := `INSERT INTO people (name, surname) VALUES ($1, $2) RETURNING id`
	err = dbConn.QueryRowContext(ctx, query, result.Name, result.Surname).Scan(&result.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create person",
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPeople метод с инъекцией зависимостей
func (h *Handler) GetPeople(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbConn, err := h.DB.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	rows, err := dbConn.QueryContext(ctx, "SELECT id, name FROM people")
	if err != nil {
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
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			continue
		}
		people = append(people, p)
	}

	c.JSON(http.StatusOK, people)
}

// GetPersonByID метод с инъекцией зависимостей
func (h *Handler) GetPersonByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Person ID must be an integer",
		})
		return
	}

	dbConn, err := h.DB.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Database unavailable",
		})
		return
	}

	var person models.Person
	err = dbConn.QueryRowContext(ctx, "SELECT id, name FROM people WHERE id = $1", id).
		Scan(&person.ID, &person.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Person not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch person",
		})
		return
	}

	c.JSON(http.StatusOK, person)
}
