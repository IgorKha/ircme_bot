package bot

import (
	"context"

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
