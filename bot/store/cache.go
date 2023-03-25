package store

import (
	"telegpt/bot/models"
	"time"
)

type (
	Cache struct {
		expiration   time.Duration
		userSessions map[string]*models.UserSession
	}
)

func NewCache(expiration time.Duration) *Cache {
	return &Cache{
		expiration:   expiration,
		userSessions: make(map[string]*models.UserSession),
	}
}

func (c *Cache) GetUserSection(userId string) *models.UserSession {
	userSession, ok := c.userSessions[userId]
	if !ok {
		return nil
	}
	userSession.LastAccess = time.Now()
	return userSession
}

func (c *Cache) SetUserSession(userId string, user *models.User) {
	session := models.NewUserSection(user)
	c.userSessions[userId] = session

	go func(holdedLastAccess time.Time) {
		for {
			<-time.After(c.expiration - time.Since(holdedLastAccess))
			// Remove after expiration time if user don't access their session
			if session.LastAccess == holdedLastAccess {
				// Delete only when user haven't change their token
				if c.userSessions[userId] == session {
					delete(c.userSessions, userId)
				}
				return
			} else {
				holdedLastAccess = session.LastAccess
			}
		}
	}(session.LastAccess)
}
