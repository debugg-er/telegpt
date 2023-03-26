package bot

import (
	"telegpt/bot/models"
	"telegpt/bot/store"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	Context struct {
		Session   *models.UserSession
		UserStore *store.UserStore
		BotAPI    *telegram.BotAPI
		Update    telegram.Update
	}
)
