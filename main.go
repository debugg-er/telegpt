package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"telegpt/config"
	"telegpt/models"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getCompletion(client *http.Client, question string) (chan string, error) {
	completionReqBody := []byte(fmt.Sprintf(`{"model":"gpt-3.5-turbo","stream":true,"messages":[{"role":"user","content":"%s"}]}`, question))
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(completionReqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.Get().OpenAIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(resp.Body)
	out := make(chan string)
	go func() {
		defer resp.Body.Close()
		defer close(out)
		for scanner.Scan() {
			chunk := scanner.Text()
			if !strings.HasPrefix(chunk, "data: ") {
				continue
			}
			var jsonChunk models.ChunkChatCompletion
			if err := json.Unmarshal([]byte(chunk[len("data: "):]), &jsonChunk); err != nil {
				out <- "parse chunk failed"
			}

			if jsonChunk.Choices[0].FinishReason == "stop" {
				return
			}

			// fmt.Println("Content: " + jsonChunk.Choices[0].Delta.Content)
			out <- jsonChunk.Choices[0].Delta.Content
		}
		if scanner.Err() != nil {
			out <- scanner.Err().Error()
		}
	}()
	return out, nil
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
					message = ""
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

func init() {
}

func main() {
	err := config.Load("chatbot")
	if err != nil {
		panic(err)
	}
	bot, err := telegram.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	// bot.Debug = true // Enable debugging
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := telegram.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		out, err := getCompletion(client, update.Message.Text)
		if err != nil {
			msg := telegram.NewMessage(update.Message.Chat.ID, err.Error())
			bot.Send(msg)
			continue
		}

		outInterval := chan2IntervalChan(out, time.Millisecond*500)
		msg := telegram.NewMessage(update.Message.Chat.ID, <-outInterval)
		message, _ := bot.Send(msg)
		for chunk := range outInterval {
			fmt.Print(chunk)
			editMsg := telegram.NewEditMessageText(update.Message.Chat.ID, message.MessageID, message.Text+chunk)
			message, _ = bot.Send(editMsg)
		}
	}
}
