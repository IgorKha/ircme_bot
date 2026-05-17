package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type commandMessageClient interface {
	SendMessage(ctx context.Context, params *telego.SendMessageParams) (*telego.Message, error)
	DeleteMessage(ctx context.Context, params *telego.DeleteMessageParams) error
}

type CommandHandler struct {
	bot    commandMessageClient
	logger *slog.Logger
}

func NewCommandHandler(bot commandMessageClient, logger *slog.Logger) *CommandHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &CommandHandler{
		bot:    bot,
		logger: logger,
	}
}

func (h *CommandHandler) Handle(ctx context.Context, update telego.Update) error {
	if update.Message == nil {
		return nil
	}

	return h.handleMeCommand(ctx, update.Message)
}

func (h *CommandHandler) handleMeCommand(ctx context.Context, message *telego.Message) error {
	commandText, ok := extractMeCommandText(message.Text)
	if !ok {
		return nil
	}

	h.logger.Debug("handling /me command", "chat_id", message.Chat.ID, "message_id", message.MessageID)

	chatID := telego.ChatID{ID: message.Chat.ID}
	response := buildIRCMeMessageHTML(resolveMessageSender(message), commandText)
	sendParams := tu.Message(chatID, response).WithParseMode(telego.ModeHTML)
	if message.ReplyToMessage != nil {
		sendParams = sendParams.WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.ReplyToMessage.MessageID,
		})
	}
	if _, err := h.bot.SendMessage(ctx, sendParams); err != nil {
		return fmt.Errorf("send /me response in chat %d: %w", message.Chat.ID, err)
	}

	if err := h.bot.DeleteMessage(ctx, tu.Delete(chatID, message.MessageID)); err != nil {
		return fmt.Errorf("delete /me command message %d in chat %d: %w", message.MessageID, message.Chat.ID, err)
	}

	return nil
}

func extractMeCommandText(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}

	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return "", false
	}

	if !strings.HasPrefix(parts[0], "/") {
		return "", false
	}

	command := strings.TrimPrefix(parts[0], "/")
	command = strings.SplitN(command, "@", 2)[0]
	if command != "me" {
		return "", false
	}

	return strings.Join(parts[1:], " "), true
}
