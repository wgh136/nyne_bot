package messages

type _BotMessagesZhCN struct{}

func (b _BotMessagesZhCN) ClearHistoryMessage() string {
	return "已清除历史记录"
}

func (b _BotMessagesZhCN) BanUserMessage(username string) string {
	return "用户 " + username + " 已被封禁"
}

func (b _BotMessagesZhCN) BanUserPollTitle(username string) string {
	return "是否封禁用户 " + username + "? " + "需要 同意 - 反对 >= 5"
}

func (b _BotMessagesZhCN) BanUserPollAgree() string {
	return "同意"
}

func (b _BotMessagesZhCN) BanUserPollDisagree() string {
	return "反对"
}

func (b _BotMessagesZhCN) ModelNotFound(modelName string) string {
	return "模型 " + modelName + " 不存在"
}

func (b _BotMessagesZhCN) SetModelSuccess() string {
	return "成功设置模型"
}

func (b _BotMessagesZhCN) JoinGroupQuestionPrefix(username string) string {
	return username + " 请回答此问题: "
}

func (b _BotMessagesZhCN) UserFailedToAnswerQuestion(username string) string {
	return username + " 回答问题错误, 加入群组失败"
}

func (b _BotMessagesZhCN) UserAnswerQuestionTimeout(username string) string {
	return username + " 没有及时回答问题"
}
