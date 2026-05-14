package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/igorkha/ircme_bot/internal/bot"
	"github.com/igorkha/ircme_bot/internal/metrics"
	"github.com/mymmrac/telego"
)

const (
	tokenEnvKey      = "TELEGRAM_BOT_TOKEN"
	longPollingTime  = 30
	minWorkerThreads = 8
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("bot stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	token := strings.TrimSpace(os.Getenv(tokenEnvKey))
	if token == "" {
		return errors.New("TELEGRAM_BOT_TOKEN is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tgBot, err := telego.NewBot(token)
	if err != nil {
		return err
	}

	updates, err := tgBot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Timeout:        longPollingTime,
		AllowedUpdates: []string{"inline_query", "chosen_inline_result", "message"},
	})
	if err != nil {
		return err
	}

	counter := metrics.NewCounter()
	inlineHandler := bot.NewInlineHandler(tgBot, counter, logger)
	commandHandler := bot.NewCommandHandler(tgBot, logger)
	dispatcher := bot.NewDispatcher(inlineHandler, commandHandler)
	workers := maxWorkerCount(runtime.NumCPU() * 4)
	logger.Info("bot started", "workers", workers, "mode", "inline+commands", "architecture", "dispatcher")

	jobs := make(chan telego.Update, workers*8)
	var wg sync.WaitGroup

	for workerID := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for update := range jobs {
				if err := dispatcher.Handle(ctx, update); err != nil {
					logger.Error("update handling failed", "worker", id, "error", err)
				}
			}
		}(workerID)
	}

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			logger.Info("bot stopped")
			return nil
		case update, ok := <-updates:
			if !ok {
				close(jobs)
				wg.Wait()
				logger.Info("updates channel closed")
				return nil
			}

			select {
			case jobs <- update:
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				logger.Info("bot stopped")
				return nil
			}
		}
	}
}

func maxWorkerCount(calculated int) int {
	if calculated < minWorkerThreads {
		return minWorkerThreads
	}
	return calculated
}
