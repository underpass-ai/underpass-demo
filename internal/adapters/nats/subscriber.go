// Package nats implements the EventSubscriber port for NATS.
package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

const subjectPolicyUpdated = "tool_learning.policy.updated"

// Subscriber listens for policy update events on NATS.
type Subscriber struct {
	conn *nats.Conn
	sub  *nats.Subscription
}

// NewSubscriber connects to NATS and returns a subscriber.
func NewSubscriber(url string) (*Subscriber, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Subscriber{conn: conn}, nil
}

// Subscribe returns a channel that receives policy update events.
func (s *Subscriber) Subscribe(ctx context.Context) (<-chan domain.PolicyUpdateEvent, error) {
	ch := make(chan domain.PolicyUpdateEvent, 64)

	sub, err := s.conn.Subscribe(subjectPolicyUpdated, func(msg *nats.Msg) {
		var evt domain.PolicyUpdateEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			return
		}
		select {
		case ch <- evt:
		case <-ctx.Done():
		}
	})
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	s.sub = sub

	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
		close(ch)
	}()

	return ch, nil
}

// Close unsubscribes and drains the NATS connection.
func (s *Subscriber) Close() {
	if s.sub != nil {
		_ = s.sub.Unsubscribe()
	}
	if s.conn != nil {
		_ = s.conn.Drain()
	}
}
