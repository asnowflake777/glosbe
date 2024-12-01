package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func AddInlineButtonData(text, data string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, data),
		),
	)
}

func RmInlineButtonData(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	editMarkupMsg := tgbotapi.NewEditMessageReplyMarkup(
		message.Chat.ID,
		message.MessageID,
		tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)},
	)
	if _, err := bot.Send(editMarkupMsg); err != nil {
		return err
	}
	return nil
}
