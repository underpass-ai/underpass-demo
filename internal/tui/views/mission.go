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

// LogEntry represents a single line in the Ship's Log.
type LogEntry struct {
	Time    string
	Level   string // INFO, WARN, CRITICAL, EVENT, AGENT, KERNEL
	Message string
}

// SharedLog is a shared log buffer between Mission and Ship's Log views.
// Bubble Tea copies models on Update, so we need a shared pointer.
type SharedLog struct {
	Entries []LogEntry
}

// ── Bubble Tea messages ─────────────────────────────────────────────────────
// policiesLoadedMsg is declared in dashboard.go (shared across views).

type kernelContextMsg struct {
	bundle *domain.ContextResult
	graph  *domain.GraphResult
	err    error
}

// MissionModel drives the spaceship scenario — the main demo experience.
// Press SPACE to advance through 10 phases of the USS Underpass mission.
type MissionModel struct {
	reader   ports.PolicyReader
	context  ports.ContextProvider
	recorder ports.SessionRecorder
	policies []domain.ToolPolicy
	phase    int
	loading  bool
	err      error
	Log      *SharedLog

	// Kernel state for phases 7-8.
	kernelBundle  *domain.ContextResult
	kernelGraph   *domain.GraphResult
	kernelLoading bool
	kernelErr     error
}

func NewMissionModel(reader ports.PolicyReader, ctx ports.ContextProvider, rec ports.SessionRecorder, log *SharedLog) MissionModel {
	return MissionModel{reader: reader, context: ctx, recorder: rec, loading: true, Log: log}
}

func (m MissionModel) Init() tea.Cmd {
	return func() tea.Msg {
		policies, err := m.reader.ReadAll(context.Background())
		return policiesLoadedMsg{policies: policies, err: err}
	}
}

func (m MissionModel) Update(msg tea.Msg) (MissionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case policiesLoadedMsg:
		m.loading = false
		m.err = msg.err
		m.policies = msg.policies
		m.phase = 0
		sort.Slice(m.policies, func(i, j int) bool {
			return m.policies[i].Confidence > m.policies[j].Confidence
		})
		m.emitLogs(0)
		m.recordPhase(0)

	case kernelContextMsg:
		m.kernelLoading = false
		m.kernelErr = msg.err
		if msg.err == nil {
			m.kernelBundle = msg.bundle
			m.kernelGraph = msg.graph
			ts := time.Now().Format("15:04:05")
			nodes := 0
			rels := 0
			tokens := uint32(0)
			snap := ""
			if msg.bundle != nil {
				nodes = len(msg.bundle.Nodes)
				rels = len(msg.bundle.Relationships)
				tokens = msg.bundle.TokenCount
				snap = msg.bundle.SnapshotID
			}
			m.Log.Entries = append(m.Log.Entries, LogEntry{
				Time: ts, Level: "KERNEL",
				Message: fmt.Sprintf("GetContext: %d nodes, %d rels, %d tokens, %s", nodes, rels, tokens, snap),
			})
			m.recordKernel(msg.bundle)
		} else {
			ts := time.Now().Format("15:04:05")
			m.Log.Entries = append(m.Log.Entries, LogEntry{
				Time: ts, Level: "KERNEL",
				Message: fmt.Sprintf("GetContext failed: %v [SIMULATED fallback]", msg.err),
			})
		}

	case tea.KeyMsg:
		if msg.String() == "enter" || msg.String() == " " {
			if m.phase < 9 {
				m.phase++
				m.emitLogs(m.phase)
				m.recordPhase(m.phase)

				// Phase 7 → fire async kernel call.
				if m.phase == 7 && m.context != nil {
					m.kernelLoading = true
					return m, m.fetchKernelContext()
				}
			}
		}
	}
	return m, nil
}

