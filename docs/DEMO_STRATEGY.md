# USS Underpass — Demo Strategy

## The Pitch (30 seconds)

> "What if your AI agents could **learn** which tools work, **react** to problems instantly,
> reconstruct **exactly** the context they need from a graph — in 3K tokens, not 128K —
> and **recover** from mistakes by rolling back to a previous decision point?"
>
> That's what we built. And we're going to show it with a starship that breaks down.

---

## The Four Pillars

The demo showcases four capabilities that, **individually**, are impressive.
**Together**, they create something no one else is showing: a fully autonomous, self-healing AI fleet.

### Pillar 1 — Thompson Sampling (Tool-Learning)

**What**: AI agents learn which tools work best through Bayesian exploration/exploitation.
Beta(alpha, beta) priors track success/failure per tool. Hard constraints filter dangerous tools.

**Why it matters**: Every AI framework gives agents tools. *Nobody teaches agents which tools
to trust.* When a tool degrades (higher latency, more errors), the system adapts automatically.
No human tunes weights. No hardcoded fallback chains. Pure math.

**Demo moment**: Engine coolant rupture. `eng.thrust` error rate spikes from 12% to 42%.
Thompson Sampling detects it. Hard constraint (`max_error_rate=20%`) kicks in. Agents
stop using the broken tool — automatically. Zero human intervention.

### Pillar 2 — Event-Driven Agents

**What**: Agents are not polling, not waiting for instructions, not running on cron.
NATS events trigger specific agents for specific jobs. When something happens, the right
agent fires — instantly, surgically.

**Why it matters**: This is how **production** systems work. Not "give the LLM a loop and
let it figure it out." Each event type has a registered handler. Each handler knows exactly
what tools it needs, what context to load, and what to do. It's reactive, not speculative.

**Demo moment**: Sensor detects anomaly → `diagnostic-agent` fires automatically.
Cascade failure detected → `repair-agent` activates. Hull breach confirmed →
`structural-agent` deploys. Nobody orchestrates this. The events drive it.

**Architecture**:
```
NATS Event                    Agent Fleet
─────────────                 ───────────
sensor.anomaly.detected   →   diagnostic-agent    (scan, assess)
engine.failure.critical   →   repair-agent         (eng.thrust, power.reroute)
hull.integrity.warning    →   structural-agent     (hull.seal, shield.mod)
policy.updated            →   ranking-agent        (Thompson Sampling refresh)
context.rehydrated        →   recovery-agent       (resume from checkpoint)
```

### Pillar 3 — Surgical Context Reconstruction (Kernel)

**What**: Instead of dumping 128K tokens of "everything" into the prompt, the kernel
reconstructs **exactly** the context each agent needs — from a Neo4j graph, role-scoped,
token-budgeted. An 8B-parameter model with 3K precise tokens outperforms a 70B model
drowning in 100K tokens of noise.

**Why it matters**: The industry is racing toward bigger context windows and bigger models.
We go the opposite direction. The rehydration kernel queries the task graph, loads the
root node + neighbors + relationships + expanded detail, and renders a **bounded context
bundle** — typically 2,000-4,000 tokens. That's it. No padding. No "just in case." Every
token earns its place.

**The math** (benchmark-measured, real content, `go test -v ./internal/benchmark/`):
```
Scenario                    Traditional   Surgical    Ratio
────────────────────────    ───────────   ────────    ─────
Simple diagnostic (1 turn)     3,120         96      32.5×
Multi-turn repair (15 turns)   6,190        394      15.7×
Complex rehydration            7,103        504      14.1×
```

These are measured from real scenario content with the same token estimator (len/4).
The traditional content includes: system prompt, tool descriptions with JSON schemas,
conversation history with tool call/result pairs, RAG chunks, workspace file dumps, and
previous attempt logs. The surgical content is what the kernel actually produces.

**Apples-to-apples cost (same model, fewer tokens):**
```
                        Claude Sonnet, 1,000 calls/day, Scenario 2
                        ───────────────────────────────────────────
Traditional:            $18.57/day
Surgical:                $1.18/day
Savings:                15.7×
```

**Blended model routing (95% local + 5% Opus):**
```
Traditional (all Sonnet):       $18.57/day
Surgical (95% local, 5% Opus):   $0.33/day
Combined savings:                56×
```

