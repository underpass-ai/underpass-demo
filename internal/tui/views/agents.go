package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type agentTickMsg struct{}

// AgentDispatch represents a NATS event triggering an agent.
type AgentDispatch struct {
	Event     string
	Agent     string
	Model     string
	Tools     []string
	Tokens    int
	Status    string // PENDING, FIRING, RUNNING, COMPLETED
	StartedAt time.Time
	Duration  time.Duration
}

// AgentsModel visualizes event-driven agent dispatch in real time.
type AgentsModel struct {
	dispatches []AgentDispatch
	step       int
	animating  bool
}

func NewAgentsModel() AgentsModel {
	return AgentsModel{}
}

func (m AgentsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return agentTickMsg{}
	}
}

func (m AgentsModel) Update(msg tea.Msg) (AgentsModel, tea.Cmd) {
	switch msg.(type) {
	case agentTickMsg:
		if m.step == 0 {
			m.dispatches = initialDispatches()
			m.step = 1
			m.animating = true
			return m, m.tick()
		}

		if m.step <= len(m.dispatches) {
			idx := m.step - 1
			m.dispatches[idx].Status = "FIRING"
			m.dispatches[idx].StartedAt = time.Now()
			m.step++
			return m, m.tickFast()
		}

		// Progress running agents
		allDone := true
		for i := range m.dispatches {
			d := &m.dispatches[i]
			switch d.Status {
			case "FIRING":
				d.Status = "RUNNING"
				allDone = false
			case "RUNNING":
				d.Duration += 400 * time.Millisecond
				if d.Duration >= agentExpectedDuration(d.Agent) {
					d.Status = "COMPLETED"
				} else {
					allDone = false
				}
			case "PENDING":
				allDone = false
			}
		}

		if allDone {
			m.animating = false
			return m, nil
		}
		return m, m.tick()
	}
	return m, nil
}

func (m AgentsModel) tick() tea.Cmd {
	return tea.Tick(600*time.Millisecond, func(_ time.Time) tea.Msg {
		return agentTickMsg{}
	})
}

func (m AgentsModel) tickFast() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(_ time.Time) tea.Msg {
		return agentTickMsg{}
	})
}

func (m AgentsModel) View() string {
	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(heading.Render("  EVENT-DRIVEN AGENT DISPATCH"))
	b.WriteString("\n\n")

	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	b.WriteString(desc.Render("  NATS events trigger specific agents. No polling. No orchestrator."))
	b.WriteString("\n")
	b.WriteString(desc.Render("  Each agent gets surgical context (394 tokens) and filtered tools."))
	b.WriteString("\n\n")

	if len(m.dispatches) == 0 {
		b.WriteString(desc.Render("  Initializing dispatch table..."))
		return b.String()
	}

	// Dispatch table
	sep := "  " + strings.Repeat("-", 90)
	b.WriteString(heading.Render(fmt.Sprintf("  %-30s %-22s %-14s %-10s %s",
		"Event", "Agent", "Model", "Tokens", "Status")))
	b.WriteString("\n" + sep + "\n")

	for _, d := range m.dispatches {
		eventStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
		agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
		modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("183"))

		var statusRendered string
		switch d.Status {
		case "PENDING":
			statusRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("PENDING")
		case "FIRING":
			statusRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Bold(true).Render(">>> FIRING")
		case "RUNNING":
			spinner := spinnerFrame(d.Duration)
			statusRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true).
				Render(fmt.Sprintf("%s RUNNING %s", spinner, formatDuration(d.Duration)))
		case "COMPLETED":
			statusRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("120")).
				Render(fmt.Sprintf("DONE %s", formatDuration(d.Duration)))
		}

		b.WriteString(fmt.Sprintf("  %s %s %s %5d     %s\n",
			eventStyle.Render(fmt.Sprintf("%-30s", d.Event)),
			agentStyle.Render(fmt.Sprintf("%-22s", d.Agent)),
			modelStyle.Render(fmt.Sprintf("%-14s", d.Model)),
			d.Tokens,
			statusRendered))
	}
	b.WriteString(sep)

	// Tool assignments
	b.WriteString("\n\n")
	b.WriteString(heading.Render("  TOOL ASSIGNMENTS (Thompson Sampling filtered)"))
	b.WriteString("\n\n")

	for _, d := range m.dispatches {
		if d.Status == "PENDING" {
			continue
		}
		agentLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
		toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			agentLabel.Render(fmt.Sprintf("%-22s", d.Agent)),
			toolStyle.Render(strings.Join(d.Tools, ", "))))
	}

	// Architecture note
	b.WriteString("\n\n")
	archStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("183"))
	b.WriteString(archStyle.Render("  Architecture:"))
	b.WriteString("\n")
	b.WriteString(desc.Render("  NATS Subject  -->  Agent Handler  -->  Kernel (394 tokens)  -->  Tools (filtered)"))
	b.WriteString("\n")
	b.WriteString(desc.Render("  No loop. No polling. No central orchestrator. Pure event-driven."))

	if m.animating {
		b.WriteString("\n\n")
		b.WriteString(desc.Render("  Dispatching agents..."))
	} else if m.step > 0 {
		b.WriteString("\n\n")
		complete := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
		b.WriteString(complete.Render("  All agents dispatched and completed. Fleet autonomous."))
	}

	return b.String()
}

func initialDispatches() []AgentDispatch {
	return []AgentDispatch{
		{
			Event:  "sensor.anomaly.detected",
			Agent:  "diagnostic-agent",
			Model:  "qwen3-8b",
			Tools:  []string{"scan.deep", "comm.burst"},
			Tokens: 96,
			Status: "PENDING",
		},
		{
			Event:  "engine.failure.critical",
			Agent:  "repair-agent",
			Model:  "qwen3-8b",
			Tools:  []string{"eng.thrust", "hull.seal", "power.reroute"},
			Tokens: 394,
			Status: "PENDING",
		},
		{
			Event:  "hull.integrity.warning",
			Agent:  "structural-agent",
			Model:  "qwen3-8b",
			Tools:  []string{"hull.seal", "shield.mod"},
			Tokens: 280,
			Status: "PENDING",
		},
		{
			Event:  "policy.updated",
			Agent:  "ranking-agent",
			Model:  "qwen3-8b",
			Tools:  []string{"(internal: Thompson Sampling refresh)"},
			Tokens: 150,
			Status: "PENDING",
		},
		{
			Event:  "repair.strategy.failing",
			Agent:  "strategy-agent",
			Model:  "claude-opus",
			Tools:  []string{"(graph analysis, branch decision)"},
			Tokens: 504,
			Status: "PENDING",
		},
		{
			Event:  "context.rehydrated",
			Agent:  "recovery-agent",
			Model:  "qwen3-8b",
			Tools:  []string{"hull.seal", "life.recycle", "shield.mod"},
			Tokens: 394,
			Status: "PENDING",
		},
	}
}

func agentExpectedDuration(agent string) time.Duration {
	switch agent {
	case "diagnostic-agent":
		return 1200 * time.Millisecond
	case "repair-agent":
		return 2400 * time.Millisecond
	case "structural-agent":
		return 1800 * time.Millisecond
	case "ranking-agent":
		return 800 * time.Millisecond
	case "strategy-agent":
		return 3000 * time.Millisecond
	case "recovery-agent":
		return 2000 * time.Millisecond
	default:
		return 1600 * time.Millisecond
	}
}

func spinnerFrame(d time.Duration) string {
	frames := []string{"|", "/", "-", "\\"}
	idx := int(d.Milliseconds()/200) % len(frames)
	return frames[idx]
}

func formatDuration(d time.Duration) string {
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}