// fetchKernelContext fires an async tea.Cmd to call the kernel.
func (m MissionModel) fetchKernelContext() tea.Cmd {
	cp := m.context
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := domain.ContextRequest{
			RootNodeID:  "node:mission:engine-core-failure",
			Role:        "implementer",
			Phase:       "rehydration",
			TokenBudget: 4000,
			Scopes:      []string{"engine", "hull", "power"},
		}
		bundle, bundleErr := cp.GetContext(ctx, req)
		if bundleErr != nil {
			return kernelContextMsg{err: bundleErr}
		}
		graph, graphErr := cp.GetGraphRelationships(ctx, req.RootNodeID, "mission", 3)
		return kernelContextMsg{bundle: bundle, graph: graph, err: graphErr}
	}
}

func (m *MissionModel) recordPhase(phase int) {
	if m.recorder == nil {
		return
	}
	_ = m.recorder.Record(domain.SessionRecord{
		Kind:  "phase",
		Ts:    time.Now(),
		Phase: phase,
		Data:  phaseLogs[phase],
	})
}

func (m *MissionModel) recordKernel(bundle *domain.ContextResult) {
	if m.recorder == nil || bundle == nil {
		return
	}
	_ = m.recorder.Record(domain.SessionRecord{
		Kind:  "kernel",
		Ts:    time.Now(),
		Phase: m.phase,
		Data:  bundle,
	})
}

// sysOverride replaces live metrics for a tool during a scenario phase.
type sysOverride struct {
	errRate float64
	p95ms   int64
}

// phaseLog describes what happened in a phase — used for both rendering and Ship's Log.
type phaseLog struct {
	event   string // NATS subject (empty = no event)
	agent   string // agent dispatched
	model   string // qwen3-8b or claude-opus
	tokens  int
	entries []LogEntry // level + message pairs
}

var phaseLogs = [10]phaseLog{
	0: {entries: []LogEntry{
		{"", "INFO", "Ship systems initialized. All subsystems nominal."},
		{"", "INFO", "AI agent fleet: 4 agents active, routine patrol."},
	}},
	1: {event: "sensor.anomaly.detected", agent: "diagnostic-agent", model: "qwen3-8b", tokens: 96, entries: []LogEntry{
		{"", "WARN", "eng.thrust latency 890ms → 1200ms, error rate climbing"},
	}},
	2: {event: "engine.failure.critical", agent: "repair-agent", model: "qwen3-8b", tokens: 394, entries: []LogEntry{
		{"", "CRITICAL", "COOLANT RUPTURE — eng.thrust error rate 42%, latency 2800ms"},
	}},
	3: {event: "hull.integrity.warning", agent: "structural-agent", model: "qwen3-8b", tokens: 280, entries: []LogEntry{
		{"", "CRITICAL", "CASCADE: power.reroute 28%, shield.mod 22% — 3 systems compromised"},
	}},
	4: {event: "policy.updated", agent: "ranking-agent", model: "qwen3-8b", tokens: 150, entries: []LogEntry{
		{"", "INFO", "Thompson Sampling: hard constraint max_error_rate=20% engaged"},
		{"", "INFO", "FILTERED: eng.thrust (42%), power.reroute (28%), shield.mod (22%)"},
	}},
	5: {agent: "repair-agent", model: "qwen3-8b", entries: []LogEntry{
		{"", "WARN", "Engine repair attempt #3 failed. Hull stress at 76%"},
		{"", "WARN", "Strategy counterproductive — repairs causing micro-fractures"},
	}},
	6: {event: "repair.strategy.failing", agent: "strategy-agent", model: "claude-opus", tokens: 504, entries: []LogEntry{
		{"", "INFO", "MODEL ROUTING: escalating to Claude Opus ($0.006, 394 tokens)"},
		{"", "AGENT", "Task graph analysis: 3 failed attempts. Hull-first recommended."},
	}},
	7: {event: "context.rehydrated", agent: "recovery-agent", model: "qwen3-8b", tokens: 394, entries: []LogEntry{
		{"", "INFO", "REHYDRATION: checkpoint ALPHA-3 loaded, new branch created"},
		{"", "INFO", "Rehydration bundle: 394 / 4,000 token budget"},
	}},
	8: {agent: "recovery-agent", model: "qwen3-8b", entries: []LogEntry{
		{"", "INFO", "Hull sealed — error rate 6%, structural integrity 98%"},
		{"", "INFO", "Power grid stabilizing — error rate 12%"},
	}},
	9: {entries: []LogEntry{
		{"", "INFO", "Engine core repaired. All systems nominal."},
		{"", "INFO", "MISSION COMPLETE — no human intervention required"},
	}},
}

