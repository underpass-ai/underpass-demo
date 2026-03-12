package views

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/underpass-ai/underpass-demo/internal/app/ports"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

type samplingTickMsg struct{}

type samplingRound struct {
	scores []scoredTool
}

type scoredTool struct {
	tool  string
	ctx   string
	score float64
}

// SamplingModel shows live Thompson Sampling draws.
type SamplingModel struct {
	reader   ports.PolicyReader
	policies []domain.ToolPolicy
	rounds   []samplingRound
	loading  bool
	err      error
	maxRounds int
}

func NewSamplingModel(reader ports.PolicyReader) SamplingModel {
	return SamplingModel{reader: reader, loading: true, maxRounds: 20}
}

func (m SamplingModel) Init() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.reader.ReadAll(context.Background())
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

func (m SamplingModel) Update(msg tea.Msg) (SamplingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.policies = msg.policies
		m.rounds = nil
		return m, m.tick()

	case samplingTickMsg:
		if len(m.policies) == 0 || len(m.rounds) >= m.maxRounds {
			return m, nil
		}
		round := m.drawRound()
		m.rounds = append(m.rounds, round)
		return m, m.tick()
	}
	return m, nil
}

func (m SamplingModel) tick() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(_ time.Time) tea.Msg {
		return samplingTickMsg{}
	})
}

func (m SamplingModel) drawRound() samplingRound {
	scores := make([]scoredTool, len(m.policies))
	for i, p := range m.policies {
		scores[i] = scoredTool{
			tool:  p.ToolID,
			ctx:   p.ContextSignature,
			score: betaSample(p.Alpha, p.Beta),
		}
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	return samplingRound{scores: scores}
}

func (m SamplingModel) View() string {
	if m.loading {
		return "  Loading policies for sampling..."
	}
	if m.err != nil {
		return fmt.Sprintf("  Error: %v", m.err)
	}

	var b strings.Builder

	heading := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(heading.Render("  THOMPSON SAMPLING — LIVE DRAWS"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
		"  Each round draws from Beta(alpha, beta). Watch rankings shift in real-time."))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render(
		"  This is exploration vs exploitation — the core of Bayesian tool ranking."))
	b.WriteString("\n\n")

	winner := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	runner := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	rest := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	roundLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Bold(true)

	for i, r := range m.rounds {
		b.WriteString(roundLabel.Render(fmt.Sprintf("  Round %2d: ", i+1)))
		for j, s := range r.scores {
			if j >= 6 {
				b.WriteString(rest.Render("..."))
				break
			}
			if j > 0 {
				b.WriteString(rest.Render("  >  "))
			}
			label := fmt.Sprintf("%s (%.3f)", s.tool, s.score)
			switch {
			case j == 0:
				b.WriteString(winner.Render(label))
			case j <= 2:
				b.WriteString(runner.Render(label))
			default:
				b.WriteString(rest.Render(label))
			}
		}
		b.WriteString("\n")
	}

	if len(m.rounds) < m.maxRounds && len(m.policies) > 0 {
		b.WriteString("\n  Sampling...")
	}

	return b.String()
}

// betaSample draws from Beta(alpha, beta) using the gamma trick.
func betaSample(alpha, beta float64) float64 {
	x := gammaSample(alpha)
	y := gammaSample(beta)
	if x+y == 0 {
		return 0.5
	}
	return x / (x + y)
}

func gammaSample(shape float64) float64 {
	if shape < 1 {
		u := rand.Float64()
		return gammaSample(shape+1) * math.Pow(u, 1.0/shape)
	}
	d := shape - 1.0/3.0
	c := 1.0 / (3.0 * math.Sqrt(d))
	for {
		var x float64
		for {
			x = rand.NormFloat64()
			if 1+c*x > 0 {
				break
			}
		}
		v := (1 + c*x) * (1 + c*x) * (1 + c*x)
		u := rand.Float64()
		if u < 1-0.0331*(x*x)*(x*x) {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}