The **14-32x token reduction** is the defensible, measured number. Not 2000x. Not 42x.
14-32x depending on scenario complexity, verified by running `make costbench`.

**Demo moment**: The rehydration bundle for the engine repair mission:
```
Root:       node:mission:engine-core-failure
Role:       implementer
Nodes:      7     Relationships: 6
Details:    3 node details loaded
Tokens:     2,847 / 4,000 budget     ← not 128,000. Two thousand eight hundred.
Snapshot:   snap_uss_20260312T154230Z
```

This is why a Qwen3-8B running on a single GPU can do what others need GPT-4 for.
**The context is the product, not the model.**

### Pillar 4 — Context Rehydration & Branching

**What**: When an agent realizes it's on the wrong path, it rolls back to a previous
checkpoint in the task graph and creates a new solution branch. Full audit trail. Both
the failed path and the new path are preserved in the graph.

**Why it matters**: Every agent framework runs forward. When the approach fails, they
retry the same thing or give up. Our agents can **go back in time** — load the context
from an earlier decision point, understand what was tried and failed, and branch a
completely new strategy. Like `git checkout -b` for agent reasoning.

**Demo moment**: Agent-7 tried to repair the engine directly. But the hull was compromised.
Each attempt made things worse. Agent recognizes the pattern:

```
AGENT-7 > Engine repair attempt #3 failed. Hull stress at 76%.
AGENT-7 > Current strategy is counterproductive. Initiating rehydration.
AGENT-7 > Checkpoint ALPHA-3 loaded. New branch: hull-first protocol.
```

The task graph shows both paths — the abandoned one and the new one:
```
● [1] Diagnose anomaly ................................. DONE
 │
● [2] Assess cascade damage ............................ DONE
 │
 ├── ✗ Path A: Direct engine repair .................... ABANDONED
 │    Hull stress +12%. Strategy counterproductive.
 │
 └── ◉ Path B: Hull-first protocol ..................... NEW BRANCH
      ├─ [3] Seal hull breaches ........................ IN PROGRESS
      ├─ [4] Stabilize power grid ...................... PENDING
      └─ [5] Repair engine (safe conditions) ........... PENDING
```

---

## The Closed Loop

This is what makes the demo **WOW** — the four pillars form a closed loop:

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│   Event Happens ──→ Agent Fires ──→ Kernel Builds ──→ Uses Best     │
│        ↑              (P2)          Surgical Context    Tools (P1)   │
│        │                             3K tokens (P3)       │          │
│        │                                                  │          │
│        │                                            Tool Fails?      │
│        │                                            ┌── Yes ──┐      │
│        │                                            │          │      │
│        │                                       Constraint   Wrong    │
│        │                                       Filters It   Path?    │
│        │                                            │     ┌─ Yes ─┐  │
│        │                                            ↓     ↓       │  │
│        │                                        Use Alt. Rehydrate│  │
│        │                                         Tool    + Branch │  │
│        │                                            │     (P4)    │  │
│        │                                            ↓     ↓       │  │
│        └──────────── New Events ←──────────────  Continue  New    │  │
│                                                         Branch   │  │
│                                                           └──────┘  │
│                                                                      │
│   NO HUMAN IN THE LOOP                                               │
│   P1=Thompson Sampling  P2=Event-Driven  P3=Surgical Context  P4=Rehydration  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

**Key message**: The system is autonomous. Events trigger the right agent. The kernel
reconstructs precisely the context needed (not 128K — just 3K). Thompson Sampling picks
the best tools. When the approach fails, the agent rehydrates context and branches a new
strategy. All without a human pressing buttons.

---

## Intelligent Model Routing — The Fifth Dimension

The system is not anti-big-models. It's **anti-waste**. Thompson Sampling doesn't just
rank tools — it ranks **model + tool combinations**. The system learns when a Qwen3-8B
is enough and when to escalate to Claude or GPT-4.

### How It Works

```
Task Complexity        Model Selected           Cost
────────────────       ──────────────           ──────
Routine scan           Qwen3-8B (local GPU)     $0.00
Diagnostics            Qwen3-8B (local GPU)     $0.00
Repair execution       Qwen3-8B (local GPU)     $0.00
Strategy assessment    Claude Sonnet (API)      $0.003
Rehydration decision   Claude Opus (API)        $0.015
Novel failure mode     GPT-4o (API)             $0.010
```

