package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	OpenAIKey                   string `mapstructure:"OPEN_AI_KEY"`
	TelegramBotToken            string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	FirebaseCredential          string `mapstructure:"FIREBASE_CREDENTIAL"`
	FirebaseProjectId           string `mapstructure:"FIREBASE_PROJECT_ID"`
	FirebaseTokenCollectionName string `mapstructure:"FIREBASE_TOKEN_COLLECTION_NAME"`
}

var (
	once   *sync.Once = &sync.Once{}
	config *Config    = new(Config)
)

func load() error {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(".env")

	if err := v.ReadInConfig(); err != nil {
		return err
	}
	if err := v.Unmarshal(config); err != nil {
		return err
	}
	// fmt.Println(config)
	return nil
}

func Get() *Config {
	once.Do(func() {
		if err := load(); err != nil {
			panic(err)
		}
	})
	return config
}
