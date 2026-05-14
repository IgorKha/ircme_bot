package bot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

type updateHandler interface {
	Handle(ctx context.Context, update telego.Update) error
}

type Dispatcher struct {
	handlers []updateHandler
}

func NewDispatcher(handlers ...updateHandler) *Dispatcher {
	return &Dispatcher{handlers: handlers}
}

func (d *Dispatcher) Handle(ctx context.Context, update telego.Update) error {
	for _, handler := range d.handlers {
		if err := handler.Handle(ctx, update); err != nil {
			return fmt.Errorf("dispatch to %T: %w", handler, err)
		}
	}

	return nil
}
