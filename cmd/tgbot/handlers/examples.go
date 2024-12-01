package handlers

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"glosbe/pkg/glosbe"
	"log"
	"sync"
)

const (
	ExamplesData    = "/examples"
	NextExampleData = "/next"
)

type ExampleHandler struct {
	b *tgbotapi.BotAPI
	c *glosbe.Client
	d *sync.Map
}

func NewExampleHandler(b *tgbotapi.BotAPI, c *glosbe.Client) *ExampleHandler {
	return &ExampleHandler{b: b, c: c, d: new(sync.Map)}
}

func (h *ExampleHandler) Handle(ctx context.Context, update tgbotapi.Update) error {
	if err := h.handleQuery(update); err != nil {
		return err
	}
	msg := update.CallbackQuery.Message
	text := msg.Text
	if msg.ReplyToMessage != nil {
		text = msg.ReplyToMessage.Text
	}
	examples, err := h.c.Examples(ctx, &glosbe.TranslateRequest{Src: "ru", Dst: "sr", Text: text})
	if err != nil && !errors.Is(err, glosbe.ErrNotFound) {
		return err
	}
	if err = h.sendExample(update, examples); err != nil {
		return err
	}
	if len(examples) > 1 {
		h.d.Store(msg.Chat.ID, examples[1:])
	}
	return nil
}

func (h *ExampleHandler) Next(_ context.Context, update tgbotapi.Update) error {
	if err := h.handleQuery(update); err != nil {
		return err
	}
	examples := h.loadExamples(update.CallbackQuery.Message.Chat.ID)
	if len(examples) == 0 {
		return nil
	}
	if err := h.sendExample(update, examples); err != nil {
		return err
	}
	if len(examples) > 1 {
		h.d.Store(update.CallbackQuery.Message.Chat.ID, examples[1:])
	}
	return nil
}

func (h *ExampleHandler) loadExamples(chatID int64) []glosbe.Example {
	var examples []glosbe.Example
	loaded, ok := h.d.LoadAndDelete(chatID)
	if !ok {
		return nil
	}
	examples, ok = loaded.([]glosbe.Example)
	if !ok {
		log.Println("[ERROR] failed to cast loaded data to glosbe's examples")
		return nil
	}
	return examples
}

func (h *ExampleHandler) handleQuery(update tgbotapi.Update) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := h.b.Request(callback); err != nil {
		return err
	}
	if update.CallbackQuery.Message.ReplyMarkup != nil {
		if err := RmInlineButtonData(h.b, update.CallbackQuery.Message); err != nil {
			return err
		}
	}
	return nil
}

func (h *ExampleHandler) sendExample(update tgbotapi.Update, examples []glosbe.Example) error {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Примеров нет(")
	if len(examples) > 0 {
		msg.Text = formatExample(examples[0])
	}
	if len(examples) > 1 {
		msg.ReplyMarkup = AddInlineButtonData("Ещё", "/next")
	}

	if _, err := h.b.Send(msg); err != nil {
		return err
	}
	return nil
}

func formatExample(example glosbe.Example) string {
	return fmt.Sprintf("%s\n%s\n", example.SrcLangText, example.DstLangText)
}
