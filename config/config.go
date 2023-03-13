package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	OpenAIKey        string `mapstructure:"openAIKey"`
	TelegramBotToken string `mapstructure:"telegramBotToken"`
}

var (
	config *Config = new(Config)
)

func Load(name string) error {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(name)
	v.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	if err := v.Unmarshal(config); err != nil {
		return err
	}
	fmt.Println(config)
	return nil
}

func Get() *Config {
	if config == nil {
		panic("Config has not been initialized")
	}
	return config
}
