package models

type FullUser struct {
	ID         string `json:"id" db:"id"`
	First_Name string `json:"first_name" db:"first_name"`
	Last_Name  string `json:"last_name" db:"last_name"`
	Email      string `json:"email" db:"email"`
}

type UserIdPasswordHash struct {
	User_Id       string
	Password_Hash string
}

type APIUser struct {
	First_Name string `json:"first_name" db:"first_name"`
	Last_Name  string `json:"last_name" db:"last_name"`
	Email      string `json:"email" db:"email"`
}

type UserLeagueAssociation struct {
	User_Id   string `json:"user_id" db:"user_id"`
	League_id string `json:"league_id" db:"league_id"`
}

type SignupRequest struct {
	First_Name string `json:"first_name"`
	Last_Name  string `json:"last_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}