func (m *MissionModel) emitLogs(phase int) {
	pl := phaseLogs[phase]
	ts := time.Now().Format("15:04:05")
	if pl.event != "" {
		m.Log.Entries = append(m.Log.Entries, LogEntry{
			Time: ts, Level: "EVENT",
			Message: fmt.Sprintf("%s → %s", pl.event, pl.agent),
		})
	}
	if pl.agent != "" && pl.tokens > 0 {
		m.Log.Entries = append(m.Log.Entries, LogEntry{
			Time: ts, Level: "AGENT",
			Message: fmt.Sprintf("[%s/%s] dispatched (%d tokens)", pl.agent, pl.model, pl.tokens),
		})
	}
	for _, e := range pl.entries {
		m.Log.Entries = append(m.Log.Entries, LogEntry{Time: ts, Level: e.Level, Message: e.Message})
	}
}

// Mission styles.
var (
	mBanner = lipgloss.NewStyle().Bold(true)
	mNarr   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	mHint   = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
	mAgent  = lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Bold(true)
	mLog    = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))
	mGrBdr  = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))
	mDone   = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
	mFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true)
	mActv   = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	mPend   = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	mBndl   = lipgloss.NewStyle().Foreground(lipgloss.Color("147"))
	mEvt    = lipgloss.NewStyle().Foreground(lipgloss.Color("222")).Bold(true)
	mSim    = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
)

func (m MissionModel) View() string {
	if m.loading {
		return "  Initializing ship systems..."
	}
	if m.err != nil {
		return fmt.Sprintf("  SYSTEM ERROR: %v", m.err)
	}

	var b strings.Builder
	switch m.phase {
	case 0:
		m.phaseCruise(&b)
	case 1:
		m.phaseAnomaly(&b)
	case 2:
		m.phaseEngineFailure(&b)
	case 3:
		m.phaseCascade(&b)
	case 4:
		m.phaseAdapt(&b)
	case 5:
		m.phaseWrongPath(&b)
	case 6:
		m.phaseEscalation(&b)
	case 7:
		m.phaseRehydration(&b)
	case 8:
		m.phaseNewBranch(&b)
	default:
		m.phaseResolution(&b)
	}
	return b.String()
}

// ─── Phase renderers ────────────────────────────────────────────────────────

func (m MissionModel) phaseCruise(b *strings.Builder) {
	writeBanner(b, "120", "NOMINAL", "ALL SYSTEMS OPERATIONAL")
	writeNarr(b,
		"USS Underpass cruising through Sector 7-G. Deep space exploration mission.",
		"AI agent fleet: 4 agents active, performing routine maintenance.",
		"Thompson Sampling has ranked all ship systems by observed reliability.",
	)
	b.WriteString(m.renderSystems(nil, 0))
	writeHint(b, "Press SPACE to continue mission...")
}

func (m MissionModel) phaseAnomaly(b *strings.Builder) {
	writeBanner(b, "222", "WARNING", "SENSOR ANOMALY DETECTED")

	writeEvent(b, "sensor.anomaly.detected", "diagnostic-agent")

	writeNarr(b,
		"Deep scan array detecting intermittent power surges in starboard engine.",
		"eng.thrust latency increasing: 890ms -> 1200ms. Error rate climbing.",
		"diagnostic-agent fired automatically from NATS event.",
	)
	b.WriteString(m.renderSystems(map[string]sysOverride{
		"eng.thrust": {errRate: 0.18, p95ms: 1200},
	}, 0))
	writeHint(b, "Press SPACE — engine situation deteriorating...")
}

