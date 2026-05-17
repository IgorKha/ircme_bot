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

	message := update.Message

	if commandText, ok := extractCommandText(message.Text, "me"); ok {
		h.logger.Debug("handling /me command", "chat_id", message.Chat.ID, "message_id", message.MessageID)
		return h.executeCommand(ctx, message, buildIRCMeMessageHTML(resolveMessageSender(message), commandText))
	}

	if commandText, ok := extractCommandText(message.Text, "ame"); ok {
		h.logger.Debug("handling /ame command", "chat_id", message.Chat.ID, "message_id", message.MessageID)
		return h.executeCommand(ctx, message, buildIRCAnonymousMessageHTML(commandText))
	}

	if commandText, ok := extractCommandText(message.Text, "slap"); ok {
		h.logger.Debug("handling /slap command", "chat_id", message.Chat.ID, "message_id", message.MessageID)
		return h.executeCommand(ctx, message, buildIRCSlapMessageHTML(resolveMessageSender(message), resolveSlapTarget(commandText, message)))
	}

	return nil
}

func (h *CommandHandler) executeCommand(ctx context.Context, message *telego.Message, response string) error {
	chatID := telego.ChatID{ID: message.Chat.ID}
	sendParams := tu.Message(chatID, response).WithParseMode(telego.ModeHTML)
	if message.ReplyToMessage != nil {
		sendParams = sendParams.WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.ReplyToMessage.MessageID,
		})
	}
	if _, err := h.bot.SendMessage(ctx, sendParams); err != nil {
		return fmt.Errorf("send response in chat %d: %w", message.Chat.ID, err)
	}

	if err := h.bot.DeleteMessage(ctx, tu.Delete(chatID, message.MessageID)); err != nil {
		return fmt.Errorf("delete command message %d in chat %d: %w", message.MessageID, message.Chat.ID, err)
	}

	return nil
}

func extractCommandText(text, command string) (string, bool) {
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

	cmd := strings.TrimPrefix(parts[0], "/")
	cmd = strings.SplitN(cmd, "@", 2)[0]
	if cmd != command {
		return "", false
	}

	return strings.Join(parts[1:], " "), true
}
