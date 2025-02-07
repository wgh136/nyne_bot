package services

import (
	"context"
	"encoding/base64"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"io"
	"log"
	"net/http"
	"nyne_bot/config"
	"nyne_bot/messages"
	"nyne_bot/utils"
)

type SendMessageParams struct {
	Ctx              context.Context
	ChatID           int64
	Bot              *tgbot.Bot
	Message          string
	ReplyToMessageID int
	ParseMode        models.ParseMode
}

func sendMessage(params SendMessageParams) {
	p := tgbot.SendMessageParams{
		ChatID:    params.ChatID,
		Text:      params.Message,
		ParseMode: params.ParseMode,
	}
	if params.ReplyToMessageID != 0 {
		p.ReplyParameters = &models.ReplyParameters{
			MessageID: params.ReplyToMessageID,
			ChatID:    params.ChatID,
		}
	}
	_, err := params.Bot.SendMessage(params.Ctx, &p)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

func getMessages() messages.BotMessages {
	return messages.GetMessages(config.GetConfig().Language)
}

func imageIdToBase64(ctx context.Context, bot *tgbot.Bot, imageId string) (string, error) {
	file, err := bot.GetFile(ctx, &tgbot.GetFileParams{
		FileID: imageId,
	})
	if err != nil {
		return "", err
	}
	fileUrl := "https://api.telegram.org/file/bot" + bot.Token() + "/" + file.FilePath
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}
	defer utils.CloseAndLogError(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	mime := http.DetectContentType(data)
	encoder := base64.StdEncoding
	encoded := make([]byte, encoder.EncodedLen(len(data)))
	encoder.Encode(encoded, data)
	return "data:" + mime + ";base64," + string(encoded), nil
}

func getUserName(user *models.User) string {
	if user.Username != "" {
		return "@" + user.Username
	} else {
		splitter := " "
		if user.FirstName == "" || user.LastName == "" {
			splitter = ""
		}
		return user.FirstName + splitter + user.LastName
	}
}

func getUserNameFromMessage(message *models.Message) string {
	if message.SenderChat != nil {
		return "@" + message.SenderChat.Username
	} else {
		return getUserName(message.From)
	}
}
