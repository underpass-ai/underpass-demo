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
	ViewMission View = iota
	ViewBridge
	ViewSystems
	ViewSampling
	ViewLog
)

func viewName(v View) string {
	switch v {
	case ViewMission:
		return "Mission"
	case ViewBridge:
		return "Bridge"
	case ViewSystems:
		return "Systems"
	case ViewSampling:
		return "Thompson Sampling"
	case ViewLog:
		return "Ship's Log"
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

	mission  views.MissionModel
	bridge   views.DashboardModel
	systems  views.RankingsModel
	sampling views.SamplingModel
	log      views.EventsModel

	initialised map[View]bool
}

// NewModel creates the root model with all dependencies injected.
func NewModel(deps Deps) Model {
	return Model{
		currentView: ViewMission,
		deps:        deps,
		mission:     views.NewMissionModel(deps.PolicyReader),
		bridge:      views.NewDashboardModel(deps.PolicyReader),
		systems:     views.NewRankingsModel(deps.PolicyReader),
		sampling:    views.NewSamplingModel(deps.PolicyReader),
		log:         views.NewEventsModel(deps.EventSub),
		initialised: map[View]bool{ViewMission: true},
	}
}

func (m Model) Init() tea.Cmd {
	return m.mission.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "m":
			return m.switchView(ViewMission)
		case "b":
			return m.switchView(ViewBridge)
		case "s":
			return m.switchView(ViewSystems)
		case "t":
			return m.switchView(ViewSampling)
		case "l":
			return m.switchView(ViewLog)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to active view.
	var cmd tea.Cmd
	switch m.currentView {
	case ViewMission:
		m.mission, cmd = m.mission.Update(msg)
	case ViewBridge:
		m.bridge, cmd = m.bridge.Update(msg)
	case ViewSystems:
		m.systems, cmd = m.systems.Update(msg)
	case ViewSampling:
		m.sampling, cmd = m.sampling.Update(msg)
	case ViewLog:
		m.log, cmd = m.log.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	// Title bar.
	title := StyleTitle.Render(fmt.Sprintf(" USS UNDERPASS  |  %s ", viewName(m.currentView)))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Active view content.
	switch m.currentView {
	case ViewMission:
		b.WriteString(m.mission.View())
	case ViewBridge:
		b.WriteString(m.bridge.View())
	case ViewSystems:
		b.WriteString(m.systems.View())
	case ViewSampling:
		b.WriteString(m.sampling.View())
	case ViewLog:
		b.WriteString(m.log.View())
	}

	// Help bar at bottom.
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) switchView(target View) (tea.Model, tea.Cmd) {
	m.currentView = target

	// Always re-init to refresh data.
	m.initialised[target] = true
	var cmd tea.Cmd
	switch target {
	case ViewMission:
		cmd = m.mission.Init()
	case ViewBridge:
		cmd = m.bridge.Init()
	case ViewSystems:
		cmd = m.systems.Init()
	case ViewSampling:
		cmd = m.sampling.Init()
	case ViewLog:
		cmd = m.log.Init()
	}
	return m, cmd
}

func (m Model) renderHelp() string {
	bindings := []struct{ key, desc string }{
		{"m", "mission"},
		{"b", "bridge"},
		{"s", "systems"},
		{"t", "thompson"},
		{"l", "log"},
		{"q", "quit"},
	}

	var parts []string
	for _, bind := range bindings {
		k := StyleHelpKey.Render(bind.key)
		d := StyleHelpDesc.Render(bind.desc)
		parts = append(parts, k+" "+d)
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, "  "))
	return StyleStatusBar.Width(m.width).Render(bar)
}
