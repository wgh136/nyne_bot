package services

import (
	"context"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
)

func HandleUpdate(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	if update.Message == nil && update.PollAnswer == nil {
		return
	}
	if update.Message != nil {
		log.Printf("ID:%d;ChatID:%d;Username:%s;Content:%s", update.Message.From.ID, update.Message.Chat.ID, getUserName(update.Message.From), update.Message.Text)
	} else {
		HandleGroupMessage(ctx, bot, update)
		return
	}
	if update.Message.Chat.Type == models.ChatTypeGroup || update.Message.Chat.Type == models.ChatTypeSupergroup {
		HandleGroupMessage(ctx, bot, update)
	} else if update.Message.Chat.Type == models.ChatTypePrivate {
		HandlePrivateMessage(ctx, bot, update)
	}
}
