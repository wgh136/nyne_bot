package services

import (
	"context"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"nyne_bot/config"
	"nyne_bot/utils"
)

var (
	adminChatId int64
)

func HandlePrivateMessage(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if update.Message.From.Username == config.GetConfig().AdminUsername {
		HandleAdminMessage(ctx, bot, update)
	} else {
		HandleOtherMessage(ctx, bot, update)
	}
}

func HandleAdminMessage(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	adminChatId = update.Message.Chat.ID
	command := ParseCommand(update.Message.Text)
	if command != nil {
		switch command.Name {
		case "set_model":
			if len(command.Arguments) == 1 {
				err := SetModel(command.Arguments[0])
				if err != nil {
					sendMessage(SendMessageParams{
						Bot:              bot,
						ChatID:           update.Message.Chat.ID,
						Message:          err.Error(),
						ReplyToMessageID: update.Message.ID,
					})
				} else {
					sendMessage(SendMessageParams{
						Bot:              bot,
						ChatID:           update.Message.Chat.ID,
						Message:          getMessages().SetModelSuccess(),
						ReplyToMessageID: update.Message.ID,
					})
				}
			}
		case "reset_history":
			ClearHistory()
			sendMessage(SendMessageParams{
				Ctx:              ctx,
				Bot:              bot,
				ChatID:           update.Message.Chat.ID,
				Message:          getMessages().ClearHistoryMessage(),
				ReplyToMessageID: update.Message.ID,
			})
		}
		return
	}
	images := findImages(ctx, bot, update)
	message := update.Message.Text
	if update.Message.Caption != "" {
		message = update.Message.Caption
	}
	var content []Content
	if message != "" {
		content = append(content, Content{Type: "text", Text: message})
	}
	for _, image := range images {
		content = append(content, Content{Type: "image_url", ImageUrl: &ImageUrl{image}})
	}
	if len(content) == 0 {
		return
	}
	GptReply(bot, GptMessage{
		Role:    "user",
		Content: content,
		Name:    update.Message.From.Username,
	}, update.Message.Chat.ID, update.Message.ID, ctx)
}

func HandleOtherMessage(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	if adminChatId == 0 {
		return
	}
	_, err := bot.ForwardMessage(ctx, &tgbot.ForwardMessageParams{
		ChatID:     adminChatId,
		FromChatID: update.Message.Chat.ID,
		MessageID:  update.Message.ID,
	})
	utils.LogError(err)
}
