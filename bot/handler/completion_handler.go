package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"telegpt/bot"
	"telegpt/bot/models"
	"time"
)

func HandleCompletion(c *bot.Context) {
	m := c.Update.Message
	userSession := c.Cache.GetUserSession(m.From.UserName)
	if userSession == nil {
		user, err := c.UserStore.GetUserById(m.From.UserName)
		if err != nil {
			log.Println(err)
			sendTelegramMessage(c, "Internal server error")
			return
		}
		if user == nil {
			sendTelegramMessage(c, "Please provide your Open AI key using \"/token <your key>\"")
			return
		}
		userSession = c.Cache.SetUserSession(user.UserId, user)
	}

	initMsg, _ := sendTelegramMessage(c, "Processing...")

	userSession.GptMessages = append(userSession.GptMessages, models.GptMessageHistory{
		Role:    "user",
		Content: m.Text,
	})
	out, err := getGptCompletion(userSession, m.Text)
	if err != nil {
		log.Println(err)
		editTelegramMsg(c, initMsg, "Error occur when calling to Open AI")
		return
	}

	completion := ""
	for chunk := range out {
		completion += chunk
		editTelegramMsg(c, initMsg, completion)
	}

	// Update new message to user session and user store
	userSession.GptMessages = append(userSession.GptMessages, models.GptMessageHistory{
		Role:    "assistant",
		Content: completion,
	})
	userSession.GptMessages = takeNLastItems(userSession.GptMessages, 10)
	if err := c.UserStore.SetUser(*userSession.User); err != nil {
		log.Println(err)
	}
}

func getGptCompletion(session *models.UserSession, question string) (chan string, error) {
	completionReqBody, err := json.Marshal(models.CompletionReqBody{
		Model:    "gpt-3.5-turbo",
		Stream:   true,
		Messages: session.GptMessages,
	})
	if err != nil {
		return nil, err
	}
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
			var jsonChunk models.ChunkedCompletionResp
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
