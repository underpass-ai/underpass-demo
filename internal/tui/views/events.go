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

// EventsModel shows the real-time NATS event stream and mission log.
type EventsModel struct {
	subscriber ports.EventSubscriber
	events     []domain.PolicyUpdateEvent
	eventCh    <-chan domain.PolicyUpdateEvent
	cancel     context.CancelFunc
	watching   bool
	MissionLog *SharedLog
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

var (
	logTime     = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	logInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	logWarn     = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Bold(true)
	logCrit     = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true)
	logEvent    = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	logAgent    = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))
	logKernel   = lipgloss.NewStyle().Foreground(lipgloss.Color("115")).Bold(true)
	logNatsEvt  = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	logNatsOk   = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	logNatsFilt = lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
)

func levelStyle(level string) lipgloss.Style {
	switch level {
	case "WARN":
		return logWarn
	case "CRITICAL":
		return logCrit
	case "EVENT":
		return logEvent
	case "AGENT":
		return logAgent
	case "KERNEL":
		return logKernel
	default:
		return logInfo
	}
}

func (m EventsModel) View() string {
	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(heading.Render("  SHIP'S LOG"))
	b.WriteString("\n\n")

	// Mission log entries.
	if m.MissionLog != nil && len(m.MissionLog.Entries) > 0 {
		b.WriteString(heading.Render("  --- MISSION ---"))
		b.WriteString("\n\n")

		entries := m.MissionLog.Entries
		start := 0
		if len(entries) > 20 {
			start = len(entries) - 20
		}
		for _, e := range entries[start:] {
			ls := levelStyle(e.Level)
			fmt.Fprintf(&b, "  %s  %s  %s\n",
				logTime.Render(e.Time),
				ls.Render(fmt.Sprintf("%-8s", e.Level)),
				ls.Render(e.Message),
			)
		}
		b.WriteString("\n")
	}

	// NATS event stream.
	if m.subscriber != nil {
		b.WriteString(heading.Render("  --- NATS STREAM ---"))
		b.WriteString("\n\n")

		status := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render("CONNECTED")
		if !m.watching {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Render("RECONNECTING...")
		}
		fmt.Fprintf(&b, "  Status: %s\n\n", status)

		if len(m.events) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
				"  Waiting for events..."))
		} else {
			start := 0
			if len(m.events) > 10 {
				start = len(m.events) - 10
			}
			for _, evt := range m.events[start:] {
				fmt.Fprintf(&b, "  %s  schedule=%s  written=%s  filtered=%s\n",
					logNatsEvt.Render(evt.Ts),
					evt.Schedule,
					logNatsOk.Render(fmt.Sprintf("%d", evt.PoliciesWritten)),
					logNatsFilt.Render(fmt.Sprintf("%d", evt.PoliciesFiltered)),
				)
			}
		}
	} else if m.MissionLog == nil || len(m.MissionLog.Entries) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
			"  No events yet. Start the mission (m) and advance with SPACE."))
	}

	return b.String()
}

func (m *EventsModel) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
}
