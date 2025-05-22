package models

type League struct {
	ID          string `json:"id" db:"id"`
	League_Name string `json:"league_name" db:"league_name"`
}
