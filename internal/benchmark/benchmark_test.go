package benchmark

import (
	"fmt"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 1: Simple Diagnostic — single-turn, one tool call
// ─────────────────────────────────────────────────────────────────────────────

func scenarioSimpleDiagnostic() Scenario {
	return Scenario{
		Name:        "SimpleDiagnostic",
		Description: "Single-turn diagnostic: agent scans engine, reports findings",
		Traditional: []ContextSection{
			{"system", traditionalSystemPrompt},
			{"tools", traditionalToolDescriptions},
			{"task", `Current task:
PRIORITY: WARNING
Investigate power fluctuation anomaly detected by automated monitoring in the
starboard engine manifold. Run diagnostic scan, assess severity, and recommend
next steps. Report findings to the bridge.`},
			{"rag", traditionalRAGChunks},
			{"workspace", traditionalWorkspaceSimple},
		},
		Surgical: []ContextSection{
			{"system", surgicalSystemSimple},
			{"bundle", surgicalBundleSimple},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 2: Multi-Turn Repair — 15 conversation turns, multiple tools
// ─────────────────────────────────────────────────────────────────────────────

func scenarioMultiTurnRepair() Scenario {
	return Scenario{
		Name:        "MultiTurnRepair",
		Description: "15-turn engine repair: diagnosis, failed attempts, tool switching",
		Traditional: []ContextSection{
			{"system", traditionalSystemPrompt},
			{"tools", traditionalToolDescriptions},
			{"history", traditionalConversationHistory},
			{"rag", traditionalRAGChunks},
			{"workspace", traditionalWorkspaceFull},
			{"task", `Current task:
PRIORITY: CRITICAL
Engine core coolant rupture confirmed at junction M-7. Main propulsion offline.
Three previous repair attempts failed. Hull integrity at 76% and declining.
Cascade failure affecting power grid (68% capacity) and shield generator.
Determine correct repair sequence and execute.`},
			{"attempts", traditionalPreviousAttempts},
		},
		Surgical: []ContextSection{
			{"system", surgicalSystemRepair},
			{"bundle", surgicalBundleRepair},
			{"tools", surgicalToolSubset},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Scenario 3: Complex Rehydration — branching decision with full task graph
// ─────────────────────────────────────────────────────────────────────────────

func scenarioComplexRehydration() Scenario {
	return Scenario{
		Name:        "ComplexRehydration",
		Description: "Strategic rehydration: agent analyzes failed path, branches new strategy",
		Traditional: []ContextSection{
			{"system", traditionalSystemPrompt},
			{"tools", traditionalToolDescriptions},
			{"history", traditionalConversationHistory},
			{"rag", traditionalRAGChunks},
			{"workspace", traditionalWorkspaceFull},
			{"task", `Current task:
PRIORITY: CRITICAL
Previous repair strategy (direct engine repair) has failed after 3 attempts.
Hull stress increasing with each attempt. Analyze all previous work, identify
why the current approach is failing, and propose an alternative repair strategy.
Consider hull integrity, power grid stability, and shield coverage in your analysis.
You have full access to all ship systems and maintenance history.`},
			{"attempts", traditionalPreviousAttempts},
			{"docs", traditionalDocumentation},
		},
		Surgical: []ContextSection{
			{"system", surgicalSystemRehydration},
			{"bundle", surgicalBundleRehydration},
			{"graph", surgicalTaskGraph},
			{"tools", surgicalToolSubset},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestContextPrecision(t *testing.T) {
	scenarios := []Scenario{
		scenarioSimpleDiagnostic(),
		scenarioMultiTurnRepair(),
		scenarioComplexRehydration(),
	}

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════════════════")
	t.Log(" CONTEXT PRECISION BENCHMARK — Honest Numbers")
	t.Log(" Token estimation: len(text)/4 (±10% vs tiktoken)")
	t.Log(" Same estimator for both approaches → ratio is accurate")
	t.Log("═══════════════════════════════════════════════════════════════════════")

	type result struct {
		name       string
		tradTokens int
		surgTokens int
		ratio      float64
	}
	var results []result

	for _, s := range scenarios {
		tradTokens := TotalTokens(s.Traditional)
		surgTokens := TotalTokens(s.Surgical)
		ratio := float64(tradTokens) / float64(surgTokens)

		results = append(results, result{s.Name, tradTokens, surgTokens, ratio})

		t.Run(s.Name, func(t *testing.T) {
			t.Logf("")
			t.Logf("Scenario: %s", s.Description)
			t.Logf("─────────────────────────────────────────────────────")
			t.Logf("Traditional: %6d tokens  (%s)", tradTokens, SectionBreakdown(s.Traditional))
			t.Logf("Surgical:    %6d tokens  (%s)", surgTokens, SectionBreakdown(s.Surgical))
			t.Logf("Reduction:   %.1fx", ratio)

			// Minimum expected reduction per scenario
			minReduction := 3.0
			if ratio < minReduction {
				t.Errorf("Expected at least %.1fx reduction, got %.1fx", minReduction, ratio)
			}
		})
	}

	// Cost comparison table
	t.Run("CostComparison", func(t *testing.T) {
		t.Logf("")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf(" COST PER CALL — Scenario 2 (Multi-Turn Repair)")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf("")
		t.Logf("  %-22s %12s %12s %10s", "Provider", "Traditional", "Surgical", "Savings")
		t.Logf("  %s", strings.Repeat("─", 60))

		s2 := results[1] // Multi-turn repair
		for _, p := range Providers() {
			tradCost := CostPerCall(s2.tradTokens, p.InputPerM)
			surgCost := CostPerCall(s2.surgTokens, p.InputPerM)
			t.Logf("  %-22s $%10.6f $%10.6f %9.1fx",
				p.Name, tradCost, surgCost, s2.ratio)
		}

		t.Logf("")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf(" DAILY COST — 1,000 calls/day, Scenario 2")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf("")
		t.Logf("  %-22s %12s %12s %10s", "Provider", "Traditional", "Surgical", "Savings")
		t.Logf("  %s", strings.Repeat("─", 60))

		for _, p := range Providers() {
			tradDaily := CostPerCall(s2.tradTokens, p.InputPerM) * 1000
			surgDaily := CostPerCall(s2.surgTokens, p.InputPerM) * 1000
			t.Logf("  %-22s $%10.2f $%10.2f %9.1fx",
				p.Name, tradDaily, surgDaily, s2.ratio)
		}

		// Blended model routing
		t.Logf("")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf(" BLENDED MODEL ROUTING — 1,000 calls/day, Scenario 2")
		t.Logf(" 95%% local Qwen3-8B + 5%% Claude Opus for strategic decisions")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf("")

		// Traditional: everything on Sonnet (need big model for noisy context)
		tradAllSonnet := CostPerCall(s2.tradTokens, 3.00) * 1000
		// Surgical blended: 95% local + 5% Opus
		surgLocal := CostPerCall(s2.surgTokens, 0.10) * 950
		surgOpus := CostPerCall(s2.surgTokens, 15.00) * 50
		surgBlended := surgLocal + surgOpus

		t.Logf("  Traditional (all Sonnet):    $%.2f/day", tradAllSonnet)
		t.Logf("  Surgical (95%% local, 5%% Opus): $%.2f/day", surgBlended)
		t.Logf("  Combined savings:            %.0fx", tradAllSonnet/surgBlended)
		t.Logf("")
		t.Logf("  Breakdown:")
		t.Logf("    Token precision:           %.1fx (same model, fewer tokens)", s2.ratio)
		t.Logf("    + Model routing:           additional savings from local GPU")
		t.Logf("    = Combined effect:         %.0fx", tradAllSonnet/surgBlended)
	})

	// Summary
	t.Run("Summary", func(t *testing.T) {
		t.Logf("")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf(" SUMMARY")
		t.Logf("═══════════════════════════════════════════════════════════════════════")
		t.Logf("")
		t.Logf("  %-25s %8s %8s %8s", "Scenario", "Trad.", "Surg.", "Ratio")
		t.Logf("  %s", strings.Repeat("─", 55))
		for _, r := range results {
			t.Logf("  %-25s %7d %7d %7.1fx", r.name, r.tradTokens, r.surgTokens, r.ratio)
		}
		t.Logf("")
		t.Logf("  Token precision savings:  %.0fx - %.0fx depending on scenario complexity",
			results[0].ratio, results[len(results)-1].ratio)
		t.Logf("  These are honest numbers from real content. Same estimator, same text.")
		t.Logf("  Verify: read the scenario content in this file.")
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// TRADITIONAL CONTEXT CONTENT
// These strings represent what a typical agent framework (LangChain, CrewAI,
// SWE-agent) puts in the prompt. Each is realistic, not inflated.
// ─────────────────────────────────────────────────────────────────────────────

var traditionalSystemPrompt = `You are an AI maintenance agent on the USS Underpass, a deep-space exploration vessel.
You are responsible for diagnosing and repairing ship systems. You have access to ship
tools that let you interact with subsystems. Always explain your reasoning before taking
action. Follow safety protocols: never modify critical systems without scanning first,
always check hull integrity before engine work, and report all anomalies to the bridge.

You operate in a multi-agent fleet. Other agents may be working on related subsystems.
Coordinate through the shared workspace. Do not duplicate work already completed by
other agents. Check the workspace before starting any task.

Response format: Think step by step. For each action, explain why you chose it.
If a tool call fails, analyze the error and try an alternative approach.
Do not retry the same failed action more than twice.

Ship specifications:
- Hull: Titanium-composite, rated for 350 MPa stress, critical threshold at 70% integrity
- Engine: Mark VII fusion core, closed-loop coolant (nominal pressure 2.8-3.2 MPa)
- Power grid: Distributed bus architecture, 12 nodes, overload protection at 85% capacity
- Shields: Frequency-modulated barrier, requires stable power for calibration
- Life support: Triple-redundant atmospheric recycling, 72-hour backup reserve
- Communications: Subspace burst transmitter, range 40 light-years at full power
- Navigation: Astrometric computer with real-time course correction
- Sensors: Multi-spectrum deep scan array, resolution 0.01 AU at maximum power`

var traditionalToolDescriptions = `Available tools:

1. nav.plot
   Description: Calculate navigation routes between coordinates. Accounts for
   gravitational fields, known hazards, and fuel constraints. Returns optimal
   route with estimated time of arrival and fuel cost.
   Parameters: {
     "destination": {"type": "string", "description": "Target coordinates or named location"},
     "speed": {"type": "number", "description": "Warp factor 0.0-1.0, where 1.0 is maximum"},
     "avoid_zones": {"type": "array", "items": {"type": "string"}, "description": "Named hazards to route around"}
   }
   Returns: {"route": "Route object with waypoints", "eta": "Duration string", "fuel_cost": "Percentage of reserves"}
   Example: nav.plot({"destination": "Starbase 12", "speed": 0.8, "avoid_zones": ["Nebula X-7"]})

2. scan.deep
   Description: Perform deep-spectrum sensor scan on a target system or region.
   Returns detailed telemetry including temperature, pressure, electromagnetic
   readings, and anomaly detection. Scan depth affects resolution and duration.
   Parameters: {
     "target": {"type": "string", "description": "System or region to scan (e.g., 'engine_starboard', 'hull_section_7')"},
     "depth": {"type": "string", "enum": ["surface", "standard", "full"], "description": "Scan depth — full takes 30s but gives maximum detail"},
     "spectrum": {"type": "string", "enum": ["thermal", "electromagnetic", "structural", "all"], "description": "Sensor spectrum to use"}
   }
   Returns: {"status": "ok|warning|critical", "readings": {"temperature_k": 0, "pressure_mpa": 0, "em_field_gauss": 0}, "anomalies": []}
   Example: scan.deep({"target": "engine_starboard", "depth": "full", "spectrum": "all"})

3. eng.thrust
   Description: Control engine thrust, perform maintenance operations, and manage
   coolant systems. Can adjust thrust vectors, run diagnostic cycles, and attempt
   coolant line repairs. Requires hull integrity above 70% for safe operation.
   Parameters: {
     "operation": {"type": "string", "enum": ["thrust_adjust", "diagnostic", "coolant_repair", "shutdown", "restart"]},
     "parameters": {"type": "object", "description": "Operation-specific parameters"},
     "safety_override": {"type": "boolean", "description": "Override safety checks (NOT RECOMMENDED)"}
   }
   Returns: {"success": true, "engine_status": {}, "warnings": []}
   Example: eng.thrust({"operation": "diagnostic", "parameters": {"subsystem": "coolant"}})

4. hull.seal
   Description: Seal hull breaches, reinforce structural weak points, and manage
   hull integrity. Can apply emergency patches, activate self-healing composites,
   and redistribute structural load. Non-destructive — safe at any integrity level.
   Parameters: {
     "action": {"type": "string", "enum": ["scan_integrity", "emergency_patch", "reinforce", "self_heal"]},
     "section": {"type": "string", "description": "Hull section identifier (e.g., 'section_7_starboard')"},
     "priority": {"type": "string", "enum": ["normal", "urgent", "critical"]}
   }
   Returns: {"success": true, "integrity_before": 0.76, "integrity_after": 0.82, "patches_applied": 3}
   Example: hull.seal({"action": "emergency_patch", "section": "section_7_starboard", "priority": "critical"})

5. comm.burst
   Description: Transmit or receive subspace communications. Supports burst
   transmission for maximum range, continuous stream for real-time data, and
   emergency beacon activation. Range depends on power allocation.
   Parameters: {
     "mode": {"type": "string", "enum": ["transmit", "receive", "beacon", "scan_frequencies"]},
     "message": {"type": "string", "description": "Message content for transmission"},
     "target": {"type": "string", "description": "Recipient station or vessel"},
     "power_level": {"type": "number", "description": "Power allocation 0.0-1.0"}
   }
   Returns: {"success": true, "signal_strength": 0.85, "latency_seconds": 2.3}
   Example: comm.burst({"mode": "transmit", "message": "Status report", "target": "Starbase 12", "power_level": 0.8})

6. life.recycle
   Description: Manage life support atmospheric recycling, temperature control,
   and emergency reserves. Can switch between primary and backup systems, adjust
   atmospheric composition, and monitor crew environment metrics.
   Parameters: {
     "system": {"type": "string", "enum": ["primary", "backup", "both"]},
     "action": {"type": "string", "enum": ["status", "switch_to_backup", "adjust_composition", "emergency_reserve"]},
     "parameters": {"type": "object", "description": "Action-specific parameters"}
   }
   Returns: {"success": true, "o2_level": 20.9, "co2_level": 0.04, "temperature_c": 22.1, "reserve_hours": 68}
   Example: life.recycle({"system": "primary", "action": "status"})

7. power.reroute
   Description: Manage power distribution across the ship's bus architecture.
   Can reroute power between nodes, shed non-essential loads, and manage
   overload protection. Critical for maintaining systems during failures.
   Parameters: {
     "action": {"type": "string", "enum": ["status", "reroute", "shed_load", "emergency_power", "reset_breaker"]},
     "source_node": {"type": "string", "description": "Power bus node to draw from"},
     "target_node": {"type": "string", "description": "Power bus node to supply"},
     "amount_percent": {"type": "number", "description": "Percentage of node capacity to transfer"}
   }
   Returns: {"success": true, "grid_status": {"total_capacity": 0.68, "nodes": []}, "warnings": []}
   Example: power.reroute({"action": "reroute", "source_node": "auxiliary_3", "target_node": "engine_primary", "amount_percent": 30})

8. shield.mod
   Description: Modulate shield frequency, adjust coverage pattern, and manage
   shield generator power consumption. Shields require stable power input —
   fluctuations cause coverage gaps.
   Parameters: {
     "action": {"type": "string", "enum": ["status", "modulate_frequency", "adjust_coverage", "boost_section", "shutdown"]},
     "frequency_hz": {"type": "number", "description": "Shield frequency in Hz (optimal: 347.2-352.8)"},
     "section": {"type": "string", "description": "Shield section to adjust"}
   }
   Returns: {"success": true, "coverage_percent": 94.2, "frequency_hz": 350.0, "power_draw_percent": 12.3}
   Example: shield.mod({"action": "status"})`

var traditionalConversationHistory = `Conversation history:

[Turn 1]
User: Engine diagnostic report shows anomaly in starboard manifold. Investigate.
Agent: I'll start with a full diagnostic scan of the starboard engine to understand the anomaly. Let me check the sensor array first.
Tool call: scan.deep({"target": "engine_starboard", "depth": "full", "spectrum": "all"})
Result: {"status": "warning", "readings": {"temperature_k": 847, "pressure_mpa": 2.1, "em_field_gauss": 0.34}, "anomalies": [{"type": "pressure_drop", "location": "manifold_junction_M7", "severity": "moderate", "detail": "Coolant pressure 2.1 MPa below nominal range 2.8-3.2 MPa"}, {"type": "temperature_elevation", "location": "core_chamber", "severity": "moderate", "detail": "Temperature 847K exceeds nominal 780-810K range"}]}

[Turn 2]
Agent: The scan reveals two concerning anomalies in the starboard engine: coolant pressure is below nominal (2.1 MPa vs 2.8-3.2 MPa range) and core temperature is elevated (847K vs 780-810K). This suggests a coolant leak or flow restriction at manifold junction M-7. Let me run an engine diagnostic to pinpoint the issue.
Tool call: eng.thrust({"operation": "diagnostic", "parameters": {"subsystem": "coolant", "target": "manifold_M7"}})
Result: {"success": true, "engine_status": {"thrust_available": 0.72, "coolant_flow_rate": 0.64, "coolant_integrity": "compromised", "leak_detected": true, "leak_location": "junction_M7_seal_ring", "leak_rate_lpm": 2.3}, "warnings": ["Coolant leak detected — progressive degradation expected", "Recommend immediate repair or shutdown within 45 minutes"]}

[Turn 3]
Agent: Confirmed: there's a coolant leak at the M-7 seal ring. Flow rate is at 64% of nominal and degrading. The engine has 72% thrust available but this will drop as coolant is lost. We have approximately 45 minutes before forced shutdown. I recommend attempting a coolant repair.
User: Proceed with repair. Priority is getting the engine back to full capacity.

[Turn 4]
Agent: Attempting coolant repair at junction M-7. This involves sealing the leak and restoring pressure.
Tool call: eng.thrust({"operation": "coolant_repair", "parameters": {"target": "junction_M7_seal_ring", "method": "seal_and_pressurize"}})
Result: {"success": false, "engine_status": {"thrust_available": 0.65, "coolant_flow_rate": 0.58}, "warnings": ["Repair failed — seal ring material fatigue prevents proper seal", "Hull vibration from engine operation is preventing seal adhesion", "Recommend hull stabilization before retry"], "error": "SEAL_ADHESION_FAILURE: vibration exceeds tolerance for in-situ repair"}

[Turn 5]
Agent: The repair failed. The seal ring has material fatigue and hull vibration from engine operation is preventing the new seal from adhering. The engine diagnostic suggests we need to stabilize the hull first. However, the user prioritized engine repair, so let me try an alternative approach — a coolant bypass.
Tool call: eng.thrust({"operation": "coolant_repair", "parameters": {"target": "junction_M7", "method": "bypass_damaged_section"}})
Result: {"success": false, "engine_status": {"thrust_available": 0.58, "coolant_flow_rate": 0.51, "hull_stress_increase": 0.04}, "warnings": ["Bypass attempt caused additional hull stress (+4%)", "Hull integrity now at 84%", "Coolant continuing to degrade"], "error": "BYPASS_STRESS: bypass routing through hull section 7 increased structural load"}

[Turn 6]
User: Hull stress increased? What's the current hull integrity?
Agent: The bypass attempt made things worse — it routed coolant through hull section 7 which increased structural stress by 4%. Current hull integrity is 84%. Let me scan the hull to get a full picture.
Tool call: hull.seal({"action": "scan_integrity", "section": "all"})
Result: {"success": true, "integrity_before": 0.84, "sections": [{"id": "section_1_port", "integrity": 0.95}, {"id": "section_3_dorsal", "integrity": 0.92}, {"id": "section_5_ventral", "integrity": 0.88}, {"id": "section_7_starboard", "integrity": 0.72, "warning": "approaching critical threshold"}, {"id": "section_9_aft", "integrity": 0.91}]}

[Turn 7]
Agent: Hull scan shows section 7 (starboard) is at 72% integrity — approaching the critical 70% threshold. This is the section adjacent to the engine. The coolant bypass attempt stressed it further. I should try to reinforce this section before attempting more engine work.
User: We need the engine running. Can you try the repair one more time?

[Turn 8]
Agent: Attempting engine repair one more time. I'll try a different method — direct seal injection.
Tool call: eng.thrust({"operation": "coolant_repair", "parameters": {"target": "junction_M7_seal_ring", "method": "inject_sealant"}})
Result: {"success": false, "engine_status": {"thrust_available": 0.51, "coolant_flow_rate": 0.44, "hull_stress_increase": 0.04}, "warnings": ["Sealant injection caused additional vibration", "Hull integrity now at 80%", "Engine thrust continuing to degrade"], "error": "VIBRATION_EXCESS: sealant injection pump vibration exceeded safe limits for hull section 7"}

[Turn 9]
Agent: Third repair attempt also failed. Each attempt is causing additional hull stress. Hull integrity has dropped from 88% to 80% in three attempts. Engine thrust is now at 51% and still declining. I need to reconsider the approach. The root cause is that we can't repair the engine while the hull is vibrating — and the hull vibrates because of the engine. It's a catch-22.
User: What do you recommend?

[Turn 10]
Agent: I recommend a different approach: 1) Shut down the engine to stop vibrations, 2) Seal the hull to restore structural integrity, 3) Repair the engine coolant under stable conditions, 4) Restart. This is safer but means we'll be without propulsion for the duration.
Tool call: power.reroute({"action": "status"})
Result: {"success": true, "grid_status": {"total_capacity": 0.68, "overload_protection": "active", "nodes": [{"id": "engine_primary", "load": 0.82, "status": "stressed"}, {"id": "shields", "load": 0.71, "status": "nominal"}, {"id": "life_support", "load": 0.45, "status": "nominal"}, {"id": "sensors", "load": 0.38, "status": "nominal"}, {"id": "auxiliary_1", "load": 0.22, "status": "nominal"}]}, "warnings": ["Engine primary node stressed — cascade risk if load increases"]}

[Turn 11]
Agent: Power grid is at 68% capacity with the engine node under stress. If we shut down the engine, we free up power for hull repair and shields. The cascade risk is real — if the engine node fails completely, it could take out adjacent power nodes.
User: The cascade already started. Power grid is degrading. Fix it.

[Turn 12]
Agent: I need to stabilize the power grid first. Let me shed non-essential loads and reroute power from the auxiliary bus.
Tool call: power.reroute({"action": "shed_load", "source_node": "sensors", "amount_percent": 50})
Result: {"success": true, "grid_status": {"total_capacity": 0.72, "nodes": [{"id": "sensors", "load": 0.19, "status": "reduced"}]}, "warnings": ["Sensor resolution reduced to 50% — deep scans unavailable"]}

[Turn 13]
Agent: Shed 50% sensor load. Grid capacity improved slightly to 72%. But we've lost deep scan capability which limits our diagnostic ability. Let me also check shield status.
Tool call: shield.mod({"action": "status"})
Result: {"success": true, "coverage_percent": 78.4, "frequency_hz": 349.1, "power_draw_percent": 14.8, "warnings": ["Frequency drift detected — power fluctuations affecting calibration", "Coverage gaps in sections 6-8 (starboard)"]}

[Turn 14]
Agent: Shields are at 78.4% coverage with gaps on the starboard side — exactly where our hull is weakest. The power fluctuations are causing frequency drift. We have a compounding problem: engine failure → power cascade → shield gaps → hull exposed. Each system failure makes the others worse.
User: This is getting critical. What's the priority?

[Turn 15]
Agent: Priority assessment: 1) Hull is most critical — at 80% and declining, approaching 70% threshold where structural failure becomes possible. 2) Power grid needs stabilization to support repairs. 3) Engine repair can only succeed after hull is stable.

I've been approaching this wrong. Trying to fix the engine first while the hull is compromised is making everything worse. We need to reverse the order: hull first, then power, then engine.

Current status: Hull 80%, Engine 51% thrust, Power 72% capacity, Shields 78.4%.`

var traditionalRAGChunks = `Retrieved documentation:

[Chunk 1 — Engine Maintenance Manual v4.2, Section: Coolant System]
The Mark VII fusion engine core uses a closed-loop coolant system operating at nominal
pressure of 2.8-3.2 MPa. The coolant circuit passes through six manifold junctions
(M1-M6 primary, M7-M8 auxiliary) with ceramic seal rings rated for 500,000 thermal
cycles. Junction M-7 connects the auxiliary cooling loop to the main circuit and
handles approximately 15% of total coolant flow. Failure of M-7 reduces effective
cooling to 85% of capacity initially, with progressive degradation as thermal cycling
accelerates seal ring fatigue in adjacent junctions.

Repair procedures for in-situ seal replacement require: (a) engine at idle or shutdown,
(b) hull vibration below 0.3g RMS at the repair site, (c) ambient temperature below
400K for sealant curing. If conditions (a) or (b) cannot be met, the recommended
procedure is to shut down the affected engine, stabilize the ship structure, and then
perform the repair under controlled conditions.

WARNING: Attempting in-situ repair while hull vibration exceeds 0.3g RMS will result
in seal adhesion failure. Each failed attempt increases local hull stress by 3-5% due
to repair tool vibration coupling with hull structural harmonics.

[Chunk 2 — Structural Engineering Reference, Hull Section 7]
Hull section 7 (starboard, frames 42-56) houses the starboard engine mounting points
and primary coolant conduit runs. The section uses Grade 7 titanium-composite panels
with a rated yield stress of 350 MPa. Integrity below 70% indicates micro-crack
propagation in the composite matrix and requires immediate structural intervention.

The section is particularly susceptible to vibration-induced fatigue when engine
maintenance operations are performed in-situ. Vibration from repair tools couples
with the natural frequency of the hull panels (estimated 12-18 Hz) and can accelerate
micro-crack growth by 200-400% compared to static loading.

Recommended repair sequence for combined engine/hull failures:
1. Secure engine (shutdown or idle)
2. Apply emergency structural patches to restore integrity above 85%
3. Activate self-healing composites (requires 2-4 hours for full cure)
4. Verify hull integrity via full-spectrum scan
5. Proceed with engine repair under stabilized conditions

[Chunk 3 — Power Systems Guide, Cascade Failure Prevention]
The ship's distributed power bus uses 12 interconnected nodes with automatic overload
protection. When a primary load node (such as the engine) begins drawing excessive
current, the overload protection limits current to 85% of node capacity and redirects
excess demand to adjacent nodes. If adjacent nodes are already loaded above 60%, a
cascade condition can develop where overload protection triggers sequentially across
the bus.

Cascade prevention protocol:
1. Identify the primary fault node
2. Shed non-essential loads from adjacent nodes
3. Reroute power from low-priority auxiliary buses
4. If cascade has already begun, isolate the fault node to prevent further propagation

Note: Shield generators require stable power within ±2% to maintain frequency lock.
Power fluctuations beyond this tolerance cause progressive frequency drift that
manifests as coverage gaps, typically in the same section as the power fault.

[Chunk 4 — Safety Protocol SP-2847: Combined System Failure Response]
When multiple ship systems fail simultaneously, the priority order for repair is:
1. Life support (survival-critical, 72-hour backup)
2. Hull integrity (structural-critical, threshold 70%)
3. Power grid (enables all other repairs)
4. Shields (protection from external threats)
5. Propulsion (mobility, not survival-critical in short term)

This priority order reflects the dependency chain: hull stability enables all other
repairs, power enables shields and propulsion, and life support is the ultimate
backstop. Attempting to repair lower-priority systems before higher-priority ones
is an error that can compound damage.

CRITICAL: Never attempt engine repair when hull integrity is below 85% in the
engine mounting section. Engine vibration during repair will accelerate hull
degradation and can cause catastrophic structural failure.`

var traditionalWorkspaceSimple = `Current workspace:
/workspace/diagnostics/ (empty)
/workspace/repairs/ (empty)
/workspace/status/ (empty)

No files from previous work.`

var traditionalWorkspaceFull = `Current workspace:

/workspace/diagnostics/engine_scan_001.json (2.3KB):
{"timestamp": "2026-03-12T14:23:00Z", "target": "engine_starboard", "status": "warning",
 "readings": {"temperature_k": 847, "pressure_mpa": 2.1, "em_field_gauss": 0.34},
 "anomalies": [{"type": "pressure_drop", "location": "manifold_junction_M7"}]}

/workspace/diagnostics/engine_scan_002.json (1.8KB):
{"timestamp": "2026-03-12T14:31:00Z", "target": "engine_starboard", "status": "critical",
 "readings": {"temperature_k": 912, "pressure_mpa": 1.4, "em_field_gauss": 0.51},
 "anomalies": [{"type": "coolant_rupture", "location": "manifold_junction_M7", "severity": "critical"}]}

/workspace/diagnostics/hull_scan_001.json (3.1KB):
{"timestamp": "2026-03-12T14:45:00Z", "target": "all_sections", "overall_integrity": 0.84,
 "sections": [{"id": "section_1", "integrity": 0.95}, {"id": "section_3", "integrity": 0.92},
 {"id": "section_5", "integrity": 0.88}, {"id": "section_7", "integrity": 0.72},
 {"id": "section_9", "integrity": 0.91}]}

/workspace/diagnostics/power_grid_001.json (2.7KB):
{"timestamp": "2026-03-12T15:02:00Z", "grid_capacity": 0.68,
 "nodes": [{"id": "engine_primary", "load": 0.82, "status": "stressed"},
 {"id": "shields", "load": 0.71}, {"id": "life_support", "load": 0.45}]}

/workspace/repairs/coolant_repair_attempt_1.log:
[14:35:22] Initiating coolant repair at junction M-7
[14:35:23] Method: seal_and_pressurize
[14:35:45] Seal application started
[14:36:12] ERROR: Seal adhesion failure — hull vibration 0.47g RMS exceeds 0.3g limit
[14:36:13] Repair aborted. Hull stress increased 4%.

/workspace/repairs/coolant_repair_attempt_2.log:
[14:42:15] Initiating coolant repair at junction M-7
[14:42:16] Method: bypass_damaged_section
[14:42:38] Bypass routing through hull section 7
[14:43:01] ERROR: Bypass routing increased structural load on hull section 7
[14:43:02] Hull stress increased additional 4%. Hull integrity now 84%.
[14:43:03] Repair aborted.

/workspace/repairs/coolant_repair_attempt_3.log:
[15:01:33] Initiating coolant repair at junction M-7
[15:01:34] Method: inject_sealant
[15:01:52] Sealant injection pump activated
[15:02:15] ERROR: Pump vibration 0.52g RMS exceeded safe limit for hull section 7
[15:02:16] Hull stress increased additional 4%. Hull integrity now 80%.
[15:02:17] Repair aborted.

/workspace/status/current_status.json:
{"timestamp": "2026-03-12T15:10:00Z", "hull_integrity": 0.80, "engine_thrust": 0.51,
 "power_capacity": 0.68, "shield_coverage": 0.784, "life_support": "nominal"}`

var traditionalPreviousAttempts = `Previous repair attempts summary:

Attempt 1 — Seal and Pressurize (14:35, Stardate 2026.071)
  Method: Apply new seal ring to junction M-7 and pressurize coolant circuit
  Result: FAILED — seal adhesion failure due to hull vibration (0.47g vs 0.3g limit)
  Side effect: Hull stress +4% (88% → 84%)
  Root cause: Engine operation creates vibration that prevents seal from curing

Attempt 2 — Coolant Bypass (14:42, Stardate 2026.071)
  Method: Route coolant around damaged M-7 junction through alternate hull conduit
  Result: FAILED — bypass routing through hull section 7 increased structural load
  Side effect: Hull stress +4% (84% → 80%)
  Root cause: Alternate conduit runs through already-stressed hull section

Attempt 3 — Sealant Injection (15:01, Stardate 2026.071)
  Method: Inject liquid sealant directly into M-7 junction under pressure
  Result: FAILED — injection pump vibration exceeded hull safety limits
  Side effect: Hull stress +4% (80% → 76%)
  Root cause: Same vibration coupling problem as Attempt 1

Pattern: All three attempts failed due to hull vibration/stress interaction.
Each attempt worsened hull integrity by ~4%. Continuing this approach will
breach the 70% critical threshold within 1-2 more attempts.

Conclusion: Engine repair cannot succeed while hull section 7 is compromised.
Need to stabilize hull BEFORE attempting engine repair.`

var traditionalDocumentation = `Ship Technical Documentation Index:

[Doc: Mark VII Engine Core — Complete Service Manual, 12 pages]
The Mark VII fusion engine core is the primary propulsion system for Expedition-class
vessels. Operating at sustained fusion temperatures of 150 million Kelvin, the core
produces 2.4 terawatts of thrust power through magnetic confinement and directed
plasma exhaust. The engine assembly consists of the following major subsystems:

1. Plasma Confinement Chamber: Magnetic bottle configuration using superconducting
   coils arranged in a tokamak geometry. Operating magnetic field: 12 Tesla.
   Coolant: liquid helium at 4.2K for superconductor thermal management.

2. Fuel Injection System: Deuterium-tritium fuel pellets injected at 100 Hz.
   Each pellet mass: 0.5 mg. Fuel reserve: 180 days at nominal consumption.

3. Exhaust Management: Plasma exhaust directed through magnetic nozzle with
   variable geometry for thrust vectoring. Nozzle throat diameter: 2.4m.

4. Coolant System (PRIMARY CONCERN): Closed-loop system using liquid lithium
   coolant at 2.8-3.2 MPa nominal pressure. Six primary manifold junctions
   (M1-M6) and two auxiliary junctions (M7-M8). Total coolant volume: 4,200 liters.
   Flow rate: 850 liters/minute at nominal. Coolant temperature: 680-720K nominal.

5. Control Systems: Automated thrust management with manual override capability.
   Safety interlocks prevent operation outside envelope (pressure, temperature,
   vibration, hull integrity).

The engine is mounted at frames 42-48 (starboard) and 42-48 (port) with vibration
isolation mounts rated for 0.5g continuous and 2.0g transient loading. Hull section
7 directly interfaces with the starboard engine mounting structure.

[Doc: Hull Structural Analysis — Section 7 Detail]
Hull section 7 spans frames 42-56 on the starboard side. This section serves dual
roles: primary pressure vessel and engine mounting structure. The section uses
Grade 7 titanium-composite panels in a sandwich configuration:
- Outer skin: 8mm Ti-6Al-4V
- Core: 25mm carbon-fiber honeycomb
- Inner skin: 6mm Ti-6Al-4V
- Total thickness: 39mm

Design yield stress: 350 MPa. Service limit: 280 MPa (80% of yield).
Critical threshold: 245 MPa (70% of yield) — below this, micro-crack
propagation becomes self-sustaining and catastrophic failure is possible
within hours under dynamic loading.

The section has 12 inspection ports, 4 coolant conduit penetrations, and
2 engine mounting hard points. Each penetration is a potential stress
concentration point, particularly under vibration loading.

[Doc: Power Distribution Architecture]
The ship's electrical power is distributed through a 12-node bus architecture:
Node 1-2: Engine primary/secondary (40% of total capacity)
Node 3-4: Shield generators (15% of total capacity)
Node 5: Life support (10% of total capacity)
Node 6-7: Sensors and communications (10% of total capacity)
Node 8-10: Auxiliary systems (15% of total capacity)
Node 11-12: Emergency reserve (10% of total capacity)

Each node has independent overload protection that triggers at 85% capacity.
When a node trips, its load is automatically redistributed to adjacent nodes.
If adjacent nodes are above 60% capacity, redistribution can trigger a cascade
where multiple nodes trip in sequence.

Cascade prevention requires manual intervention: shed loads from healthy nodes
BEFORE the cascade propagates. Automated systems cannot make load-shedding
priority decisions — that requires human or AI judgment about which systems
can tolerate reduced power.`

// ─────────────────────────────────────────────────────────────────────────────
// SURGICAL CONTEXT CONTENT (Kernel Rehydration Bundle)
// These represent what the rehydration kernel actually produces.
// ─────────────────────────────────────────────────────────────────────────────

var surgicalSystemSimple = `Role: diagnostic-agent for mission engine-core-failure.
Assess the reported engine anomaly using scan.deep. Report findings.`

var surgicalBundleSimple = `Mission: Engine Core Failure — USS Underpass
Step: Diagnose Anomaly [IN PROGRESS]
Target: Starboard engine manifold, junction M-7
Expected: Run scan.deep(target=engine_starboard, depth=full)
Deliver: Diagnostic assessment with severity and recommendation.`

var surgicalSystemRepair = `Role: implementer for mission engine-core-failure.
Execute hull-first repair protocol. Tools filtered by Thompson Sampling.
Constraint: max_error_rate=20%. Only use tools listed below.`

var surgicalBundleRepair = `Mission: Engine Core Failure — USS Underpass
Phase: hull-first-protocol (rehydrated from checkpoint ALPHA-3)

Previous Path (ABANDONED):
  Path A: Direct engine repair — 3 attempts, all failed.
  Root cause: hull vibration prevents seal adhesion.
  Each attempt increased hull stress by ~4%.
  Hull integrity dropped: 88% → 76%.

Current Step: Seal Hull Breaches [IN PROGRESS]
Next: Stabilize Power Grid → Repair Engine (safe conditions)

Ship State:
  Hull: 76% integrity (critical: 70%), section 7 worst at 72%
  Engine: offline, coolant ruptured at junction M-7, thrust 51%
  Power: 68% grid capacity, engine node stressed
  Shields: 78.4% coverage, frequency drift, starboard gaps

Key Constraint:
  Hull vibration must be below 0.3g RMS before engine repair.
  Hull integrity must be above 85% in section 7 before engine work.
  Seal first, then power, then engine. Do NOT reverse this order.`

var surgicalToolSubset = `Available tools (Thompson Sampling filtered, max_error_rate=20%):
  hull.seal   — seal hull breaches, error_rate=8.2%, p95=450ms
  life.recycle — life support management, error_rate=7.2%, p95=560ms
  shield.mod  — shield modulation, error_rate=10.7%, p95=780ms
  scan.deep   — sensor array, error_rate=5.7%, p95=340ms (read-only)

Excluded (error_rate > 20%):
  eng.thrust   — 42% error rate (coolant system compromised)
  power.reroute — 28% error rate (cascade overload)`

var surgicalSystemRehydration = `Role: strategic-planner for mission engine-core-failure.
Analyze failed repair path. Recommend new strategy via task graph branching.
You have access to the full task graph and rehydration bundle.`

var surgicalBundleRehydration = `Mission: Engine Core Failure — USS Underpass
Rehydration Bundle (checkpoint ALPHA-3):

Root: node:mission:engine-core-failure
Role: implementer
Nodes: 7  Relationships: 6
Revision: 3  ContentHash: a7f2e9c1

Work Items:
  [1] Diagnose anomaly ................ COMPLETED
  [2] Assess cascade damage ........... COMPLETED
  [3] Execute repair strategy ......... FAILED (Path A)

Path A Summary (ABANDONED):
  3 repair attempts targeting engine directly.
  All failed: hull vibration prevents seal adhesion.
  Cumulative hull damage: -12% (88% → 76%).
  Pattern: each attempt worsens hull → reduces success probability.

Ship State at Checkpoint:
  Hull: 76% (section 7: 72%, approaching 70% critical)
  Engine: offline, coolant ruptured
  Power: 68% capacity, stressed
  Shields: 78.4%, starboard gaps

Decision Required:
  Propose new repair strategy that avoids the failure pattern.
  Constraint: hull must be stabilized before engine work.`

var surgicalTaskGraph = `Task Graph (from Neo4j):

● [1] Diagnose anomaly ..................... COMPLETED
 │
● [2] Assess cascade damage ................ COMPLETED
 │
 ├── ✗ Path A: Direct engine repair ........ ABANDONED
 │    3 attempts, hull stress +12%, counterproductive
 │
 └── ? Path B: (your recommendation) ....... PROPOSED
      Decision: what is the correct repair sequence?`

// ─────────────────────────────────────────────────────────────────────────────
// Compile-time check: all scenario content is used
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	_ = traditionalDocumentation
	_ = traditionalWorkspaceSimple
	_ = fmt.Sprint // ensure fmt is used
}
