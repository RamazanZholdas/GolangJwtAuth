package structs

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
