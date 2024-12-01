package handlers

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"glosbe/pkg/glosbe"
)

type TextHandler struct {
	b *tgbotapi.BotAPI
	c *glosbe.Client
}

func NewTextHandler(b *tgbotapi.BotAPI, c *glosbe.Client) *TextHandler {
	return &TextHandler{b: b, c: c}
}

func (h *TextHandler) Handle(ctx context.Context, update tgbotapi.Update) error {
	translation, err := h.c.Translate(ctx, &glosbe.TranslateRequest{Src: "ru", Dst: "sr", Text: update.Message.Text})
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, translation)
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = AddInlineButtonData("Примеры", "/examples")
	if _, err := h.b.Send(msg); err != nil {
		return err
	}
	return nil
}
