package mesh

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	ID, Topic string
	Payload   []byte
}
type Handler func(context.Context, Message) error

type RetryPolicy struct {
	Attempts            int
	BaseDelay, MaxDelay time.Duration
}

func (p RetryPolicy) Run(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < p.Attempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}
		delay := time.Duration(float64(p.BaseDelay) * math.Pow(2, float64(attempt)))
		if delay > p.MaxDelay {
			delay = p.MaxDelay
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

type Breaker struct {
	failures  atomic.Int64
	openedAt  atomic.Int64
	Threshold int64
	Cooldown  time.Duration
}

func (b *Breaker) Do(fn func() error) error {
	if opened := b.openedAt.Load(); opened > 0 && time.Since(time.Unix(0, opened)) < b.Cooldown {
		return errors.New("circuit open")
	}
	if err := fn(); err != nil {
		if b.failures.Add(1) >= b.Threshold {
			b.openedAt.Store(time.Now().UnixNano())
		}
		return err
	}
	b.failures.Store(0)
	b.openedAt.Store(0)
	return nil
}

type Deduper struct {
	mu   sync.Mutex
	ttl  time.Duration
	seen map[string]time.Time
}

func NewDeduper(ttl time.Duration) *Deduper { return &Deduper{ttl: ttl, seen: map[string]time.Time{}} }
func (d *Deduper) First(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for k, v := range d.seen {
		if now.Sub(v) > d.ttl {
			delete(d.seen, k)
		}
	}
	if _, ok := d.seen[id]; ok {
		return false
	}
	d.seen[id] = now
	return true
}

type Worker struct {
	Retry   RetryPolicy
	Breaker *Breaker
	Deduper *Deduper
	Handler Handler
}

func (w Worker) Process(ctx context.Context, msg Message) error {
	if !w.Deduper.First(msg.ID) {
		return nil
	}
	return w.Retry.Run(ctx, func() error { return w.Breaker.Do(func() error { return w.Handler(ctx, msg) }) })
}
