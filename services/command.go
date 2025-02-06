package services

import (
	"regexp"
	"strings"
)

type Command struct {
	Name      string
	Arguments []string
}

func isCommand(content string) bool {
	splits := strings.Split(content, " ")
	first := splits[0]
	if strings.Contains(first, "@") {
		first = strings.Split(first, "@")[0]
	}
	re := regexp.MustCompile(`^/[a-z][a-z0-9_]*$`)
	return re.MatchString(first)
}

func ParseCommand(text string) *Command {
	if !isCommand(text) {
		return nil
	}
	splits := strings.Split(text, " ")
	name := strings.TrimPrefix(splits[0], "/")
	if strings.Contains(name, "@") {
		name = strings.Split(name, "@")[0]
	}
	args := splits[1:]
	return &Command{
		Name:      name,
		Arguments: args,
	}
}
