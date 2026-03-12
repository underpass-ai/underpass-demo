package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/underpass-ai/underpass-demo/internal/app/ports"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

type eventReceivedMsg struct {
	event domain.PolicyUpdateEvent
}

type eventsWatchStartedMsg struct {
	ch     <-chan domain.PolicyUpdateEvent
	cancel context.CancelFunc
}

type eventsStreamClosedMsg struct{}

// EventsModel shows the real-time NATS event stream.
type EventsModel struct {
	subscriber ports.EventSubscriber
	events     []domain.PolicyUpdateEvent
	eventCh    <-chan domain.PolicyUpdateEvent
	cancel     context.CancelFunc
	watching   bool
	err        error
}

func NewEventsModel(sub ports.EventSubscriber) EventsModel {
	return EventsModel{subscriber: sub}
}

func (m EventsModel) Init() tea.Cmd {
	if m.subscriber == nil {
		return nil
	}
	return m.startWatching()
}

func (m EventsModel) startWatching() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		ch, err := m.subscriber.Subscribe(ctx)
		if err != nil {
			cancel()
			return eventsStreamClosedMsg{}
		}
		return eventsWatchStartedMsg{ch: ch, cancel: cancel}
	}
}

func (m EventsModel) waitForEvent() tea.Cmd {
	ch := m.eventCh
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return eventsStreamClosedMsg{}
		}
		return eventReceivedMsg{event: evt}
	}
}

func (m EventsModel) Update(msg tea.Msg) (EventsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case eventsWatchStartedMsg:
		m.eventCh = msg.ch
		m.cancel = msg.cancel
		m.watching = true
		return m, m.waitForEvent()

	case eventReceivedMsg:
		m.events = append(m.events, msg.event)
		return m, m.waitForEvent()

	case eventsStreamClosedMsg:
		m.watching = false
		return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
			return eventsStreamClosedMsg{}
		})
	}
	return m, nil
}

func (m EventsModel) View() string {
	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(heading.Render("  NATS EVENT STREAM"))
	b.WriteString("\n\n")

	if m.subscriber == nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
			"  No NATS connection configured. Start with --nats-url flag."))
		return b.String()
	}

	status := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render("CONNECTED")
	if !m.watching {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Render("RECONNECTING...")
	}
	b.WriteString(fmt.Sprintf("  Status: %s\n", status))
	b.WriteString(fmt.Sprintf("  Subject: %s\n\n",
		lipgloss.NewStyle().Bold(true).Render("tool_learning.policy.updated")))

	if len(m.events) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
			"  Waiting for events..."))
		return b.String()
	}

	eventStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	written := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	filtered := lipgloss.NewStyle().Foreground(lipgloss.Color("210"))

	// Show last 15 events.
	start := 0
	if len(m.events) > 15 {
		start = len(m.events) - 15
	}

	for _, evt := range m.events[start:] {
		b.WriteString(fmt.Sprintf("  %s  schedule=%s  written=%s  filtered=%s\n",
			eventStyle.Render(evt.Ts),
			evt.Schedule,
			written.Render(fmt.Sprintf("%d", evt.PoliciesWritten)),
			filtered.Render(fmt.Sprintf("%d", evt.PoliciesFiltered)),
		))
	}

	return b.String()
}

func (m *EventsModel) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}
