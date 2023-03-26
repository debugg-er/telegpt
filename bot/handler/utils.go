package handler

import (
	"telegpt/bot"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendTelegramMessage(c *bot.Context, content string) (telegram.Message, error) {
	msg := telegram.NewMessage(c.Update.Message.Chat.ID, content)
	msg.ParseMode = telegram.ModeMarkdown
	return c.BotAPI.Send(msg)
}

func editTelegramMsg(c *bot.Context, msg telegram.Message, content string) (telegram.Message, error) {
	editMsg := telegram.NewEditMessageText(msg.Chat.ID, msg.MessageID, content)
	editMsg.ParseMode = telegram.ModeMarkdown
	return c.BotAPI.Send(editMsg)
}

func chan2IntervalChan(in chan string, duration time.Duration) chan string {
	out := make(chan string)
	ticker := time.NewTicker(duration)
	message := ""
	go func() {
		defer close(out)
		for {
			select {
			case chunk, ok := <-in:
				if !ok {
					out <- message
					return
				}
				message += chunk
			case <-ticker.C:
				out <- message
				message = ""
			}
		}
	}()
	return out
}

func takeNLastItems[T any](arr []T, n int) []T {
	if len(arr) <= n {
		return arr
	}
	return arr[len(arr)-n:]
}