Thompson Sampling tracks success/failure per **(model, tool, context)** triple.
If Qwen3-8B starts failing on complex diagnostic tasks, the system automatically
routes those tasks to a more capable model. If the big model is overkill for routine
scans, it stops using it there. **No human configures routing rules.** The math does it.

### Why This Matters

1. **Cost optimization**: 95% of tasks run on local GPU ($0). Only complex decisions
   hit the API. Total cost drops 50-100x vs "GPT-4 for everything."
2. **Latency optimization**: Local model responds in 200ms. API call takes 2-3s.
   Routine tasks stay fast.
3. **Resilience**: If the API is down, local model handles everything. If the local
   model degrades, API picks up the slack. Thompson Sampling adapts automatically.
4. **Best-of-breed**: Use Claude for reasoning, GPT-4 for code generation, local
   model for execution — whatever the data shows works best.

### Demo Moment

In the spaceship scenario:
- **Phases 0-4**: Qwen3-8B handles everything (routine ops, diagnosis, adaptation)
- **Phase 5**: Small model's repair strategy is failing. Confidence dropping.
- **Phase 6**: System escalates to Claude Opus for the rehydration decision.
  The big model analyzes the task graph, identifies the failed path, and recommends
  hull-first protocol. Cost: $0.015 for the one call that matters.
- **Phase 7-8**: Qwen3-8B executes the new strategy. Claude validates the final state.

```
AGENT-7 [qwen3-8b]  > Repair attempt #3 failed. Escalating decision.
AGENT-7 [claude-opus] > Task graph analysis: Path A counterproductive.
                        Hull integrity must precede engine repair.
                        Recommending checkpoint rollback + hull-first branch.
AGENT-7 [qwen3-8b]  > Acknowledged. Executing hull-first protocol.
```

**Message**: "We don't pick one model. We let the math pick the right model for each job."

---

## The Spaceship Narrative

### Why a spaceship?

1. **Memorable**: People forget "tool ranking demo." People remember "the starship demo."
2. **Intuitive**: Everyone understands "engine broke, ship needs repair." No domain expertise needed.
3. **Dramatic**: Cascading failures, wrong decisions, heroic recovery — it's a *story*.
4. **Maps perfectly**: Ship subsystems = agent tools. Failures = tool degradation. Repair strategy = task graph.

### Mission: Engine Core Failure

| Phase | Status | What Happens | Pillar | Model |
|-------|--------|--------------|--------|-------|
| 0 | NOMINAL | All systems green. AI fleet on routine patrol. | Baseline | Qwen3-8B |
| 1 | WARNING | Sensor anomaly in engine. `diagnostic-agent` fires. | Event-Driven | Qwen3-8B |
| 2 | CRITICAL | Coolant rupture! Engine offline. Error rate 42%. | Tool Degradation | Qwen3-8B |
| 3 | CASCADE | Power grid and shields degrading. 3 systems down. | Cascade Realism | Qwen3-8B |
| 4 | ADAPTING | Thompson Sampling filters broken tools. Agents switch. | Thompson Sampling | Qwen3-8B |
| 5 | FAILING | Wrong repair strategy. Making things worse. | Problem Setup | Qwen3-8B |
| 6 | ESCALATION | System escalates to Claude Opus for strategic decision. | Model Routing | Claude Opus |
| 7 | REHYDRATING | Agent rolls back. New branch in task graph. | Rehydration | Claude Opus |
| 8 | RECOVERING | Hull-first protocol executing. Systems improving. | New Branch | Qwen3-8B |
| 9 | NOMINAL | Ship repaired. Fleet adapted. No human needed. | Full Loop | Qwen3-8B |

### Tool → Ship System Mapping

| Tool ID | Ship System | Normal Behavior | Degraded Behavior |
|---------|-------------|-----------------|-------------------|
| `nav.plot` | Navigation computer | Route calculation, 120ms | Drift correction needed |
| `scan.deep` | Sensor array | Deep space scanning | Intermittent readings |
| `eng.thrust` | Engine core | Propulsion control | **Coolant rupture** |
| `hull.seal` | Hull integrity | Structural repair | Micro-fractures from vibration |
| `comm.burst` | Communications | Subspace transmission | Signal degradation |
| `life.recycle` | Life support | Atmosphere recycling | Backup mode |
| `power.reroute` | Power grid | Load balancing | **Cascade overload** |
| `shield.mod` | Shield generator | Frequency modulation | **Intermittent drops** |

