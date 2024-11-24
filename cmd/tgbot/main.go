package main

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"glosbe/pkg/glosbe"
	"log"
	"os"
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

	for update := range updates {
		if update.Message != nil && update.Message.Command() == "" { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			var respText string
			resp, err := c.Translate(ctx, &glosbe.TranslateRequest{Src: "ru", Dst: "sr", Text: update.Message.Text})
			if err != nil {
				respText = fmt.Sprintf("[%s] %s", update.Message.From.UserName, err.Error())
			} else {
				for _, example := range resp.Examples {
					respText += fmt.Sprintf("%s -> %s\n", example.SrcLangText, example.DstLangText)
				}
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, respText)
			msg.ReplyToMessageID = update.Message.MessageID

			if _, err := bot.Send(msg); err != nil {
				log.Printf("[ERROR] msg: %s; err: %s", update.Message.Text, err.Error())
			}
		}
	}
}
