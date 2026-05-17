package bot

import (
	"fmt"
	"html"
	"strings"

	"github.com/mymmrac/telego"
)

func buildIRCMeMessage(user telego.User, query string) string {
	return fmt.Sprintf("%s %s", resolveDisplayName(user), normalizeIRCMeText(query))
}

func buildIRCMeMessageHTML(user telego.User, query string) string {
	name := html.EscapeString(resolveDisplayName(user))
	text := html.EscapeString(normalizeIRCMeText(query))
	return fmt.Sprintf("<i>@%s %s</i>", name, text)
}

func buildIRCAnonymousMessageHTML(query string) string {
	text := html.EscapeString(normalizeIRCMeText(query))
	return fmt.Sprintf("<i>%s</i>", text)
}

func resolveDisplayName(user telego.User) string {
	username := strings.TrimSpace(user.Username)
	if username != "" {
		return username
	}

	firstName := strings.TrimSpace(user.FirstName)
	if firstName != "" {
		return firstName
	}

	return "someone"
}

func resolveMessageSender(message *telego.Message) telego.User {
	if message.From != nil {
		return *message.From
	}

	return telego.User{}
}

func normalizeIRCMeText(query string) string {
	text := strings.TrimSpace(query)
	if text == "" {
		return "..."
	}
	return text
}
