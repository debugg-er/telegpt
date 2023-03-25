package handler

import (
	"strings"
	"telegpt/bot"
	"telegpt/bot/models"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func HandleSetToken(c *bot.Context) {
	m := c.Update.Message
	token := strings.TrimSpace(m.Text[len("/tokens"):])
	if token == "" {
		tellProblem(c, "Your token is empty")
		return
	}
	user := models.NewUser(m.From.UserName, token)
	if err := c.UserStore.SetUser(*user); err != nil {
		tellProblem(c, err.Error())
		return
	}
	c.Cache.SetUserSession(m.From.UserName, user)

	msg := telegram.NewMessage(m.Chat.ID, "Created/Updated your token to "+token)
	c.BotAPI.Send(msg)
}
