package models

type Person struct {
	ID          int    `json:"id"`
	Name        string `json:"name" binding:"required"`
	Surname     string `json:"surname" binding:"required"`
	Patronymic  string `json:"patronymic"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}