func (m MissionModel) phaseEngineFailure(b *strings.Builder) {
	writeBanner(b, "210", "CRITICAL", "ENGINE CORE FAILURE")

	writeEvent(b, "engine.failure.critical", "repair-agent")

	writeNarr(b,
		"COOLANT RUPTURE in engine core! Main propulsion OFFLINE.",
		"eng.thrust error rate: 12% -> 42%. Response time: 2800ms.",
		"repair-agent activated. Emergency protocols engaged.",
	)
	b.WriteString(m.renderSystems(map[string]sysOverride{
		"eng.thrust": {errRate: 0.42, p95ms: 2800},
	}, 0))
	writeHint(b, "Press SPACE — failure cascading through power grid...")
}

func (m MissionModel) phaseCascade(b *strings.Builder) {
	writeBanner(b, "210", "CRITICAL", "CASCADE FAILURE — MULTIPLE SYSTEMS")

	writeEvent(b, "hull.integrity.warning", "structural-agent")

	writeNarr(b,
		"Engine failure cascading through power grid. Overload protection engaged.",
		"power.reroute degraded: error rate 28%, latency 1500ms.",
		"shield.mod stressed: error rate 22%, intermittent shield drops.",
		"Three subsystems now compromised. structural-agent deployed.",
	)
	b.WriteString(m.renderSystems(map[string]sysOverride{
		"eng.thrust":    {errRate: 0.42, p95ms: 2800},
		"power.reroute": {errRate: 0.28, p95ms: 1500},
		"shield.mod":    {errRate: 0.22, p95ms: 1200},
	}, 0))
	writeHint(b, "Press SPACE — Thompson Sampling is adapting...")
}

func (m MissionModel) phaseAdapt(b *strings.Builder) {
	writeBanner(b, "117", "ADAPTING", "THOMPSON SAMPLING RESPONSE")

	writeEvent(b, "policy.updated", "ranking-agent (Thompson Sampling refresh)")

	writeNarr(b,
		"Tool-learning pipeline detected the degradation automatically.",
		"Hard constraint engaged: max_error_rate = 20%.",
		"eng.thrust (42%) FILTERED — agents can no longer select it.",
		"power.reroute (28%) FILTERED — agents switch to alternatives.",
		"shield.mod (22%) FILTERED — coverage managed by backup mode.",
	)
	b.WriteString(m.renderSystems(map[string]sysOverride{
		"eng.thrust":    {errRate: 0.42, p95ms: 2800},
		"power.reroute": {errRate: 0.28, p95ms: 1500},
		"shield.mod":    {errRate: 0.22, p95ms: 1200},
	}, 0.20))
	writeHint(b, "Press SPACE — repair-agent attempts engine repair...")
}

func (m MissionModel) phaseWrongPath(b *strings.Builder) {
	writeBanner(b, "210", "FAILING", "WRONG REPAIR STRATEGY")
	writeNarr(b,
		"repair-agent attempted direct engine repair while hull was compromised.",
		"Each repair attempt increasing hull stress. hull.seal degrading.",
		"Engine getting WORSE: error rate 42% -> 55%. Strategy is failing.",
	)

	writeAgentLog(b, "REPAIR-AGENT [qwen3-8b]",
		"Engine repair attempt #3 failed. Hull stress at 76%.",
		"Vibration from repairs causing micro-fractures in hull plates.",
		"Current strategy is counterproductive. Confidence dropping.",
	)

	b.WriteString(m.renderSystems(map[string]sysOverride{
		"eng.thrust":    {errRate: 0.55, p95ms: 3200},
		"hull.seal":     {errRate: 0.19, p95ms: 900},
		"power.reroute": {errRate: 0.32, p95ms: 1800},
		"shield.mod":    {errRate: 0.24, p95ms: 1300},
	}, 0.20))
	writeHint(b, "Press SPACE — system escalates to Claude Opus...")
}

