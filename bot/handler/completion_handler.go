package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"telegpt/bot"
	"telegpt/bot/models"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func HandleCompletion(c *bot.Context) {
	m := c.Update.Message
	userSession := c.Cache.GetUserSection(m.From.UserName)
	if userSession == nil {
		user, err := c.UserStore.GetUserById(m.From.UserName)
		if err != nil {
			tellProblem(c, err.Error())
			return
		}
		if user == nil {
			tellProblem(c, "Please provide your Open AI key using \"/token <your key>\"")
			return
		}
		c.Cache.SetUserSession(user.UserId, user)
		userSession = c.Cache.GetUserSection(m.From.UserName)
	}

	initMsg, _ := c.BotAPI.Send(telegram.NewMessage(m.Chat.ID, "Processing..."))

	out, err := getGptCompletion(userSession, m.Text)
	if err != nil {
		editMsg := telegram.NewEditMessageText(m.Chat.ID, initMsg.MessageID, err.Error())
		c.BotAPI.Send(editMsg)
		return
	}

	completion := ""
	for chunk := range out {
		completion += chunk
		editMsg := telegram.NewEditMessageText(m.Chat.ID, initMsg.MessageID, completion)
		c.BotAPI.Send(editMsg)
	}
}

func getGptCompletion(session *models.UserSession, question string) (chan string, error) {
	completionReqBody := []byte(fmt.Sprintf(`{"model":"gpt-3.5-turbo","stream":true,"messages":[{"role":"user","content":"%s"}]}`, question))
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(completionReqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+session.UserOpenAIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := session.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthorized to Open AI, double check your key")
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
	return chan2IntervalChan(out, time.Millisecond*500), nil
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
