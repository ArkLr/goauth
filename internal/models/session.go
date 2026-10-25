package models

type Session struct {
	UserId       string
	SessionToken string
	CSRFToken    string
	Expires      int
}