func (m MissionModel) phaseEscalation(b *strings.Builder) {
	writeBanner(b, "183", "ESCALATING", "MODEL ROUTING — STRATEGIC DECISION")

	writeNarr(b,
		"Qwen3-8B repair strategy failing after 3 attempts.",
		"Thompson Sampling detects: (qwen3-8b, eng.thrust) success rate dropping.",
		"System escalates strategic decision to Claude Opus (API call).",
	)

	writeAgentLog(b, "REPAIR-AGENT [qwen3-8b]",
		"Escalating strategic decision. Local model insufficient for this.",
	)

	writeAgentLog(b, "STRATEGY-AGENT [claude-opus]",
		"Analyzing task graph. 3 failed repair attempts detected.",
		"Pattern: each attempt worsens hull integrity by ~4%.",
		"Root cause: repairing engine with compromised hull is unsafe.",
		"RECOMMENDATION: Initiate context rehydration. Hull-first protocol.",
	)

	// Cost callout
	costBox := lipgloss.NewStyle().Foreground(lipgloss.Color("147")).Bold(true)
	b.WriteString("\n")
	b.WriteString(costBox.Render("  COST: This ONE strategic call = $0.006 (Opus, 394 surgical tokens)"))
	b.WriteString("\n")
	b.WriteString(costBox.Render("        vs $0.093 if we sent 6,190 traditional tokens"))
	b.WriteString("\n")
	b.WriteString(costBox.Render("        95% of calls stayed on local GPU ($0). This is the 5% that matters."))
	b.WriteString("\n")

	writeHint(b, "Press SPACE — initiating CONTEXT REHYDRATION...")
}

func (m MissionModel) phaseRehydration(b *strings.Builder) {
	writeBanner(b, "183", "REHYDRATING", "CONTEXT REHYDRATION — KERNEL ROLLBACK")

	writeEvent(b, "context.rehydrated", "recovery-agent (resume from checkpoint)")

	writeAgentLog(b, "STRATEGY-AGENT [claude-opus]",
		"ACTION: Roll back to checkpoint ALPHA-3 (damage assessment).",
		"Creating new solution branch: hull-first protocol.",
		"Rehydration bundle loaded from Neo4j graph.",
	)

	if m.kernelLoading {
		b.WriteString("\n  " + mActv.Render("Fetching context from kernel...") + "\n")
		writeHint(b, "Waiting for kernel response...")
		return
	}

	// Use kernel data if available, otherwise hardcoded (simulated).
	bundle := m.kernelBundle
	simulated := bundle == nil

	if simulated {
		b.WriteString("  " + mSim.Render("[SIMULATED]") + "\n")
	}

	m.renderTaskGraph(b, bundle, simulated)
	m.renderBundleMetadata(b, bundle, simulated)

	writeHint(b, "Press SPACE — hull-first protocol in action...")
}

func (m MissionModel) phaseNewBranch(b *strings.Builder) {
	writeBanner(b, "117", "RECOVERING", "HULL-FIRST PROTOCOL — NEW BRANCH ACTIVE")

	simulated := m.kernelBundle == nil
	if simulated {
		b.WriteString("  " + mSim.Render("[SIMULATED]") + "\n")
	}

	writeNarr(b,
		"New strategy working! Hull sealed first, then power stabilized.",
		"hull.seal back to optimal: error rate 6%. Structural integrity 98%.",
		"power.reroute recovering: error rate 12%. Grid stabilizing.",
		"Engine repair now safe to attempt under stable conditions.",
	)

	writeAgentLog(b, "RECOVERY-AGENT [qwen3-8b]",
		"Hull integrity restored. Safe to proceed with engine repair.",
		"Rehydrated context provided correct repair sequence.",
		"Task graph branch B validated. Executing step [5].",
	)

	b.WriteString(m.renderSystems(map[string]sysOverride{
		"hull.seal":     {errRate: 0.06, p95ms: 380},
		"eng.thrust":    {errRate: 0.25, p95ms: 1400},
		"power.reroute": {errRate: 0.12, p95ms: 800},
		"shield.mod":    {errRate: 0.14, p95ms: 850},
	}, 0.20))
	writeHint(b, "Press SPACE — mission resolution...")
}

