package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/underpass-ai/underpass-demo/internal/app/ports"
	"github.com/underpass-ai/underpass-demo/internal/tui/views"
)

// View enumerates the available TUI screens.
type View int

const (
	ViewDashboard View = iota
	ViewRankings
	ViewSampling
	ViewEvents
	ViewDegradation
)

func viewName(v View) string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewRankings:
		return "Rankings"
	case ViewSampling:
		return "Sampling"
	case ViewEvents:
		return "Events"
	case ViewDegradation:
		return "Degradation"
	default:
		return "Unknown"
	}
}

// Deps groups all injected dependencies for the TUI.
type Deps struct {
	PolicyReader ports.PolicyReader
	EventSub     ports.EventSubscriber
}

// Model is the root Bubble Tea model following fleetctl's pattern.
type Model struct {
	currentView View
	deps        Deps
	width       int
	height      int

	dashboard   views.DashboardModel
	rankings    views.RankingsModel
	sampling    views.SamplingModel
	events      views.EventsModel
	degradation views.DegradationModel

	initialised map[View]bool
}

// NewModel creates the root model with all dependencies injected.
func NewModel(deps Deps) Model {
	return Model{
		currentView: ViewDashboard,
		deps:        deps,
		dashboard:   views.NewDashboardModel(deps.PolicyReader),
		rankings:    views.NewRankingsModel(deps.PolicyReader),
		sampling:    views.NewSamplingModel(deps.PolicyReader),
		events:      views.NewEventsModel(deps.EventSub),
		degradation: views.NewDegradationModel(deps.PolicyReader),
		initialised: map[View]bool{ViewDashboard: true},
	}
}

func (m Model) Init() tea.Cmd {
	return m.dashboard.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "d":
			return m.switchView(ViewDashboard)
		case "r":
			return m.switchView(ViewRankings)
		case "s":
			return m.switchView(ViewSampling)
		case "e":
			return m.switchView(ViewEvents)
		case "!":
			return m.switchView(ViewDegradation)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to active view.
	var cmd tea.Cmd
	switch m.currentView {
	case ViewDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ViewRankings:
		m.rankings, cmd = m.rankings.Update(msg)
	case ViewSampling:
		m.sampling, cmd = m.sampling.Update(msg)
	case ViewEvents:
		m.events, cmd = m.events.Update(msg)
	case ViewDegradation:
		m.degradation, cmd = m.degradation.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	// Title bar.
	title := StyleTitle.Render(fmt.Sprintf(" UNDERPASS DEMO  |  %s ", viewName(m.currentView)))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Active view content.
	switch m.currentView {
	case ViewDashboard:
		b.WriteString(m.dashboard.View())
	case ViewRankings:
		b.WriteString(m.rankings.View())
	case ViewSampling:
		b.WriteString(m.sampling.View())
	case ViewEvents:
		b.WriteString(m.events.View())
	case ViewDegradation:
		b.WriteString(m.degradation.View())
	}

	// Help bar at bottom.
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) switchView(target View) (tea.Model, tea.Cmd) {
	m.currentView = target

	if m.initialised[target] {
		// Re-init to refresh data.
		var cmd tea.Cmd
		switch target {
		case ViewDashboard:
			cmd = m.dashboard.Init()
		case ViewRankings:
			cmd = m.rankings.Init()
		case ViewSampling:
			cmd = m.sampling.Init()
		case ViewEvents:
			cmd = m.events.Init()
		case ViewDegradation:
			cmd = m.degradation.Init()
		}
		return m, cmd
	}

	m.initialised[target] = true
	var cmd tea.Cmd
	switch target {
	case ViewDashboard:
		cmd = m.dashboard.Init()
	case ViewRankings:
		cmd = m.rankings.Init()
	case ViewSampling:
		cmd = m.sampling.Init()
	case ViewEvents:
		cmd = m.events.Init()
	case ViewDegradation:
		cmd = m.degradation.Init()
	}
	return m, cmd
}

func (m Model) renderHelp() string {
	bindings := []struct{ key, desc string }{
		{"d", "dashboard"},
		{"r", "rankings"},
		{"s", "sampling"},
		{"e", "events"},
		{"!", "degradation"},
		{"q", "quit"},
	}

	var parts []string
	for _, b := range bindings {
		k := StyleHelpKey.Render(b.key)
		d := StyleHelpDesc.Render(b.desc)
		parts = append(parts, k+" "+d)
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, "  "))
	return StyleStatusBar.Width(m.width).Render(bar)
}
