package embedded

import (
	"context"
	"time"

	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// EventSimulator satisfies ports.EventSubscriber with pre-canned ship events.
type EventSimulator struct{}

func NewEventSimulator() *EventSimulator { return &EventSimulator{} }

func (s *EventSimulator) Subscribe(ctx context.Context) (<-chan domain.PolicyUpdateEvent, error) {
	ch := make(chan domain.PolicyUpdateEvent, 64)
	go func() {
		defer close(ch)
		events := []domain.PolicyUpdateEvent{
			{Event: "policy.updated", Ts: time.Now().Format(time.RFC3339), Schedule: "deep_space_patrol", PoliciesWritten: 8, PoliciesFiltered: 0},
			{Event: "policy.updated", Ts: time.Now().Add(2 * time.Second).Format(time.RFC3339), Schedule: "deep_space_patrol", PoliciesWritten: 8, PoliciesFiltered: 0},
			{Event: "policy.updated", Ts: time.Now().Add(4 * time.Second).Format(time.RFC3339), Schedule: "engine_anomaly", PoliciesWritten: 8, PoliciesFiltered: 1},
			{Event: "policy.updated", Ts: time.Now().Add(6 * time.Second).Format(time.RFC3339), Schedule: "cascade_response", PoliciesWritten: 8, PoliciesFiltered: 2},
			{Event: "policy.updated", Ts: time.Now().Add(8 * time.Second).Format(time.RFC3339), Schedule: "cascade_response", PoliciesWritten: 8, PoliciesFiltered: 3},
			{Event: "policy.updated", Ts: time.Now().Add(10 * time.Second).Format(time.RFC3339), Schedule: "hull_first_protocol", PoliciesWritten: 8, PoliciesFiltered: 1},
			{Event: "policy.updated", Ts: time.Now().Add(12 * time.Second).Format(time.RFC3339), Schedule: "recovery", PoliciesWritten: 8, PoliciesFiltered: 0},
		}
		for _, evt := range events {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
			select {
			case <-ctx.Done():
				return
			case ch <- evt:
			}
		}
		<-ctx.Done()
	}()
	return ch, nil
}

func (s *EventSimulator) Close() {}
