package models

import (
	"net/http"
)

type (
	User struct {
		UserId      string              `firestore:"userId"`
		OpenAIToken string              `firestore:"token"`
		GptMessages []GptMessageHistory `firestore:"gptMessageHistory"`
	}
	UserSession struct {
		*User
		HttpClient      http.Client
		IsEnteringToken bool
	}
)

func NewUserSession(user *User) *UserSession {
	return &UserSession{
		User:            user,
		HttpClient:      http.Client{},
		IsEnteringToken: false,
	}
}
