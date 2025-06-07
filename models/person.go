package models

import (
	"time"
)

// Person представляет информацию о человеке
type Person struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" binding:"required,min=2,max=100" db:"name"`
	Surname     string    `json:"surname" binding:"required,min=2,max=100" db:"surname"`
	Patronymic  string    `json:"patronymic,omitempty" binding:"max=100" db:"patronymic"`
	Gender      string    `json:"gender,omitempty" binding:"omitempty,oneof=male female other" db:"gender"`
	Age         int       `json:"age,omitempty" binding:"omitempty,min=1,max=120" db:"age"`
	Nationality string    `json:"nationality,omitempty" binding:"omitempty,len=2" db:"nationality"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// PersonFilter содержит параметры фильтрации для поиска людей
type PersonFilter struct {
	Name        string `json:"name,omitempty" form:"name"`
	Surname     string `json:"surname,omitempty" form:"surname"`
	Gender      string `json:"gender,omitempty" form:"gender"`
	AgeFrom     *int   `json:"age_from,omitempty" form:"age_from"`
	AgeTo       *int   `json:"age_to,omitempty" form:"age_to"`
	Nationality string `json:"nationality,omitempty" form:"nationality"`
}

// UpdatePersonRequest содержит поля для частичного обновления
type UpdatePersonRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Surname     *string `json:"surname,omitempty" binding:"omitempty,min=2,max=100"`
	Patronymic  *string `json:"patronymic,omitempty" binding:"omitempty,max=100"`
	Gender      *string `json:"gender,omitempty" binding:"omitempty,oneof=male female other"`
	Age         *int    `json:"age,omitempty" binding:"omitempty,min=1,max=120"`
	Nationality *string `json:"nationality,omitempty" binding:"omitempty,len=2"`
}

// ErrorResponse стандартный формат ответа об ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
