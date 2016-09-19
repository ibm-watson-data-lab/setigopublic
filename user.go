package main

type User struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}
