package services

import (
	"context"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"math/rand"
	"nyne_bot/config"
	"nyne_bot/utils"
	"strings"
	"time"
)

var (
	polls []PollStatus
)

const (
	minAgreeVotes = 3
)

func HandleGroupMessage(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	if HandleNewMember(ctx, bot, update) {
		return
	}
	if HandlePollUpdate(ctx, bot, update) {
		return
	}
	if HandleCommand(ctx, bot, update) {
		return
	}
	if CheckKeywordReplies(ctx, bot, update) {
		return
	}
	if HandleReplyToMe(ctx, bot, update) {
		return
	}
}

func HandleNewMember(ctx context.Context, bot *tgbot.Bot, update *models.Update) bool {
	if update.Message == nil || update.Message.NewChatMembers == nil {
		return false
	}
	if update.Message.Chat.Type != models.ChatTypeSupergroup {
		return true
	}
	for _, member := range update.Message.NewChatMembers {
		handleNewMember(ctx, bot, &member, update.Message.Chat.ID)
	}
	return true
}

func handleNewMember(ctx context.Context, bot *tgbot.Bot, user *models.User, chatID int64) {
	success, err := bot.RestrictChatMember(ctx, &tgbot.RestrictChatMemberParams{
		ChatID: chatID,
		UserID: user.ID,
		Permissions: &models.ChatPermissions{
			CanSendMessages: false,
		},
	})
	if !success {
		utils.LogError(err)
		return
	}
	questions := config.GetConfig().JoinGroupQuestions
	index := rand.Int() % len(questions)
	q := questions[index]
	var options []models.InputPollOption
	for _, option := range q.Options {
		options = append(options, models.InputPollOption{Text: option})
	}
	f := false
	pollMsg, err := bot.SendPoll(ctx, &tgbot.SendPollParams{
		ChatID:          chatID,
		Question:        getMessages().JoinGroupQuestionPrefix(user.Username) + q.Question,
		Options:         options,
		OpenPeriod:      120,
		Type:            "quiz",
		CorrectOptionID: q.Answer,
		IsAnonymous:     &f,
	})
	if err != nil {
		utils.LogError(err)
		return
	}
	polls = append(polls, PollStatus{
		Poll: pollMsg.Poll,
		Check: func(poll *models.Poll, uid int64, answer int) bool {
			if uid != user.ID {
				return false
			}
			if poll.CorrectOptionID == answer {
				_, err := bot.RestrictChatMember(ctx, &tgbot.RestrictChatMemberParams{
					ChatID: chatID,
					UserID: uid,
					Permissions: &models.ChatPermissions{
						CanSendMessages:       true,
						CanInviteUsers:        true,
						CanManageTopics:       true,
						CanPinMessages:        true,
						CanSendAudios:         true,
						CanSendDocuments:      true,
						CanSendPhotos:         true,
						CanSendVideos:         true,
						CanSendVideoNotes:     true,
						CanSendVoiceNotes:     true,
						CanSendPolls:          true,
						CanSendOtherMessages:  true,
						CanAddWebPagePreviews: true,
						CanChangeInfo:         true,
					},
				})
				utils.LogError(err)
			} else {
				_, err := bot.BanChatMember(ctx, &tgbot.BanChatMemberParams{
					ChatID:    chatID,
					UserID:    uid,
					UntilDate: int(time.Now().Add(5 * time.Minute).Unix()),
				})
				utils.LogError(err)
				sendMessage(SendMessageParams{
					Ctx:     ctx,
					Bot:     bot,
					ChatID:  chatID,
					Message: getMessages().UserFailedToAnswerQuestion(getUserName(user)),
				})
			}
			_, err = bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
				ChatID:    chatID,
				MessageID: pollMsg.ID,
			})
			utils.LogError(err)
			return true
		},
	})
	go func() {
		time.Sleep(2 * time.Minute)
		for i, p := range polls {
			if p.Poll.ID == pollMsg.Poll.ID {
				_, err = bot.DeleteMessage(ctx, &tgbot.DeleteMessageParams{
					ChatID:    chatID,
					MessageID: pollMsg.ID,
				})
				utils.LogError(err)
				polls = append(polls[:i], polls[i+1:]...)
				break
			}
		}
	}()
}

