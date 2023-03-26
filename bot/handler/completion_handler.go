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
	"telegpt/bot/util"
	"time"
)

func HandleCompletion(c *bot.Context) {
	initMsg, _ := ReplyTelegramMsg(c, "Processing...")

	// Update user question
	c.Session.GptMessages = append(c.Session.GptMessages, models.GptMessageHistory{
		Role:    "user",
		Content: c.Update.Message.Text,
	})
	// Do completion
	out, err := getGptCompletion(c.Session)
	if err != nil {
		log.Println(err)
		EditTelegramMsg(c, initMsg, "Error occur when calling to Open AI")
		return
	}
	completion := ""
	for chunk := range out {
		completion += chunk
		EditTelegramMsg(c, initMsg, completion)
	}

	// Update new message to user session and user store
	c.Session.GptMessages = append(c.Session.GptMessages, models.GptMessageHistory{
		Role:    "assistant",
		Content: completion,
	})
	c.Session.GptMessages = util.TakeNLastItems(c.Session.GptMessages, 10)
	if err := c.UserStore.CreateOrUpdateUser(c.Session.User); err != nil {
		log.Println(err)
	}
}

// You must update your question to `GptMessages` field in `session`
func getGptCompletion(session *models.UserSession) (chan string, error) {
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
	req.Header.Set("Authorization", "Bearer "+session.OpenAIToken)
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

			out <- jsonChunk.Choices[0].Delta.Content
		}
		if scanner.Err() != nil {
			out <- scanner.Err().Error()
		}
	}()
	return util.Chan2IntervalChan(out, time.Millisecond*500), nil
}
