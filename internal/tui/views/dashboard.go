package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/underpass-ai/underpass-demo/internal/app/ports"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

var (
	dashHeading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	dashValue   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	dashMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
)

type policiesLoadedMsg struct {
	policies []domain.ToolPolicy
	err      error
}

// DashboardModel shows the overview of the tool-learning system.
type DashboardModel struct {
	reader   ports.PolicyReader
	policies []domain.ToolPolicy
	loading  bool
	err      error
	loadedAt time.Time
}

// NewDashboardModel creates a new dashboard.
func NewDashboardModel(reader ports.PolicyReader) DashboardModel {
	return DashboardModel{reader: reader, loading: true}
}

func (m DashboardModel) Init() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.reader.ReadAll(context.Background())
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.policies = msg.policies
		m.loadedAt = time.Now()
		sort.Slice(m.policies, func(i, j int) bool {
			return m.policies[i].Confidence > m.policies[j].Confidence
		})
	}
	return m, nil
}

func (m DashboardModel) View() string {
	if m.loading {
		return "  Loading policies..."
	}
	if m.err != nil {
		return fmt.Sprintf("  Error: %v", m.err)
	}

	var b strings.Builder

	// Summary stats.
	totalSamples := int64(0)
	contexts := map[string]bool{}
	var highConf, medConf, lowConf int
	for _, p := range m.policies {
		totalSamples += p.NSamples
		contexts[p.ContextSignature] = true
		switch {
		case p.Confidence >= 0.95:
			highConf++
		case p.Confidence >= 0.85:
			medConf++
		default:
			lowConf++
		}
	}

	b.WriteString(dashHeading.Render("  POLICY OVERVIEW"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Policies active     %s\n", dashValue.Render(fmt.Sprintf("%d", len(m.policies)))))
	b.WriteString(fmt.Sprintf("  Context signatures  %s\n", dashValue.Render(fmt.Sprintf("%d", len(contexts)))))
	b.WriteString(fmt.Sprintf("  Total invocations   %s\n", dashValue.Render(fmt.Sprintf("%d", totalSamples))))
	b.WriteString(fmt.Sprintf("  Last refresh        %s\n", dashMuted.Render(m.loadedAt.Format("15:04:05"))))
	b.WriteString("\n")

	// Confidence distribution.
	b.WriteString(dashHeading.Render("  CONFIDENCE DISTRIBUTION"))
	b.WriteString("\n\n")

	bar := func(count int, color lipgloss.Color, label string) string {
		width := count * 2
		if width > 40 {
			width = 40
		}
		block := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", width))
		return fmt.Sprintf("  %-12s %s %d", label, block, count)
	}

	b.WriteString(bar(highConf, lipgloss.Color("120"), ">= 95%"))
	b.WriteString("\n")
	b.WriteString(bar(medConf, lipgloss.Color("222"), "85-95%"))
	b.WriteString("\n")
	b.WriteString(bar(lowConf, lipgloss.Color("210"), "< 85%"))
	b.WriteString("\n\n")

	// Top 5 and bottom 3.
	b.WriteString(dashHeading.Render("  TOP 5 TOOLS"))
	b.WriteString("\n\n")
	for i, p := range m.policies {
		if i >= 5 {
			break
		}
		confStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
		b.WriteString(fmt.Sprintf("  %d. %-16s %-18s  conf=%s  err=%.1f%%  p95=%dms\n",
			i+1, p.ToolID, p.ContextSignature,
			confStyle.Render(fmt.Sprintf("%.4f", p.Confidence)),
			p.ErrorRate*100, p.P95LatencyMs))
	}

	if len(m.policies) > 3 {
		b.WriteString("\n")
		b.WriteString(dashHeading.Render("  BOTTOM 3 TOOLS"))
		b.WriteString("\n\n")
		start := len(m.policies) - 3
		for i := start; i < len(m.policies); i++ {
			p := m.policies[i]
			confStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
			b.WriteString(fmt.Sprintf("  %d. %-16s %-18s  conf=%s  err=%.1f%%  p95=%dms\n",
				i+1, p.ToolID, p.ContextSignature,
				confStyle.Render(fmt.Sprintf("%.4f", p.Confidence)),
				p.ErrorRate*100, p.P95LatencyMs))
		}
	}

	return b.String()
}
