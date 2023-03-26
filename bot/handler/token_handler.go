package handler

import (
	"log"
	"telegpt/bot"
)

func HandleStartSettingToken(c *bot.Context) {
	c.Session.IsEnteringToken = true
	SendTelegramMsg(c, "Please enter you token")
}

func HandleEnterToken(c *bot.Context) {
	token := c.Update.Message.Text
	if token == "" {
		SendTelegramMsg(c, "Please enter you token")
		return
	}
	c.Session.OpenAIToken = token
	if err := c.UserStore.CreateOrUpdateUser(c.Session.User); err != nil {
		log.Println(err)
		SendTelegramMsg(c, "Internal server error")
		return
	}

	c.Session.IsEnteringToken = false
	SendTelegramMsg(c, "Saved your token ("+token+")")
}
