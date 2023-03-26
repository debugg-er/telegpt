package main

import (
	"log"
	"strings"
	"telegpt/bot"
	"telegpt/bot/config"
	"telegpt/bot/handler"
	"telegpt/bot/store"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	botAPI    *telegram.BotAPI
	userStore *store.UserStore
	userCache *store.UserCache
)

func init() {
	// Init user firestore
	var err error
	userStore, err = store.NewUserStore()
	if err != nil {
		panic(err)
	}
	log.Println("Connected to Firestore")

	// Create telegpt user userCache
	userCache = store.NewUserCache(userStore)
	log.Println("Created user cache store")

	// Create telegram bot
	botAPI, err = telegram.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		panic(err)
	}
	log.Printf("Authorized on Telegram Account %s", botAPI.Self.UserName)
}

func handle(update telegram.Update) {
	userSession, err := userCache.GetUserSession(update.Message.From.UserName)
	if err != nil {
		log.Println(err)
		return
	}
	userSession.Mu.Lock()
	defer userSession.Mu.Unlock()

	ctx := &bot.Context{
		BotAPI:    botAPI,
		UserStore: userStore,
		Session:   userSession.Value,
		Update:    update,
	}
	if strings.HasPrefix(update.Message.Text, "/token") {
		handler.HandleStartSettingToken(ctx)
		return
	}
	if userSession.Value.IsEnteringToken {
		handler.HandleEnterToken(ctx)
		return
	}
	if userSession.Value.OpenAIToken == "" {
		handler.SendTelegramMsg(ctx, "Please provide your Open AI key using /token")
		return
	}
	handler.HandleCompletion(ctx)
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