func (m MissionModel) phaseResolution(b *strings.Builder) {
	writeBanner(b, "120", "NOMINAL", "SHIP RECOVERED — ALL SYSTEMS OPERATIONAL")
	writeNarr(b,
		"Engine core repaired under safe conditions. All systems nominal.",
		"Thompson Sampling gradually restoring confidence scores.",
		"The fleet adapted automatically. No human intervention needed.",
		"",
		"What you just saw:",
		"  [P1] Thompson Sampling learned which tools were failing",
		"  [P2] NATS events fired the right agent for each problem",
		"  [P3] Kernel built surgical context: 394 tokens, not 6,190",
		"  [P4] Agent rolled back to checkpoint, branched new strategy",
		"  [P5] System escalated to Claude Opus only when it mattered",
	)

	b.WriteString(m.renderSystems(nil, 0))

	complete := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("120"))
	b.WriteString("\n")
	b.WriteString(complete.Render("  MISSION COMPLETE. The USS Underpass sails on."))
	b.WriteString("\n")
}

// ─── Kernel data renderers ──────────────────────────────────────────────────

func (m MissionModel) renderTaskGraph(b *strings.Builder, bundle *domain.ContextResult, simulated bool) {
	graphTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("183"))
	b.WriteString("\n")
	b.WriteString(graphTitle.Render("  --- TASK GRAPH --- USS Underpass: Engine Core Failure ---"))
	b.WriteString("\n\n")

	if !simulated && bundle != nil {
		m.renderKernelGraph(b, bundle)
	} else {
		m.renderHardcodedGraph(b)
	}
}

