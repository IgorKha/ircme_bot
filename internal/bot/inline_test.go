package bot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/mymmrac/telego"
)

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

	if textContent.MessageText != "<i>@nick hello world</i>" {
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
