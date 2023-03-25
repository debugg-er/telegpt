package handler

import (
	"telegpt/bot"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func tellProblem(c *bot.Context, message string) {
	msg := telegram.NewMessage(c.Update.Message.Chat.ID, message)
	c.BotAPI.Send(msg)
}
