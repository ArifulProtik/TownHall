package notification

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const channelPrefix = "notifications:"

// Broker fans events out to online subscribers. Memory covers single
// instance/dev and tests; Redis covers multi-instance. Callers only see
// the interface, so scaling never changes domain code.
type Broker interface {
	Publish(ctx context.Context, e Event) error
	Subscribe(userID string) (<-chan Event, func())
}

func channelFor(userID string) string { return channelPrefix + userID }

// NewBroker picks Redis when redisURL is set, memory otherwise. Empty,
// localhost, and unreachable URLs fall back to memory so `make dev`
// works without containers.
func NewBroker(redisURL string) Broker {
	if strings.TrimSpace(redisURL) == "" {
		return NewMemoryBroker()
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr(redisURL)})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return NewMemoryBroker()
	}
	return NewRedisBroker(rdb)
}

func redisAddr(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "redis://")
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return "localhost:6379"
	}
	return s
}

type memoryBroker struct {
	mu   sync.RWMutex
	subs map[string]map[chan Event]struct{}
}

func NewMemoryBroker() Broker {
	return &memoryBroker{subs: make(map[string]map[chan Event]struct{})}
}

func (b *memoryBroker) Publish(_ context.Context, e Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[e.Recipient] {
		select {
		case ch <- e:
		default:
		}
	}
	return nil
}

func (b *memoryBroker) Subscribe(userID string) (<-chan Event, func()) {
	ch := make(chan Event, 32)
	b.mu.Lock()
	if b.subs[userID] == nil {
		b.subs[userID] = make(map[chan Event]struct{})
	}
	b.subs[userID][ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subs[userID], ch)
		close(ch)
	}
}

type redisBroker struct {
	rdb *redis.Client
}

func NewRedisBroker(rdb *redis.Client) Broker { return &redisBroker{rdb: rdb} }

func (b *redisBroker) Publish(ctx context.Context, e Event) error {
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return b.rdb.Publish(ctx, channelFor(e.Recipient), raw).Err()
}

func (b *redisBroker) Subscribe(userID string) (<-chan Event, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan Event, 32)
	sub := b.rdb.Subscribe(ctx, channelFor(userID))
	go func() {
		defer close(out)
		defer func() { _ = sub.Close() }()
		for {
			msg, err := sub.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			var e Event
			if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
				continue
			}
			select {
			case out <- e:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, cancel
}