---

## Is It Enough for WOW?

### What we HAVE (already built)

- Tool-learning pipeline: seed lake → DuckDB → Thompson Sampling → Valkey → NATS (production-grade)
- Rehydration kernel: Neo4j graph → context bundles → gRPC API (production-grade)
- Starship demo runner: 2-phase mission with real LLM (validated with Qwen3-8B)
- TUI client: Bubble Tea with 5 views, mTLS security, hexagonal architecture
- 14 integration tests, E2E pipeline test, policy correctness validation
- Full Helm charts, Docker Compose, CI/CD

### What would push it to SUPER WOW

| Enhancement | Impact | Effort |
|-------------|--------|--------|
| Event-driven agent dispatch (NATS → agent activation) | HIGH | Medium |
| Live terminal recording (VHS/asciinema) with narration | HIGH | Low |
| Web dashboard companion (real-time Grafana-style) | MEDIUM | High |
| Multi-agent coordination (agents communicating) | HIGH | High |
| Side-by-side: "without Underpass" vs "with Underpass" | HIGH | Low |
| Real model running live (not pre-recorded) | HIGH | Medium |

### Verdict

**YES, it's WOW.** The four pillars together tell a story nobody else is telling:

| Them | Us |
|------|-----|
| LangChain/CrewAI: "Chain tools together" | Agents **learn** which tools to trust (Bayesian math) |
| AutoGPT/OpenDevin: "Loop until done" | Agents **react** to events, **rollback** when wrong |
| Everyone: "Stuff 128K tokens in context" | **3K surgical tokens** from a graph. Cheaper. Faster. More accurate. |
| Everyone: "Use GPT-4 for everything" | **Right model for each job** — 8B local for routine, Opus for strategy |
| Everyone: "Retry on failure" | **Branch a new strategy** from a checkpoint in the task graph |
| Everyone: "Pick one model provider" | **Best-of-breed routing** — Claude, GPT-4, Qwen, whatever the data shows |

**The precision thesis**: The industry races toward bigger models and bigger context windows.
We prove you need neither for 95% of tasks. Small models + surgical context + learned tool
selection + intelligent model routing = a system that's cheaper, faster, and more resilient.
And when you DO need a big model, the system knows exactly when to escalate.

The spaceship narrative makes it accessible. The math makes it credible. The architecture
makes it production-real. The cost comparison makes CTOs pay attention.

---

## Content Strategy

### Target Audiences

1. **AI Engineers / ML Engineers**: Care about Thompson Sampling, Bayesian tool selection, production patterns
2. **Platform Engineers / SREs**: Care about event-driven architecture, observability, self-healing
3. **Engineering Leaders / CTOs**: Care about autonomous agents that don't need human babysitting
4. **AI Community (general)**: Care about the narrative, the innovation, the demo

### Content Pieces

#### 1. Short Video (2-3 min) — LinkedIn, X/Twitter

