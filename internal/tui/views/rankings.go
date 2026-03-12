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

// RankingsModel shows a ranked table of all policies.
type RankingsModel struct {
	reader   ports.PolicyReader
	policies []domain.ToolPolicy
	loading  bool
	err      error
}

func NewRankingsModel(reader ports.PolicyReader) RankingsModel {
	return RankingsModel{reader: reader, loading: true}
}

func (m RankingsModel) Init() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.reader.ReadAll(context.Background())
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

func (m RankingsModel) Update(msg tea.Msg) (RankingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.policies = msg.policies
		sort.Slice(m.policies, func(i, j int) bool {
			return m.policies[i].Confidence > m.policies[j].Confidence
		})
	}
	return m, nil
}

func (m RankingsModel) View() string {
	if m.loading {
		return "  Loading..."
	}
	if m.err != nil {
		return fmt.Sprintf("  Error: %v", m.err)
	}

	var b strings.Builder

	header := fmt.Sprintf("  %-3s %-16s %-18s %7s %8s %8s %10s %8s %7s",
		"#", "Tool", "Context", "Samples", "Alpha", "Beta", "Confidence", "ErrRate", "P95ms")
	sep := "  " + strings.Repeat("-", 100)

	headingStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(headingStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(sep)
	b.WriteString("\n")

	for i, p := range m.policies {
		confColor := lipgloss.Color("210") // red
		if p.Confidence >= 0.95 {
			confColor = lipgloss.Color("120") // green
		} else if p.Confidence >= 0.85 {
			confColor = lipgloss.Color("222") // yellow
		}

		confStyle := lipgloss.NewStyle().Foreground(confColor).Bold(p.Confidence >= 0.95)
		conf := confStyle.Render(fmt.Sprintf("%10.4f", p.Confidence))

		b.WriteString(fmt.Sprintf("  %-3d %-16s %-18s %7d %8.0f %8.0f %s %7.1f%% %7d\n",
			i+1, p.ToolID, p.ContextSignature,
			p.NSamples, p.Alpha, p.Beta, conf,
			p.ErrorRate*100, p.P95LatencyMs))
	}

	b.WriteString(sep)
	return b.String()
}
