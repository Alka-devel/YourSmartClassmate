package main

import (
	"context"
	"sync"

	maxigobot "github.com/maxigo-bot/maxigo-bot"
)

type waiterStore struct {
	mu      sync.Mutex
	pending map[int64]chan maxigobot.Context
}

var waiters = &waiterStore{pending: make(map[int64]chan maxigobot.Context)}

func (w *waiterStore) Middleware(next maxigobot.HandlerFunc) maxigobot.HandlerFunc {
	return func(c maxigobot.Context) error {
		chatID := c.Chat()
		w.mu.Lock()
		ch, ok := w.pending[chatID]
		if ok {
			delete(w.pending, chatID)
		}
		w.mu.Unlock()
		if ok {
			select {
			case ch <- c:
			default:
			}
			return nil
		}
		return next(c)
	}
}

func (w *waiterStore) Wait(ctx context.Context, chatID int64) (maxigobot.Context, error) {
	ch := make(chan maxigobot.Context, 1)

	w.mu.Lock()
	w.pending[chatID] = ch
	w.mu.Unlock()

	select {
	case c := <-ch:
		return c, nil
	case <-ctx.Done():
		w.mu.Lock()
		delete(w.pending, chatID)
		w.mu.Unlock()
		return nil, ctx.Err()
	}
}
