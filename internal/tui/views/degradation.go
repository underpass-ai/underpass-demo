package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/underpass-ai/underpass-demo/internal/app/ports"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

type degradationStepMsg struct{}

// DegradationModel simulates a tool degradation scenario.
type DegradationModel struct {
	reader   ports.PolicyReader
	policies []domain.ToolPolicy
	step     int
	loading  bool
	err      error
}

func NewDegradationModel(reader ports.PolicyReader) DegradationModel {
	return DegradationModel{reader: reader, loading: true}
}

func (m DegradationModel) Init() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.reader.ReadAll(context.Background())
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

func (m DegradationModel) Update(msg tea.Msg) (DegradationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.policies = msg.policies
		m.step = 0
		sort.Slice(m.policies, func(i, j int) bool {
			return m.policies[i].Confidence > m.policies[j].Confidence
		})

	case tea.KeyMsg:
		if msg.String() == "enter" || msg.String() == " " {
			m.step++
		}
	}
	return m, nil
}

func (m DegradationModel) View() string {
	if m.loading {
		return "  Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("  Error: %v", m.err)
	}

	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	alert := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("210"))
	ok := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	b.WriteString(heading.Render("  DEGRADATION SCENARIO — press SPACE to advance"))
	b.WriteString("\n\n")

	switch {
	case m.step == 0:
		b.WriteString(ok.Render("  [NORMAL] All systems operational"))
		b.WriteString("\n\n")
		b.WriteString("  All tools within SLO thresholds.\n")
		b.WriteString("  Thompson Sampling ranks tools by observed success rate.\n\n")
		b.WriteString(m.renderPolicies(nil, 0))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("  Press SPACE to simulate repo.test degradation..."))

	case m.step == 1:
		b.WriteString(alert.Render("  [ALERT] repo.test error rate spiking!"))
		b.WriteString("\n\n")
		b.WriteString("  repo.test error rate: 5% -> 35%\n")
		b.WriteString("  Root cause: flaky test infrastructure\n\n")
		b.WriteString(m.renderPolicies(map[string]float64{"repo.test": 0.35}, 0))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("  Press SPACE to apply hard constraints (max_error_rate=10%)..."))

	case m.step == 2:
		b.WriteString(alert.Render("  [FILTERED] repo.test excluded by constraint"))
		b.WriteString("\n\n")
		b.WriteString("  Hard constraint: max_error_rate = 10%\n")
		b.WriteString("  repo.test (35%) FILTERED — agents stop using it\n")
		b.WriteString("  repo.build (8%) still eligible — agents use it instead\n\n")
		b.WriteString(m.renderPolicies(map[string]float64{"repo.test": 0.35}, 0.10))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("  Press SPACE to see recovery..."))

	default:
		b.WriteString(ok.Render("  [RECOVERED] repo.test back to normal"))
		b.WriteString("\n\n")
		b.WriteString("  Error rate normalized: 35% -> 8%\n")
		b.WriteString("  Thompson Sampling gradually restores confidence\n")
		b.WriteString("  Exploration ensures the tool gets re-tried\n\n")
		b.WriteString(m.renderPolicies(nil, 0))
		b.WriteString("\n\n")
		b.WriteString(ok.Render("  The fleet adapted automatically. No human intervention needed."))
	}

	return b.String()
}

func (m DegradationModel) renderPolicies(overrides map[string]float64, maxErrorRate float64) string {
	var b strings.Builder
	sep := "  " + strings.Repeat("-", 85)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s %-18s %10s %8s %8s", "Tool", "Context", "Confidence", "ErrRate", "Status")))
	b.WriteString("\n")
	b.WriteString(sep)
	b.WriteString("\n")

	for _, p := range m.policies {
		errRate := p.ErrorRate
		if override, ok := overrides[p.ToolID]; ok {
			errRate = override
		}

		status := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render("OK")
		if maxErrorRate > 0 && errRate > maxErrorRate {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true).Render("FILTERED")
		}

		confStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
		if errRate > 0.15 {
			confStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
		}

		b.WriteString(fmt.Sprintf("  %-16s %-18s %s %7.1f%% %8s\n",
			p.ToolID, p.ContextSignature,
			confStyle.Render(fmt.Sprintf("%10.4f", p.Confidence)),
			errRate*100, status))
	}
	b.WriteString(sep)
	return b.String()
}
