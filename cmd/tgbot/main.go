package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"glosbe/pkg/glosbe"
)

var queries = sync.Map{}

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
		if update.Message != nil && update.Message.Command() == "" {
			log.Printf("NewMessage: [%s] %s", update.Message.From.UserName, update.Message.Text)

			resp, err := c.Translate(ctx, &glosbe.TranslateRequest{Src: "ru", Dst: "sr", Text: update.Message.Text})
			if err != nil {
				if !errors.Is(err, glosbe.ErrNotFound) {
					handleErr(err)
					continue
				}
				respText := fmt.Sprintf("Ничего не нашлось. Попробуй изменить запрос")
				if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, respText)); err != nil {
					handleErr(err)
				}
				continue
			}

			if len(resp.Examples) == 0 {
				handleErr(err)
				continue
			}

			respText := formatExample(resp.Examples[0])
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, respText)
			msg.ReplyToMessageID = update.Message.MessageID
			if len(resp.Examples) > 1 {
				queries.Store(update.Message.Chat.ID, resp.Examples[1:])
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Ещё", "/next"),
					),
				)
			}
			if _, err := bot.Send(msg); err != nil {
				log.Printf("[ERROR] msg: %s; err: %s", update.Message.Text, err.Error())
			}
		}
		if update.CallbackQuery != nil && update.CallbackData() == "/next" {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				handleErr(err)
				continue
			}
			respText := "Примеры кончились"
			var examples []glosbe.Example
			loaded, ok := queries.LoadAndDelete(update.CallbackQuery.Message.Chat.ID)
			if ok {
				examples, _ = loaded.([]glosbe.Example)
				respText = formatExample(examples[0])
				examples = examples[1:]
			}

			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, respText)
			if update.CallbackQuery.Message.ReplyMarkup != nil {
				editMarkupMsg := tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)},
				)
				if _, err := bot.Send(editMarkupMsg); err != nil {
					handleErr(err)
					continue
				}
			}
			if len(examples) > 0 {
				queries.Store(update.CallbackQuery.Message.Chat.ID, examples)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Ещё", "/next"),
					),
				)
			}
			if _, err := bot.Send(msg); err != nil {
				handleErr(err)
				continue
			}
		}
	}
}

func formatExample(example glosbe.Example) string {
	return fmt.Sprintf("%s -> %s\n", example.SrcLangText, example.DstLangText)
}

func handleErr(err error) {
	log.Printf("[ERROR] err: %s", err.Error())
}
