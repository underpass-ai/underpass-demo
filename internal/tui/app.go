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
	ViewAgents
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
	case ViewAgents:
		return "Agent Dispatch"
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
	agents   views.AgentsModel
	log      views.EventsModel

	initialized map[View]bool
}

// NewModel creates the root model with all dependencies injected.
func NewModel(deps Deps) Model {
	sharedLog := &views.SharedLog{}
	logView := views.NewEventsModel(deps.EventSub)
	logView.MissionLog = sharedLog
	return Model{
		currentView: ViewMission,
		deps:        deps,
		mission:     views.NewMissionModel(deps.PolicyReader, sharedLog),
		bridge:      views.NewDashboardModel(deps.PolicyReader),
		systems:     views.NewRankingsModel(deps.PolicyReader),
		sampling:    views.NewSamplingModel(deps.PolicyReader),
		agents:      views.NewAgentsModel(),
		log:         logView,
		initialized: map[View]bool{ViewMission: true},
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
		case "a":
			return m.switchView(ViewAgents)
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
	case ViewAgents:
		m.agents, cmd = m.agents.Update(msg)
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
	case ViewAgents:
		b.WriteString(m.agents.View())
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

	// Mission keeps its phase state — don't re-init once started.
	// Other views re-init to refresh data (policies, sampling rounds, etc.).
	if target == ViewMission {
		if !m.initialized[target] {
			m.initialized[target] = true
			return m, m.mission.Init()
		}
		return m, nil
	}

	m.initialized[target] = true
	var cmd tea.Cmd
	switch target {
	case ViewBridge:
		cmd = m.bridge.Init()
	case ViewSystems:
		cmd = m.systems.Init()
	case ViewSampling:
		cmd = m.sampling.Init()
	case ViewAgents:
		cmd = m.agents.Init()
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
		{"a", "agents"},
		{"l", "log"},
		{"q", "quit"},
	}

	parts := make([]string, 0, len(bindings))
	for _, bind := range bindings {
		k := StyleHelpKey.Render(bind.key)
		d := StyleHelpDesc.Render(bind.desc)
		parts = append(parts, k+" "+d)
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, "  "))
	return StyleStatusBar.Width(m.width).Render(bar)
}