**Format**: Terminal recording with voiceover narration.
**Tool**: [VHS](https://github.com/charmbracelet/vhs) (from Charmbracelet, same team as Bubble Tea).
Produces clean GIF/MP4 from a tape file.

**Script outline**:
1. (0:00-0:15) Hook: "What happens when your AI agent's tools start failing?"
2. (0:15-0:45) Show TUI — spaceship cruising, all green
3. (0:45-1:15) Engine failure, cascade, Thompson Sampling adapts
4. (1:15-1:45) Wrong strategy, agent realizes, REHYDRATION moment
5. (1:45-2:15) New branch, recovery, "no human needed"
6. (2:15-2:30) Call to action: "This is Underpass. Link in comments."

**Key**: Fast cuts. Bold terminal colors. Dramatic music optional.
The VHS recording should show the TUI at 1.5x speed during "boring" parts
and real-time during the dramatic moments.

#### 2. Technical Deep Dive (15-20 min) — YouTube

**Format**: Screen recording + architecture diagrams + live terminal.

**Outline**:
1. The problem: Why current agent frameworks break in production
2. Pillar 1: Thompson Sampling — the math, the pipeline, live demo
3. Pillar 2: Event-driven agents — NATS architecture, agent dispatch
4. Pillar 3: Context rehydration — the kernel, task graph, rollback
5. The spaceship mission — full walkthrough
6. Architecture diagram — how it all connects
7. What's next / call to participate

#### 3. Blog Post — Medium / dev.to / Hashnode

**Title options**:
- "AI Agents That Learn, React, and Recover: Building a Self-Healing Agent Fleet"
- "Beyond Tool Calling: How Bayesian Sampling Teaches Agents Which Tools to Trust"
- "Event-Driven AI Agents: Why Your Agent Framework Needs a Message Bus"

**Structure**:
1. The premise (spaceship hook)
2. The three pillars (with code snippets and diagrams)
3. The closed loop (architecture diagram)
4. Results / what we learned
5. Try it yourself (GitHub link)

**Key**: Include the Thompson Sampling formula, the task graph diagram,
and the event-driven architecture. Technical credibility matters.

#### 4. GitHub README — Star-worthy

The underpass-demo repo README should be a mini-landing page:
- GIF/screenshot at the top (terminal recording)
- One-line description
- "Quick Start" with `make demo` (embedded mode, zero infra)
- Architecture diagram
- Link to blog post and video

#### 5. LinkedIn Post — Personal Brand

**Format**: Text post with video/GIF attachment.

**Structure**:
- Hook line (question or bold statement)
- 3 bullet points (one per pillar)
- "We built this and open-sourced it"
- Video/GIF attachment
- Relevant hashtags

**Example**:
> Your AI agent uses 128,000 tokens of context per call.
> Ours uses 3,000. And it's more accurate.
>
> We built an agent fleet that:
> - **Learns** which tools to trust (Bayesian Thompson Sampling, not vibes)
> - **Reacts** to problems instantly (event-driven, not polling)
> - **Focuses** on exactly the right context (graph-based, 3K tokens, not 128K)
> - **Routes** to the right model (Qwen3-8B for routine, Claude Opus for strategy)
> - **Recovers** from mistakes (rolls back to a checkpoint, branches a new strategy)
>
> 95% of tasks run on a local 8B model. When the problem is truly hard,
> the system escalates to Claude or GPT-4 — automatically.
> No human in the loop. The math decides everything.
>
> Here's a 2-minute demo with a spaceship that breaks down. [video]

---

## Production Order

### Phase 1 — Demo Complete (current sprint)

1. Finish TUI spaceship retheme (mission view, bridge, systems, sampling, log)
2. Add embedded mode (zero-infra, `--embedded` flag)
3. Add event-driven agent dispatch view (new TUI tab showing NATS → agent activation)
4. Record VHS tape for short video

### Phase 2 — Content Production (next sprint)

5. Write blog post (Medium)
6. Record YouTube deep dive
7. Create LinkedIn post with video
8. Polish GitHub README with GIF

### Phase 3 — Community Engagement

9. Post to Hacker News / Reddit r/MachineLearning
10. Share in AI Discord communities
11. Submit to AI/DevOps conferences (KubeCon, AI Engineer Summit)
12. Engage with comments and feedback

---

## Recording Setup

### VHS (Charmbracelet)

VHS produces reproducible terminal recordings from a `.tape` file:

```
# demo.tape
Output demo.mp4
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

Type "make demo"
Enter
Sleep 2s
# ... keypress sequence for the mission phases
```

**Advantages**: Reproducible, pixel-perfect, no screen recording artifacts.
Same ecosystem as Bubble Tea (Charmbracelet). Professional look.

### Narration

Record voiceover separately, sync in post-production.
Keep it conversational, not corporate. "Watch what happens when the engine fails..."

---

## Key Messages (for all content)

1. **"AI agents that learn, react, focus, and recover"** — the tagline
2. **"3K tokens, not 128K"** — the precision headline (gets attention)
3. **"No human in the loop"** — the promise
4. **"Thompson Sampling, not vibes"** — math > heuristics
5. **"The context is the product, not the model"** — the paradigm shift
6. **"Right model, right task, right time"** — intelligent model routing
7. **"Event-driven, not loop-driven"** — production architecture
8. **"Git for agent decisions"** — context rehydration explained simply
9. **"The spaceship that repairs itself"** — the narrative hook
10. **"Qwen3-8B for routine. Claude Opus when it matters."** — cost-efficiency story
