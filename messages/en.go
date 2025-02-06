package messages

type _BotMessagesEn struct{}

func (b _BotMessagesEn) ClearHistoryMessage() string {
	return "History cleared"
}

func (b _BotMessagesEn) BanUserMessage(username string) string {
	return "User " + username + " banned"
}

func (b _BotMessagesEn) BanUserPollTitle(username string) string {
	return "Ban " + username + "?"
}

func (b _BotMessagesEn) BanUserPollAgree() string {
	return "Agree"
}

func (b _BotMessagesEn) BanUserPollDisagree() string {
	return "Disagree"
}

func (b _BotMessagesEn) ModelNotFound(modelName string) string {
	return "Model " + modelName + " not found"
}

func (b _BotMessagesEn) SetModelSuccess() string {
	return "Model set successfully"
}

func (b _BotMessagesEn) JoinGroupQuestionPrefix(username string) string {
	return username + " Please answer the following question: "
}

func (b _BotMessagesEn) UserFailedToAnswerQuestion(username string) string {
	return username + " failed to join the group"
}

func (b _BotMessagesEn) UserAnswerQuestionTimeout(username string) string {
	return username + " failed to answer the question in time"
}
