package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type inlineQueryAnswerer interface {
	AnswerInlineQuery(ctx context.Context, params *telego.AnswerInlineQueryParams) error
}

type usageCounter interface {
	Inc() uint64
}

type InlineHandler struct {
	bot     inlineQueryAnswerer
	counter usageCounter
	logger  *slog.Logger
}

func NewInlineHandler(bot inlineQueryAnswerer, counter usageCounter, logger *slog.Logger) *InlineHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &InlineHandler{
		bot:     bot,
		counter: counter,
		logger:  logger,
	}
}

func (h *InlineHandler) Handle(ctx context.Context, update telego.Update) error {
	if update.InlineQuery != nil {
		return h.handleInlineQuery(ctx, update.InlineQuery)
	}

	if update.ChosenInlineResult != nil {
		count := h.counter.Inc()
		h.logger.Info("usage counter incremented", "count", count)
		return nil
	}

	return nil
}

func (h *InlineHandler) handleInlineQuery(ctx context.Context, query *telego.InlineQuery) error {
	response := buildIRCMeMessage(query.From, query.Query)
	formattedResponse := buildIRCMeMessageHTML(query.From, query.Query)
	result := tu.ResultArticle("ircme-"+query.ID, "Send IRC /me style message", tu.TextMessage(formattedResponse).WithParseMode(telego.ModeHTML)).
		WithDescription(response)

	params := tu.InlineQuery(query.ID, result).WithCacheTime(0).WithIsPersonal()
	if err := h.bot.AnswerInlineQuery(ctx, params); err != nil {
		return fmt.Errorf("answer inline query %s: %w", query.ID, err)
	}

	return nil
}
