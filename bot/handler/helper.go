package handler

import (
	"telegpt/bot"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func SendTelegramMsg(c *bot.Context, content string) (telegram.Message, error) {
	msg := telegram.NewMessage(c.Update.Message.Chat.ID, content)
	msg.ParseMode = telegram.ModeMarkdown
	return c.BotAPI.Send(msg)
}

func ReplyTelegramMsg(c *bot.Context, content string) (telegram.Message, error) {
	msg := telegram.NewMessage(c.Update.Message.Chat.ID, content)
	msg.ParseMode = telegram.ModeMarkdown
	msg.ReplyToMessageID = c.Update.Message.MessageID
	return c.BotAPI.Send(msg)
}

func EditTelegramMsg(c *bot.Context, msg telegram.Message, content string) (telegram.Message, error) {
	editMsg := telegram.NewEditMessageText(msg.Chat.ID, msg.MessageID, content)
	editMsg.ParseMode = telegram.ModeMarkdown
	return c.BotAPI.Send(editMsg)
}