func (m MissionModel) renderKernelGraph(b *strings.Builder, bundle *domain.ContextResult) {
	statusStyle := func(status string) lipgloss.Style {
		switch status {
		case "done":
			return mDone
		case "abandoned":
			return mFail
		case "active":
			return mActv
		default:
			return mPend
		}
	}

	statusLabel := func(status string) string {
		switch status {
		case "done":
			return "DONE"
		case "abandoned":
			return "ABANDONED"
		case "active":
			return "IN PROGRESS"
		default:
			return "PENDING"
		}
	}

	statusIcon := func(status string) string {
		switch status {
		case "done":
			return "*"
		case "abandoned":
			return "x"
		case "active":
			return "o"
		default:
			return "-"
		}
	}

	// Build adjacency: sourceID → []targetNode
	type edge struct {
		rel  domain.GraphRelationship
		node domain.GraphNode
	}
	nodeMap := make(map[string]domain.GraphNode)
	for _, n := range bundle.Nodes {
		nodeMap[n.ID] = n
	}
	children := make(map[string][]edge)
	for _, r := range bundle.Relationships {
		if target, ok := nodeMap[r.TargetID]; ok {
			children[r.SourceID] = append(children[r.SourceID], edge{rel: r, node: target})
		}
	}

	// Walk the tree from root.
	var walk func(nodeID, prefix string, last bool)
	seq := 0
	walk = func(nodeID, prefix string, last bool) {
		n, ok := nodeMap[nodeID]
		if !ok {
			return
		}
		seq++
		st := statusStyle(n.Status)
		icon := statusIcon(n.Status)
		label := statusLabel(n.Status)

		connector := "|-- "
		if last {
			connector = "'-- "
		}
		if prefix == "" {
			// Root node.
			b.WriteString(st.Render(fmt.Sprintf("   %s [%d] %s %s %s",
				icon, seq, n.Label, strings.Repeat(".", max(1, 50-len(n.Label))), label)))
			b.WriteString("\n")
		} else {
			b.WriteString(mGrBdr.Render(prefix+connector) + st.Render(fmt.Sprintf("%s [%d] %s %s %s",
				icon, seq, n.Label, strings.Repeat(".", max(1, 45-len(n.Label)-len(prefix))), label)))
			b.WriteString("\n")
		}

		edges := children[nodeID]
		for i, e := range edges {
			childPrefix := prefix
			if prefix == "" {
				childPrefix = "    "
			} else if last {
				childPrefix = prefix + "     "
			} else {
				childPrefix = prefix + "|    "
			}

			// Show relationship label if present (e.g. "Path A", "Path B").
			if e.rel.Label != "" {
				b.WriteString(mGrBdr.Render(childPrefix))
				b.WriteString("\n")
			}

			walk(e.node.ID, childPrefix, i == len(edges)-1)
		}
	}

	// Find root (the mission node).
	rootID := bundle.RootNodeID
	if rootID == "" && len(bundle.Nodes) > 0 {
		rootID = bundle.Nodes[0].ID
	}
	walk(rootID, "", true)

	// Show node details if present.
	if len(bundle.Details) > 0 {
		for _, d := range bundle.Details {
			if n, ok := nodeMap[d.NodeID]; ok && n.Status == "abandoned" {
				b.WriteString(mGrBdr.Render("         ") + mFail.Render("  "+d.Description))
				b.WriteString("\n")
				if model, ok := d.Properties["model"]; ok {
					extra := "Model: " + model
					if esc, ok := d.Properties["escalated_to"]; ok {
						extra += ". Escalated to " + esc + "."
					}
					b.WriteString(mGrBdr.Render("         ") + mFail.Render("  "+extra))
					b.WriteString("\n")
				}
			}
		}
	}

	b.WriteString("\n")
}

