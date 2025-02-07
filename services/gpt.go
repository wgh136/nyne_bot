package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"nyne_bot/config"
	"nyne_bot/messages"
	"nyne_bot/utils"
)

type GptMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
	Name    string    `json:"name,omitempty"`
}

type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageUrl *ImageUrl `json:"image_url,omitempty"`
}

type ImageUrl struct {
	URL string `json:"url"`
}

type GptRequest struct {
	Model       string       `json:"model"`
	Messages    []GptMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_completion_tokens"`
}

var (
	chatHistory        []GptMessage
	isInit             bool
	model              config.GptModel
	waitingForGptReply bool
)

func InitGpt() {
	ClearHistory()
	if len(config.GetConfig().GptConfig.Models) == 0 {
		panic("No GPT model found")
	}
	model = config.GetConfig().GptConfig.Models[0]
}

// GptReply replies to the user's message using current GPT model. Messages will be stored in chatHistory.
// If the length of chatHistory exceeds the maximum history length, it will be cleared.
func GptReply(bot *tgbot.Bot, newMessage GptMessage, chatID int64, replyToMessageID int, ctx context.Context) {
	if !isInit {
		InitGpt()
		isInit = true
	}
	chatHistory = append(chatHistory, newMessage)
	err := fetchGptMessage()
	if err != nil {
		sendMessage(SendMessageParams{
			Ctx:              ctx,
			Bot:              bot,
			ChatID:           chatID,
			Message:          "Error: " + err.Error(),
			ReplyToMessageID: replyToMessageID,
		})
	} else {
		sendMessage(SendMessageParams{
			Ctx:              ctx,
			Bot:              bot,
			ChatID:           chatID,
			Message:          RenderMarkdown(chatHistory[len(chatHistory)-1].Content[0].Text),
			ReplyToMessageID: replyToMessageID,
			ParseMode:        models.ParseModeMarkdown,
		})
	}
	checkHistory(bot, chatID)
}

func fetchGptMessage() error {
	if waitingForGptReply {
		return errors.New("GPT is still replying")
	}
	waitingForGptReply = true
	defer func() {
		waitingForGptReply = false
	}()
	req := GptRequest{
		Model:       model.Name,
		Messages:    chatHistory,
		Temperature: 0.7,
		MaxTokens:   config.GetConfig().GptConfig.MaxTokens,
	}
	reqData, err := json.Marshal(req)
	log.Println(string(reqData))
	if err != nil {
		return err
	}
	reqReader := bytes.NewReader(reqData)
	client := http.Client{}
	u, _ := url.Parse(model.ApiUrl)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+model.ApiToken)
	headers.Set("Content-Type", "application/json")
	httpReq := http.Request{
		Method: http.MethodPost,
		URL:    u,
		Header: headers,
		Body:   io.NopCloser(reqReader),
	}
	res, err := client.Do(&httpReq)
	if err != nil {
		return err
	}
	defer utils.CloseAndLogError(res.Body)
	data, err := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return errors.New("Invalid status code " + res.Status + ": " + string(data))
	}
	var gptRes map[string]interface{}
	err = json.Unmarshal(data, &gptRes)
	if err != nil {
		return err
	}
	content := gptRes["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	chatHistory = append(chatHistory, GptMessage{
		Role:    "assistant",
		Content: []Content{{Type: "text", Text: content}},
	})
	return nil
}

func checkHistory(bot *tgbot.Bot, chatID int64) {
	length := 0
	for _, message := range chatHistory {
		if message.Role == "assistant" {
			length++
		}
	}
	if length > config.GetConfig().GptConfig.MaxHistory {
		ClearHistory()
		sendMessage(SendMessageParams{
			Bot:              bot,
			ChatID:           chatID,
			Message:          messages.GetMessages(config.GetConfig().Language).ClearHistoryMessage(),
			ReplyToMessageID: 0,
		})
	}
}

func ClearHistory() {
	chatHistory = make([]GptMessage, 0)
	chatHistory = append(chatHistory, GptMessage{
		Role:    "system",
		Content: []Content{{Type: "text", Text: config.GetConfig().GptConfig.SystemMsg}},
	})
}

func SetModel(name string) error {
	for _, m := range config.GetConfig().GptConfig.Models {
		if m.Name == name {
			model = m
			return nil
		}
	}
	return errors.New(getMessages().ModelNotFound(name))
}
