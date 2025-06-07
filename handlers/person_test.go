package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-people-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPersonService struct {
	mock.Mock
}

func (m *MockPersonService) Enrich(ctx context.Context, name string) (*models.Person, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*models.Person), args.Error(1)
}

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(sql.Result), argsMock.Error(1)
}

func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(*sql.Rows), argsMock.Error(1)
}

func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(*sql.Row)
}

type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestCreatePerson(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mockEnrich     func(*MockPersonService)
		mockDB         func(*MockDatabase)
		expectedStatus int
	}{
		{
			name:  "successful creation",
			input: `{"name": "Иван", "surname": "Иванов", "patronymic": "Иванович"}`,
			mockEnrich: func(m *MockPersonService) {
				m.On("Enrich", mock.Anything, "Иван").Return(&models.Person{
					Gender:      "male",
					Age:         30,
					Nationality: "RU",
				}, nil)
			},
			mockDB: func(m *MockDatabase) {
				mockRow := &sql.Row{}
				m.On("QueryRowContext", mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "INSERT INTO people") &&
							strings.Contains(query, "VALUES ($1, $2, $3, $4, $5, $6)")
					}),
					[]interface{}{"Иван", "Иванов", "Иванович", "male", 30, "RU"},
				).Return(mockRow)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid input",
			input:          `{"name": "Иван"}`,
			mockEnrich:     func(m *MockPersonService) {},
			mockDB:         func(m *MockDatabase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "enrichment failed",
			input: `{"name": "Иван", "surname": "Иванов", "patronymic": "Иванович"}`,
			mockEnrich: func(m *MockPersonService) {
				m.On("Enrich", mock.Anything, "Иван").Return(&models.Person{}, errors.New("enrichment error"))
			},
			mockDB:         func(m *MockDatabase) {},
			expectedStatus: http.StatusFailedDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockPersonService{}
			mockDB := &MockDatabase{}
			tt.mockEnrich(mockService)
			tt.mockDB(mockDB)

			originalService := personService
			originalDB := database
			personService = mockService
			database = mockDB
			defer func() {
				personService = originalService
				database = originalDB
			}()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/people", strings.NewReader(tt.input))
			c.Request.Header.Set("Content-Type", "application/json")

			CreatePerson(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetPeople(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockDB         func(*MockDatabase)
		expectedStatus int
	}{
		{
			name:  "successful fetch",
			query: "",
			mockDB: func(m *MockDatabase) {
				rows := &sql.Rows{}
				m.On("QueryContext", mock.Anything,
					"SELECT id, name, surname, patronymic, gender, age, nationality, created_at, updated_at FROM people ORDER BY id LIMIT $1 OFFSET $2",
					[]interface{}{10, 0},
				).Return(rows, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "with filters",
			query: "?name=Иван&age_from=20&age_to=30",
			mockDB: func(m *MockDatabase) {
				rows := &sql.Rows{}
				m.On("QueryContext", mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "SELECT") &&
							strings.Contains(query, "WHERE name ILIKE")
					}),
					mock.Anything,
				).Return(rows, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDatabase{}
			tt.mockDB(mockDB)

			originalDB := database
			database = mockDB
			defer func() { database = originalDB }()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/people"+tt.query, nil)

			GetPeople(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestBuildFilterQuery(t *testing.T) {
	tests := []struct {
		name     string
		filter   models.PersonFilter
		limit    int
		offset   int
		expected string
	}{
		{
			name:     "empty filter",
			filter:   models.PersonFilter{},
			limit:    10,
			offset:   0,
			expected: "SELECT id, name, surname, patronymic, gender, age, nationality, created_at, updated_at FROM people ORDER BY id LIMIT $1 OFFSET $2",
		},
		{
			name: "name filter",
			filter: models.PersonFilter{
				Name: "Иван",
			},
			limit:    5,
			offset:   10,
			expected: "SELECT id, name, surname, patronymic, gender, age, nationality, created_at, updated_at FROM people WHERE name ILIKE $1 ORDER BY id LIMIT $2 OFFSET $3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, _ := buildFilterQuery(tt.filter, tt.limit, tt.offset)
			assert.Equal(t, tt.expected, query)
		})
	}
}
