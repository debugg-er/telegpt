package bot

import (
	"telegpt/bot/store"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	Context struct {
		Cache     *store.Cache
		UserStore *store.UserStore
		BotAPI    *telegram.BotAPI
		Update    telegram.Update
	}
)