func HandlePollUpdate(_ context.Context, _ *tgbot.Bot, update *models.Update) bool {
	if update.PollAnswer == nil {
		return false
	}
	for i, poll := range polls {
		if poll.Poll.ID == update.PollAnswer.PollID {
			answer := update.PollAnswer.OptionIDs[0]
			poll.Poll.Options[answer].VoterCount++
			poll.Poll.TotalVoterCount++
			if poll.Check(poll.Poll, update.PollAnswer.User.ID, answer) {
				polls = append(polls[:i], polls[i+1:]...)
			}
			break
		}
	}
	return true
}

func HandleCommand(ctx context.Context, bot *tgbot.Bot, update *models.Update) bool {
	command := ParseCommand(update.Message.Text)
	if command == nil {
		return false
	}
	switch command.Name {
	case "ban":
		handleBanCommand(ctx, bot, update)
	case "set_model":
		if update.Message.From.Username != config.GetConfig().AdminUsername {
			return true
		}
		if len(command.Arguments) == 1 {
			err := SetModel(command.Arguments[0])
			if err != nil {
				sendMessage(SendMessageParams{
					Ctx:              ctx,
					Bot:              bot,
					ChatID:           update.Message.Chat.ID,
					Message:          err.Error(),
					ReplyToMessageID: update.Message.ID,
				})
			} else {
				sendMessage(SendMessageParams{
					Ctx:              ctx,
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
	case "gpt":
		message := strings.Join(command.Arguments, " ")
		if update.Message.ReplyToMessage != nil {
			message = update.Message.ReplyToMessage.Text + "\n" + message
		}
		if update.Message.Caption != "" {
			message = update.Message.Caption + "\n" + message
		}
		images := findImages(ctx, bot, update)
		content := []Content{{Type: "text", Text: message}}
		for _, image := range images {
			content = append(content, Content{Type: "image_url", ImageUrl: &ImageUrl{image}})
		}
		GptReply(bot, GptMessage{
			Role:    "user",
			Content: content,
			Name:    update.Message.From.Username,
		}, update.Message.Chat.ID, update.Message.ID, ctx)
	}

	return true
}

func handleBanCommand(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	// check if the user is admin
	member, err := bot.GetChatMember(ctx, &tgbot.GetChatMemberParams{
		ChatID: update.Message.Chat.ID,
		UserID: update.Message.From.ID,
	})
	if err != nil {
		log.Println(err)
		return
	}

	var targetMessage *models.Message

	if update.Message.ReplyToMessage != nil {
		targetMessage = update.Message.ReplyToMessage
	} else {
		return
	}
	if member.Type == models.ChatMemberTypeAdministrator || member.Type == models.ChatMemberTypeOwner {
		err := banMember(ctx, bot, *targetMessage)
		if err != nil {
			sendMessage(SendMessageParams{
				Ctx:              ctx,
				Bot:              bot,
				ChatID:           update.Message.Chat.ID,
				Message:          "Error: " + err.Error(),
				ReplyToMessageID: update.Message.ID,
			})
			utils.LogError(err)
		} else {
			sendMessage(SendMessageParams{
				Ctx:              ctx,
				Bot:              bot,
				ChatID:           update.Message.Chat.ID,
				Message:          getMessages().BanUserMessage(getUserName(targetMessage.From)),
				ReplyToMessageID: update.Message.ID,
			})
		}
	} else {
		// send a poll to vote for the ban
		f := false
		poll, err := bot.SendPoll(ctx, &tgbot.SendPollParams{
			ChatID:   update.Message.Chat.ID,
			Question: getMessages().BanUserPollTitle(getUserName(targetMessage.From)),
			Options: []models.InputPollOption{
				{Text: getMessages().BanUserPollAgree()},
				{Text: getMessages().BanUserPollDisagree()},
			},
			OpenPeriod:  600,
			IsAnonymous: &f,
		})
		if err != nil {
			utils.LogError(err)
			return
		}
		polls = append(polls, PollStatus{
			Poll: poll.Poll,
			Check: func(poll *models.Poll, uid int64, answer int) bool {
				v := poll.Options[0].VoterCount
				v -= poll.Options[1].VoterCount
				if v >= minAgreeVotes {
					err := banMember(ctx, bot, *targetMessage)
					if err != nil {
						sendMessage(SendMessageParams{
							Ctx:     ctx,
							Bot:     bot,
							ChatID:  update.Message.Chat.ID,
							Message: "Error: " + err.Error(),
						})
						utils.LogError(err)
					} else {
						sendMessage(SendMessageParams{
							Ctx:     ctx,
							Bot:     bot,
							ChatID:  update.Message.Chat.ID,
							Message: getMessages().BanUserMessage(getUserName(targetMessage.From)),
						})
					}
					return true
				}
				return false
			},
		})
		go func() {
			time.Sleep(10 * time.Minute)
			for i, p := range polls {
				if p.Poll.ID == poll.Poll.ID {
					polls = append(polls[:i], polls[i+1:]...)
					break
				}
			}
		}()
	}
}

func CheckKeywordReplies(ctx context.Context, bot *tgbot.Bot, update *models.Update) bool {
	content := update.Message.Text
	for _, kr := range config.GetConfig().KeywordReplies {
		for _, keyword := range kr.Keywords {
			if !strings.Contains(content, keyword) {
				return false
			}
		}
		sendMessage(SendMessageParams{
			Ctx:              ctx,
			ChatID:           update.Message.Chat.ID,
			Bot:              bot,
			Message:          kr.Reply,
			ReplyToMessageID: update.Message.ID,
		})
		return true
	}
	return false
}

func HandleReplyToMe(ctx context.Context, bot *tgbot.Bot, update *models.Update) bool {
	if update.Message.ReplyToMessage == nil {
		return false
	}
	if update.Message.ReplyToMessage.From.ID != bot.ID() {
		return false
	}
	username := update.Message.From.Username
	images := findImages(ctx, bot, update)
	message := update.Message.Text
	if update.Message.Caption != "" {
		message = update.Message.Caption
	}
	content := []Content{{Type: "text", Text: message}}
	for _, image := range images {
		content = append(content, Content{Type: "image_url", ImageUrl: &ImageUrl{image}})
	}
	GptReply(bot, GptMessage{
		Role:    "user",
		Content: content,
		Name:    username,
	}, update.Message.Chat.ID, update.Message.ID, ctx)
	return true
}

type PollStatus struct {
	Poll  *models.Poll
	Check func(poll *models.Poll, uid int64, answer int) bool
}

func findImages(ctx context.Context, bot *tgbot.Bot, update *models.Update) []string {
	var images []string
	if update.Message.Photo != nil {
		for _, photo := range update.Message.Photo {
			i, err := imageIdToBase64(ctx, bot, photo.FileID)
			if err != nil {
				log.Println(err)
				continue
			}
			images = append(images, i)
		}
	}
	if update.Message.ReplyToMessage != nil {
		if update.Message.ReplyToMessage.Photo != nil {
			for _, photo := range update.Message.ReplyToMessage.Photo {
				i, err := imageIdToBase64(ctx, bot, photo.FileID)
				if err != nil {
					log.Println(err)
					continue
				}
				images = append(images, i)
			}
		}
	}
	return images
}

// banMember bans a member from the chat. User who is not anonymous will be banned by their ID, otherwise by their sender chat ID.
func banMember(ctx context.Context, bot *tgbot.Bot, message models.Message) error {
	member := message.From
	chatID := message.Chat.ID
	isAnonymous := member.Username == "Channel_Bot"
	if !isAnonymous {
		_, err := bot.BanChatMember(ctx, &tgbot.BanChatMemberParams{
			ChatID: chatID,
			UserID: member.ID,
		})
		return err
	} else {
		_, err := bot.BanChatSenderChat(ctx, &tgbot.BanChatSenderChatParams{
			ChatID:       chatID,
			SenderChatID: int(message.SenderChat.ID),
		})
		return err
	}
}
