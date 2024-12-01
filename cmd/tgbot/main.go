package main

import (
	"context"
	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"glosbe/cmd/tgbot/handlers"
	"log"
	"os"

	"glosbe/pkg/glosbe"
)

func main() {
	ctx := context.Background()
	c := glosbe.NewClient(resty.New())
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	textHandler := handlers.NewTextHandler(bot, c)
	examplesHandler := handlers.NewExampleHandler(bot, c)
	for update := range updates {
		if update.Message != nil && update.Message.Command() == "" {
			log.Printf("NewMessage: [%s] %s", update.Message.From.UserName, update.Message.Text)
			if err = textHandler.Handle(ctx, update); err != nil {
				handleErr(err)
			}
		}
		if update.CallbackQuery != nil && update.CallbackData() == handlers.ExamplesData {
			msg := update.CallbackQuery.Message
			log.Printf("Examples: [%s] %s", msg.From.UserName, msg.Text)
			if err = examplesHandler.Handle(ctx, update); err != nil {
				handleErr(err)
			}
		}
		if update.CallbackQuery != nil && update.CallbackData() == handlers.NextExampleData {
			msg := update.CallbackQuery.Message
			log.Printf("NextExample: [%s] %s", msg.From.UserName, msg.Text)
			if err = examplesHandler.Next(ctx, update); err != nil {
				handleErr(err)
			}
		}
	}
}

func handleErr(err error) {
	log.Printf("[ERROR] err: %s", err.Error())
}
