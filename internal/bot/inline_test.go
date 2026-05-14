package bot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/mymmrac/telego"
)

type mockBot struct {
	answerInlineQueryParams *telego.AnswerInlineQueryParams
	answerInlineQueryErr    error
	answerInlineQueryCalled bool

	sendMessageParams *telego.SendMessageParams
	sendMessageErr    error
	sendMessageCalled bool

	deleteMessageParams *telego.DeleteMessageParams
	deleteMessageErr    error
	deleteMessageCalled bool
}

func (m *mockBot) AnswerInlineQuery(_ context.Context, params *telego.AnswerInlineQueryParams) error {
	m.answerInlineQueryCalled = true
	m.answerInlineQueryParams = params
	return m.answerInlineQueryErr
}

func (m *mockBot) SendMessage(_ context.Context, params *telego.SendMessageParams) (*telego.Message, error) {
	m.sendMessageCalled = true
	m.sendMessageParams = params
	if m.sendMessageErr != nil {
		return nil, m.sendMessageErr
	}

	return &telego.Message{}, nil
}

func (m *mockBot) DeleteMessage(_ context.Context, params *telego.DeleteMessageParams) error {
	m.deleteMessageCalled = true
	m.deleteMessageParams = params
	return m.deleteMessageErr
}

type mockCounter struct {
	value uint64
}

func (m *mockCounter) Inc() uint64 {
	m.value++
	return m.value
}

func TestResolveDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user telego.User
		want string
	}{
		{
			name: "username has priority",
			user: telego.User{
				Username:  "nick",
				FirstName: "John",
			},
			want: "nick",
		},
		{
			name: "fallback to first name",
			user: telego.User{
				FirstName: "John",
			},
			want: "John",
		},
		{
			name: "fallback to someone",
			user: telego.User{},
			want: "someone",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveDisplayName(tc.user); got != tc.want {
				t.Fatalf("resolveDisplayName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildIRCMeMessageHTML(t *testing.T) {
	t.Parallel()

	got := buildIRCMeMessageHTML(
		telego.User{
			Username: "<nick>",
		},
		"1 < 2 & 3",
	)

	want := "&lt;nick&gt; thinks that: <i>1 &lt; 2 &amp; 3</i>"
	if got != want {
		t.Fatalf("buildIRCMeMessageHTML() = %q, want %q", got, want)
	}
}

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

			got, ok := extractMeCommandText(tc.text)
			if ok != tc.ok {
				t.Fatalf("extractMeCommandText() ok = %v, want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("extractMeCommandText() text = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHandleInlineQuerySuccess(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	counter := &mockCounter{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewInlineHandler(mockBot, counter, logger)

	update := telego.Update{
		InlineQuery: &telego.InlineQuery{
			ID:    "query-id",
			Query: "hello world",
			From: telego.User{
				Username: "nick",
			},
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if !mockBot.answerInlineQueryCalled {
		t.Fatalf("AnswerInlineQuery was not called")
	}

	if counter.value != 0 {
		t.Fatalf("counter value = %d, want 0", counter.value)
	}

	if mockBot.answerInlineQueryParams == nil {
		t.Fatalf("AnswerInlineQuery params are nil")
	}

	if mockBot.answerInlineQueryParams.InlineQueryID != "query-id" {
		t.Fatalf("InlineQueryID = %q, want %q", mockBot.answerInlineQueryParams.InlineQueryID, "query-id")
	}

	if len(mockBot.answerInlineQueryParams.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(mockBot.answerInlineQueryParams.Results))
	}

	article, ok := mockBot.answerInlineQueryParams.Results[0].(*telego.InlineQueryResultArticle)
	if !ok {
		t.Fatalf("result has unexpected type %T", mockBot.answerInlineQueryParams.Results[0])
	}

	textContent, ok := article.InputMessageContent.(*telego.InputTextMessageContent)
	if !ok {
		t.Fatalf("input message content has unexpected type %T", article.InputMessageContent)
	}

	if textContent.ParseMode != telego.ModeHTML {
		t.Fatalf("parse mode = %q, want %q", textContent.ParseMode, telego.ModeHTML)
	}

	if textContent.MessageText != "nick thinks that: <i>hello world</i>" {
		t.Fatalf("message text = %q", textContent.MessageText)
	}
}

func TestHandleChosenInlineResultIncrementsCounter(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	counter := &mockCounter{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewInlineHandler(mockBot, counter, logger)

	update := telego.Update{
		ChosenInlineResult: &telego.ChosenInlineResult{
			ResultID: "result-id",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if counter.value != 1 {
		t.Fatalf("counter value = %d, want 1", counter.value)
	}

	if mockBot.answerInlineQueryCalled {
		t.Fatalf("AnswerInlineQuery should not be called for ChosenInlineResult")
	}
}

func TestHandleInlineQueryErrorDoesNotIncrementCounter(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{answerInlineQueryErr: errors.New("telegram error")}
	counter := &mockCounter{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewInlineHandler(mockBot, counter, logger)

	update := telego.Update{
		InlineQuery: &telego.InlineQuery{
			ID:    "query-id",
			Query: "hello",
			From:  telego.User{Username: "nick"},
		},
	}

	if err := handler.Handle(context.Background(), update); err == nil {
		t.Fatalf("Handle() error = nil, want non-nil")
	}

	if counter.value != 0 {
		t.Fatalf("counter value = %d, want 0", counter.value)
	}
}

func TestInlineHandlerIgnoresMessageUpdates(t *testing.T) {
	t.Parallel()

	mockBot := &mockBot{}
	counter := &mockCounter{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewInlineHandler(mockBot, counter, logger)

	update := telego.Update{
		Message: &telego.Message{
			MessageID: 1,
			Chat:      telego.Chat{ID: 100},
			Text:      "/me hello",
		},
	}

	if err := handler.Handle(context.Background(), update); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if mockBot.answerInlineQueryCalled {
		t.Fatalf("AnswerInlineQuery should not be called for Message updates")
	}
	if counter.value != 0 {
		t.Fatalf("counter value = %d, want 0", counter.value)
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
	if mockBot.sendMessageParams.Text != "nick thinks that: <i>hello world</i>" {
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
	if mockBot.sendMessageParams.Text != "John thinks that: <i>...</i>" {
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
