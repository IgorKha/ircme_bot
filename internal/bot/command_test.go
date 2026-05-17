package bot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/mymmrac/telego"
)

func TestExtractMeCommandText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
		ok   bool
	}{
		{
			name: "basic command",
			text: "/me hello world",
			want: "hello world",
			ok:   true,
		},
		{
			name: "command with bot username",
			text: "/me@ircme_bot hello world",
			want: "hello world",
			ok:   true,
		},
		{
			name: "command without payload",
			text: "/me",
			want: "",
			ok:   true,
		},
		{
			name: "unsupported command",
			text: "/start hi",
			want: "",
			ok:   false,
		},
		{
			name: "plain text is ignored",
			text: "me hello",
			want: "",
			ok:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := extractCommandText(tc.text, "me")
			if ok != tc.ok {
				t.Fatalf("extractCommandText() ok = %v, want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("extractCommandText() text = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHandleMessageMeCommandSuccess(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 55,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "/me hello world",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !mockBot.sendMessageCalled {
		t.Fatalf("SendMessage was not called")
	}
	if mockBot.sendMessageParams == nil {
		t.Fatalf("SendMessage params are nil")
	}
	if mockBot.sendMessageParams.ChatID.ID != -100123 {
		t.Fatalf("SendMessage chat id = %d, want %d", mockBot.sendMessageParams.ChatID.ID, -100123)
	}
	if mockBot.sendMessageParams.Text != "<i>@nick hello world</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
	if mockBot.sendMessageParams.ParseMode != telego.ModeHTML {
		t.Fatalf("SendMessage parse mode = %q, want %q", mockBot.sendMessageParams.ParseMode, telego.ModeHTML)
	}

	if !mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage was not called")
	}
	if mockBot.deleteMessageParams == nil {
		t.Fatalf("DeleteMessage params are nil")
	}
	if mockBot.deleteMessageParams.ChatID.ID != -100123 {
		t.Fatalf("DeleteMessage chat id = %d, want %d", mockBot.deleteMessageParams.ChatID.ID, -100123)
	}
	if mockBot.deleteMessageParams.MessageID != 55 {
		t.Fatalf("DeleteMessage message id = %d, want %d", mockBot.deleteMessageParams.MessageID, 55)
	}
}

func TestHandleMessageWithoutMeCommandDoesNothing(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 55,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "hello world",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageCalled {
		t.Fatalf("SendMessage should not be called")
	}
	if mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage should not be called")
	}
}

func TestHandleMessageMeCommandWithoutTextUsesDefault(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 78,
			From:      &telego.User{FirstName: "John"},
			Chat:      telego.Chat{ID: 12345},
			Text:      "/me",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageParams == nil {
		t.Fatalf("SendMessage params are nil")
	}
	if mockBot.sendMessageParams.Text != "<i>@John ...</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
	if mockBot.sendMessageParams.ParseMode != telego.ModeHTML {
		t.Fatalf("SendMessage parse mode = %q, want %q", mockBot.sendMessageParams.ParseMode, telego.ModeHTML)
	}
}

func TestHandleMessageMeCommandSendError(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{sendMessageErr: errors.New("telegram send error")}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 1,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: 101},
			Text:      "/me hi",
		},
	}

	if err := handler.Handle(context.Background(), update); err == nil {
		t.Fatalf("Handle() error = nil, want non-nil")
	}

	if mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage should not be called when SendMessage fails")
	}
}

func TestHandleMessageMeCommandDeleteError(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{deleteMessageErr: errors.New("telegram delete error")}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 2,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: 202},
			Text:      "/me hi",
		},
	}

	if err := handler.Handle(context.Background(), update); err == nil {
		t.Fatalf("Handle() error = nil, want non-nil")
	}

	if !mockBot.sendMessageCalled {
		t.Fatalf("SendMessage should be called before DeleteMessage")
	}
	if !mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage should be called")
	}
}

func TestCommandHandlerIgnoresInlineUpdates(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		InlineQuery: &telego.InlineQuery{
			ID:    "query-id",
			Query: "hello",
			From:  telego.User{Username: "nick"},
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageCalled {
		t.Fatalf("SendMessage should not be called for InlineQuery updates")
	}
	if mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage should not be called for InlineQuery updates")
	}
}

func TestHandleSlapCommandWithTarget(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 99,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "/slap bob",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !mockBot.sendMessageCalled {
		t.Fatalf("SendMessage was not called")
	}
	if mockBot.sendMessageParams.Text != "<i>* @nick slaps @bob around a bit with a large trout</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
	if mockBot.sendMessageParams.ParseMode != telego.ModeHTML {
		t.Fatalf("SendMessage parse mode = %q, want %q", mockBot.sendMessageParams.ParseMode, telego.ModeHTML)
	}
	if !mockBot.deleteMessageCalled {
		t.Fatalf("DeleteMessage was not called")
	}
	if mockBot.deleteMessageParams.MessageID != 99 {
		t.Fatalf("DeleteMessage message id = %d, want %d", mockBot.deleteMessageParams.MessageID, 99)
	}
}

func TestHandleSlapCommandWithoutTargetSelfSlap(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 100,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "/slap",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageParams.Text != "<i>* @nick slaps @nick around a bit with a large trout</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
}

func TestHandleSlapCommandWithAtPrefixTarget(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 101,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "/slap @bob",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageParams.Text != "<i>* @nick slaps @bob around a bit with a large trout</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
}

func TestHandleSlapCommandWithReplyTarget(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewCommandHandler(mockBot, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 102,
			From:      &telego.User{Username: "nick"},
			Chat:      telego.Chat{ID: -100123},
			Text:      "/slap",
			ReplyToMessage: &telego.Message{
				MessageID: 50,
				From:      &telego.User{Username: "alice"},
			},
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.sendMessageParams.Text != "<i>* @nick slaps @alice around a bit with a large trout</i>" {
		t.Fatalf("SendMessage text = %q", mockBot.sendMessageParams.Text)
	}
}
