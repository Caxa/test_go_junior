package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseOperations(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t) // Ваша функция инициализации тестовой БД

	t.Run("Create Person", func(t *testing.T) {
		person := Person{
			Name:        "Test",
			Surname:     "User",
			Age:         30,
			Gender:      "male",
			Nationality: "US",
		}
		id, err := db.CreatePerson(ctx, person)
		assert.NoError(t, err)
		assert.NotZero(t, id)
	})

	t.Run("Get Person", func(t *testing.T) {
		person, err := db.GetPerson(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, "Test", person.Name)
	})
}
