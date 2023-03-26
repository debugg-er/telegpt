package store

import (
	"telegpt/bot/models"
	"telegpt/bot/util"
)

type (
	UserCache struct {
		cache *util.CacheLoader[string, *models.UserSession]
	}
)

func NewUserCache(userStore *UserStore) *UserCache {
	return &UserCache{
		cache: util.NewCacheLoader(loadUser(userStore)),
	}
}

func (uc UserCache) GetUserSession(userId string) (*util.Item[*models.UserSession], error) {
	return uc.cache.Get(userId)
}

func loadUser(userStore *UserStore) func(userId string) (*models.UserSession, error) {
	return func(userId string) (*models.UserSession, error) {
		user, err := userStore.GetUserById(userId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			user = &models.User{UserId: userId}
			if userStore.CreateOrUpdateUser(user); err != nil {
				return nil, err
			}
		}
		return models.NewUserSession(user), nil
	}
}
