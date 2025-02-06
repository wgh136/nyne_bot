package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"nyne_bot/utils"
	"os"
	"path/filepath"
)

type Config struct {
	Token              string              `yaml:"token"`
	KeywordReplies     []KeywordReply      `yaml:"keyword_replies"`
	GptConfig          GptConfig           `yaml:"gpt_config"`
	JoinGroupQuestions []JoinGroupQuestion `yaml:"join_group_questions"`
	Language           string              `yaml:"language"`
	AdminUsername      string              `yaml:"admin_username"`
}

type KeywordReply struct {
	Keywords []string `yaml:"keywords"`
	Reply    string   `yaml:"reply"`
}

type GptModel struct {
	Name     string `yaml:"name"`
	ApiUrl   string `yaml:"api_url"`
	ApiToken string `yaml:"api_token"`
}

type GptConfig struct {
	Models     []GptModel `yaml:"models"`
	SystemMsg  string     `yaml:"system_msg"`
	MaxHistory int        `yaml:"max_history"`
	MaxTokens  int        `yaml:"max_tokens"`
}

type JoinGroupQuestion struct {
	Question string   `yaml:"question"`
	Options  []string `yaml:"options"`
	Answer   int      `yaml:"answer"`
}

var config Config

func ReadConfig() {
	defer checkConfig()
	homeDir, err := os.UserHomeDir()
	utils.PanicIfError(err)
	configFile := filepath.Join(homeDir, "nyne_bot_config.yaml")
	file, err := os.Open(configFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(configFile)
		utils.PanicIfError(err)
		err = yaml.NewEncoder(file).Encode(Config{})
		utils.PanicIfError(err)
		return
	}
	utils.PanicIfError(err)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	data, err := os.ReadFile(configFile)
	utils.PanicIfError(err)
	err = yaml.Unmarshal(data, &config)
	utils.PanicIfError(err)
}

func checkConfig() {
	if config.Token == "" {
		homeDir, err := os.UserHomeDir()
		utils.PanicIfError(err)
		configFile := filepath.Join(homeDir, "nyne_bot_config.yaml")
		panic("Token is empty.\nPlease provide a token in " + configFile)
	}
}

func GetConfig() Config {
	return config
}
