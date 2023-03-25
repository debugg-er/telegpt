package main

import (
	"log"
	"strings"
	"telegpt/bot"
	"telegpt/bot/config"
	"telegpt/bot/handler"
	"telegpt/bot/store"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	cache     *store.Cache
	botAPI    *telegram.BotAPI
	userStore *store.UserStore
)

func init() {
	var err error
	// Init user firestore
	userStore, err = store.NewUserStore()
	if err != nil {
		panic(err)
	}
	log.Println("Connected to Firestore")

	// Create telegpt cache
	cache = store.NewCache(time.Second * 10)
	log.Println("Created user cache store")

	// Create telegram bot
	botAPI, err = telegram.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		panic(err)
	}
	log.Printf("Authorized on Telegram Account %s", botAPI.Self.UserName)
}

func handle(update telegram.Update) {
	ctx := &bot.Context{
		BotAPI:    botAPI,
		Cache:     cache,
		Update:    update,
		UserStore: userStore,
	}

	if strings.HasPrefix(update.Message.Text, "/token ") {
		handler.HandleSetToken(ctx)
	} else {
		handler.HandleCompletion(ctx)
	}
}

func main() {
	u := telegram.NewUpdate(0)
	u.Timeout = 60

	updates, err := botAPI.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		go handle(update)
	}
}
