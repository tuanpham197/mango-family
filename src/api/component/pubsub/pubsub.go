// Package pubsub — local pub/sub theo mẫu learn_go: mutation publish event,
// subscriber (component/subscriber) đẩy sang WebSocket hub.
package pubsub

import (
	"sync"

	"github.com/google/uuid"
)

type Message struct {
	Topic       string    // common.Topic*Changed
	HouseholdID uuid.UUID // hub broadcast theo hộ
}

type PubSub interface {
	Publish(msg Message)
	Subscribe() <-chan Message
}

type localPubSub struct {
	mu   sync.RWMutex
	subs []chan Message
}

func NewLocal() PubSub { return &localPubSub{} }

func (p *localPubSub) Publish(msg Message) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, ch := range p.subs {
		select {
		case ch <- msg:
		default: // subscriber chậm thì bỏ qua — FE có fallback refetch khi focus
		}
	}
}

func (p *localPubSub) Subscribe() <-chan Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	ch := make(chan Message, 64)
	p.subs = append(p.subs, ch)
	return ch
}
