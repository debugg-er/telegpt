package models

import (
	"net/http"
	"time"
)

type (
	User struct {
		UserId          string              `firestore:"userId"`
		UserOpenAIToken string              `firestore:"token"`
		GptMessages     []GptMessageHistory `firestore:"gptMessageHistory"`
	}
	UserSession struct {
		*User
		HttpClient http.Client
		LastAccess time.Time
	}
)

func NewUser(userId string, openAIToken string, gptMessages []GptMessageHistory) *User {
	return &User{
		UserId:          userId,
		UserOpenAIToken: openAIToken,
		GptMessages:     gptMessages,
	}
}

func NewUserSection(user *User) *UserSession {
	return &UserSession{
		User:       user,
		HttpClient: http.Client{},
		LastAccess: time.Now(),
	}
}
