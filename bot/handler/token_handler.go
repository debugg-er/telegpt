package handler

import (
	"log"
	"strings"
	"telegpt/bot"
	"telegpt/bot/models"
)

func HandleSetToken(c *bot.Context) {
	m := c.Update.Message
	token := strings.TrimSpace(m.Text[len("/tokens"):])
	if token == "" {
		sendTelegramMessage(c, "Your token is empty")
		return
	}
	user := models.NewUser(m.From.UserName, token, nil)
	if err := c.UserStore.SetUser(*user); err != nil {
		log.Println(err)
		sendTelegramMessage(c, "Internal server error")
		return
	}
	c.Cache.SetUserSession(m.From.UserName, user)

	sendTelegramMessage(c, "Saved your token ("+token+")")
}