func (m MissionModel) renderHardcodedGraph(b *strings.Builder) {
	b.WriteString(mDone.Render("   * [1] Diagnose anomaly ................................. DONE"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |"))
	b.WriteString("\n")
	b.WriteString(mDone.Render("   * [2] Assess cascade damage ............................ DONE"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |-- ") + mFail.Render("x Path A: Direct engine repair ............... ABANDONED"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |   ") + mFail.Render("  3 attempts. Hull stress +12%. Counterproductive."))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |   ") + mFail.Render("  Model: qwen3-8b. Escalated to claude-opus."))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    |"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("    '-- ") + mActv.Render("o Path B: Hull-first protocol ................ NEW BRANCH"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("         |-- ") + mActv.Render("[3] Seal hull breaches ..................... IN PROGRESS"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("         |-- ") + mPend.Render("[4] Stabilize power grid ................... PENDING"))
	b.WriteString("\n")
	b.WriteString(mGrBdr.Render("         '-- ") + mPend.Render("[5] Repair engine (safe conditions) ........ PENDING"))
	b.WriteString("\n\n")
}

func (m MissionModel) renderBundleMetadata(b *strings.Builder, bundle *domain.ContextResult, simulated bool) {
	graphTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("183"))
	b.WriteString(graphTitle.Render("  --- REHYDRATION BUNDLE ---"))
	b.WriteString("\n\n")

	if !simulated && bundle != nil {
		b.WriteString(mBndl.Render("   Root:       ") + bundle.RootNodeID + "\n")
		b.WriteString(mBndl.Render("   Nodes:      ") + fmt.Sprintf("%d    Relationships: %d", len(bundle.Nodes), len(bundle.Relationships)) + "\n")
		b.WriteString(mBndl.Render("   Details:    ") + fmt.Sprintf("%d node details loaded", len(bundle.Details)) + "\n")
		b.WriteString(mBndl.Render("   Tokens:     ") + fmt.Sprintf("%d / 4,000 budget", bundle.TokenCount) + "\n")
		b.WriteString(mBndl.Render("   Hash:       ") + bundle.ContentHash + "\n")
		b.WriteString(mBndl.Render("   Snapshot:   ") + bundle.SnapshotID + "\n")
	} else {
		b.WriteString(mBndl.Render("   Root:       ") + "node:mission:engine-core-failure\n")
		b.WriteString(mBndl.Render("   Role:       ") + "implementer\n")
		b.WriteString(mBndl.Render("   Checkpoint: ") + "ALPHA-3 (damage assessment)\n")
		b.WriteString(mBndl.Render("   Nodes:      ") + "7    Relationships: 6\n")
		b.WriteString(mBndl.Render("   Details:    ") + "3 node details loaded\n")
		b.WriteString(mBndl.Render("   Tokens:     ") + "394 / 4,000 budget\n")
		b.WriteString(mBndl.Render("   Snapshot:   ") + "snap_uss_20260312T154230Z\n")
	}
	b.WriteString("\n")

	tokenCount := uint32(394)
	if bundle != nil {
		tokenCount = bundle.TokenCount
	}

	tokenNote := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true)
	b.WriteString(tokenNote.Render(fmt.Sprintf("   %d tokens. Not 6,190. Not 128,000. %s.",
		tokenCount, numberToWords(tokenCount))))
	b.WriteString("\n")
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func writeBanner(b *strings.Builder, color, status, title string) {
	b.WriteString(mBanner.Foreground(lipgloss.Color(color)).Render(
		fmt.Sprintf("  [%s] %s", status, title)))
	b.WriteString("\n\n")
}

func writeNarr(b *strings.Builder, lines ...string) {
	for _, l := range lines {
		b.WriteString("  " + mNarr.Render(l) + "\n")
	}
	b.WriteString("\n")
}

func writeHint(b *strings.Builder, text string) {
	b.WriteString("\n  " + mHint.Render(text))
}

func writeEvent(b *strings.Builder, subject, agent string) {
	b.WriteString("  " + mEvt.Render("EVENT") + " ")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Render(subject))
	b.WriteString(mPend.Render(" -> "))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Bold(true).Render(agent))
	b.WriteString("\n\n")
}

func writeAgentLog(b *strings.Builder, header string, lines ...string) {
	b.WriteString("  " + mAgent.Render(header+":") + "\n")
	for _, l := range lines {
		b.WriteString(mLog.Render("  > "+l) + "\n")
	}
	b.WriteString("\n")
}

func (m MissionModel) renderSystems(overrides map[string]sysOverride, maxErrRate float64) string {
	var b strings.Builder
	sep := "  " + strings.Repeat("-", 72)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147"))
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s %12s %10s %8s %10s",
		"System", "Confidence", "Err Rate", "P95 ms", "Status")))
	b.WriteString("\n" + sep + "\n")

	for _, p := range m.policies {
		errRate := p.ErrorRate
		p95 := p.P95LatencyMs
		if o, ok := overrides[p.ToolID]; ok {
			errRate = o.errRate
			p95 = o.p95ms
		}

		// Status
		statusText := "OK"
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
		if maxErrRate > 0 && errRate > maxErrRate {
			statusText = "FILTERED"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true)
		} else if errRate > 0.15 {
			statusText = "DEGRADED"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
		}

		// Confidence color
		confStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
		if errRate > 0.20 {
			confStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("210"))
		} else if errRate > 0.10 {
			confStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
		}

		b.WriteString(fmt.Sprintf("  %-16s %s %9.1f%% %7dms %10s\n",
			p.ToolID,
			confStyle.Render(fmt.Sprintf("%11.1f%%", p.Confidence*100)),
			errRate*100,
			p95,
			statusStyle.Render(statusText)))
	}
	b.WriteString(sep)
	return b.String()
}

func numberToWords(n uint32) string {
	if n == 394 {
		return "Three hundred ninety-four"
	}
	return fmt.Sprintf("%d", n)
}
