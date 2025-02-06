package messages

type BotMessages interface {
	ClearHistoryMessage() string

	BanUserMessage(username string) string

	BanUserPollTitle(username string) string

	BanUserPollAgree() string

	BanUserPollDisagree() string

	ModelNotFound(modelName string) string

	SetModelSuccess() string

	JoinGroupQuestionPrefix(username string) string

	UserFailedToAnswerQuestion(username string) string

	UserAnswerQuestionTimeout(username string) string
}

func GetMessages(lang string) BotMessages {
	switch lang {
	case "zh_CN":
		return _BotMessagesZhCN{}
	default:
		return _BotMessagesEn{}
	}
}
