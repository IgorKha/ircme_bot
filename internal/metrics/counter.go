package metrics

import "sync/atomic"

type Counter struct {
	value atomic.Uint64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() uint64 {
	return c.value.Add(1)
}

func (c *Counter) Value() uint64 {
	return c.value.Load()
}
