package services

import (
	tgbot "github.com/go-telegram/bot"
	"regexp"
	"strings"
)

func RenderMarkdown(text string) string {
	builder := strings.Builder{}
	isCode := false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "```") {
			builder.WriteString(line)
			isCode = !isCode
			continue
		}
		if isCode {
			builder.WriteString(tgbot.EscapeMarkdownUnescaped(renderCodeLine(line)))
			builder.WriteRune('\n')
			continue
		}
		if line != "" {
			rendered := renderTitle(line)
			if rendered == "" {
				rendered = renderNormal(line)
			}
			builder.WriteString(rendered)
		}
		builder.WriteRune('\n')
	}
	return builder.String()
}

func renderTitle(line string) string {
	re := regexp.MustCompile(`^(#{1,6})\s(.*)$`)
	isTitle := re.MatchString(line)
	if !isTitle {
		return ""
	} else {
		content := line[strings.Index(line, " ")+1:]
		content = tgbot.EscapeMarkdown(content)
		return "*" + content + "*"
	}
}

func renderNormal(line string) string {
	if strings.HasPrefix(line, ">") {
		return ">" + tgbot.EscapeMarkdown(line[1:])
	}

	builder := strings.Builder{}
	cache := strings.Builder{}
	isEscaped := false
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c == '\\' {
			isEscaped = !isEscaped
			continue
		}
		if c == '*' && !isEscaped {
			next := findNextSymbol(line, '*', i+1)
			if next == -1 {
				cache.WriteString("\\*")
			} else {
				writeCache(&builder, &cache)
				content := tgbot.EscapeMarkdownUnescaped(string(runes[i+1 : next]))
				builder.WriteString("*" + content + "*")
				i = next
			}
		} else if c == '_' && !isEscaped {
			next := findNextSymbol(line, '_', i+1)
			if next == -1 {
				cache.WriteString("\\_")
			} else {
				writeCache(&builder, &cache)
				content := tgbot.EscapeMarkdownUnescaped(string(runes[i+1 : next]))
				builder.WriteString("_" + content + "_")
				i = next
			}
		} else if c == '~' && !isEscaped {
			if (i + 1) < len(runes) {
				nextChar := runes[i+1]
				if nextChar == ' ' {
					// Do not render strikethrough if there is a space after ~
					// ~ could be a punctuation
					cache.WriteString("\\~")
					continue
				}
			}
			next := findNextSymbol(line, '~', i+1)
			if next == -1 {
				cache.WriteString("\\~")
			} else {
				writeCache(&builder, &cache)
				content := tgbot.EscapeMarkdownUnescaped(string(runes[i+1 : next]))
				builder.WriteString("~" + content + "~")
				i = next
			}
		} else if c == '`' && !isEscaped {
			next := findNextSymbol(line, '`', i+1)
			if next == -1 {
				cache.WriteString("\\`")
			} else {
				writeCache(&builder, &cache)
				content := renderCodeLine(string(runes[i+1 : next]))
				builder.WriteString("`" + content + "`")
				i = next
			}
		} else if c == '[' && !isEscaped {
			right := findNextSymbol(line, ']', i+1)
			leftRound := findNextSymbol(line, '(', right+1)
			rightRound := findNextSymbol(line, ')', leftRound+1)
			if right == -1 || leftRound == -1 || rightRound == -1 {
				cache.WriteRune(c)
			} else {
				writeCache(&builder, &cache)
				text := tgbot.EscapeMarkdownUnescaped(string(runes[i+1 : right]))
				link := tgbot.EscapeMarkdownUnescaped(string(runes[leftRound+1 : rightRound]))
				builder.WriteString("[" + text + "](" + link + ")")
				i = rightRound
			}
		} else {
			cache.WriteRune(c)
		}
	}
	if cache.Len() > 0 {
		writeCache(&builder, &cache)
	}
	return builder.String()
}

func writeCache(builder *strings.Builder, cache *strings.Builder) {
	builder.WriteString(tgbot.EscapeMarkdownUnescaped(cache.String()))
	cache.Reset()
}

func findNextSymbol(text string, symbol rune, start int) int {
	isEscaped := false
	for i := start; i < len([]rune(text)); i++ {
		c := []rune(text)[i]
		if c == '\\' {
			isEscaped = !isEscaped
			continue
		}
		if c == symbol && !isEscaped {
			return i
		}
	}
	return -1
}

func renderCodeLine(line string) string {
	line = strings.Replace(line, "\\", "\\\\", -1)
	line = strings.Replace(line, "`", "\\`", -1)
	return line
}
