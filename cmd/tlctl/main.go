// tlctl is the USS Underpass Demo TUI — a Bubble Tea client that visualizes
// the tool-learning Thompson Sampling pipeline through a spaceship narrative.
//
// Usage:
//
//	tlctl --embedded                          # zero-infra demo mode
//	tlctl --valkey-addr=localhost:6379        # live Valkey data
//	tlctl --valkey-addr=localhost:6379 --nats-url=nats://localhost:4222
//	tlctl --embedded --kernel-addr=localhost:50054  # real kernel, rest simulated
//	tlctl --embedded --record-session         # record session to NDJSON
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/underpass-ai/underpass-demo/internal/adapters/embedded"
	kerneladapter "github.com/underpass-ai/underpass-demo/internal/adapters/kernel"
	natsadapter "github.com/underpass-ai/underpass-demo/internal/adapters/nats"
	"github.com/underpass-ai/underpass-demo/internal/adapters/session"
	valkeyadapter "github.com/underpass-ai/underpass-demo/internal/adapters/valkey"
	"github.com/underpass-ai/underpass-demo/internal/tui"
)

func main() {
	embeddedMode := flag.Bool("embedded", false, "Run with in-memory ship data (zero infrastructure)")
	idle := flag.Bool("idle", false, "Keep process alive without launching TUI (for K8s pods)")
	valkeyAddr := flag.String("valkey-addr", "localhost:6379", "Valkey address")
	valkeyPass := flag.String("valkey-pass", "", "Valkey password")
	valkeyDB := flag.Int("valkey-db", 0, "Valkey database")
	valkeyPrefix := flag.String("valkey-prefix", "tool_policy", "Valkey key prefix")
	natsURL := flag.String("nats-url", "", "NATS URL (optional, for event stream)")
	kernelAddr := flag.String("kernel-addr", "", "Kernel gRPC address (host:port). Empty = embedded simulator.")
	recordSession := flag.Bool("record-session", false, "Record session to NDJSON file (~/.config/tlctl/sessions/)")
	flag.Parse()

	if *idle {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		<-sig
		return
	}

	if err := run(*embeddedMode, *valkeyAddr, *valkeyPass, *valkeyDB, *valkeyPrefix, *natsURL, *kernelAddr, *recordSession); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(embeddedMode bool, valkeyAddr, valkeyPass string, valkeyDB int, valkeyPrefix, natsURL, kernelAddr string, recordSession bool) error {
	deps := tui.Deps{}

	if embeddedMode {
		deps.PolicyReader = embedded.NewPolicyStore()
		deps.EventSub = embedded.NewEventSimulator()
	} else {
		// Connect to Valkey.
		reader, err := valkeyadapter.NewPolicyReader(valkeyAddr, valkeyPass, valkeyDB, valkeyPrefix)
		if err != nil {
			return fmt.Errorf("valkey: %w", err)
		}
		defer func() { _ = reader.Close() }()
		deps.PolicyReader = reader

		// Optionally connect to NATS.
		if natsURL != "" {
			sub, err := natsadapter.NewSubscriber(natsURL)
			if err != nil {
				return fmt.Errorf("nats: %w", err)
			}
			defer sub.Close()
			deps.EventSub = sub
		}
	}

	// Kernel: real gRPC or embedded simulator.
	if kernelAddr != "" {
		kc, err := kerneladapter.NewGRPCClient(kernelAddr)
		if err != nil {
			return fmt.Errorf("kernel: %w", err)
		}
		defer func() { _ = kc.Close() }()
		deps.ContextProvider = kc
	} else {
		deps.ContextProvider = embedded.NewContextSimulator()
	}

	// Session recording: NDJSON or no-op.
	if recordSession {
		rec, err := session.NewNDJSONRecorder()
		if err != nil {
			return fmt.Errorf("session recorder: %w", err)
		}
		defer func() { _ = rec.Close() }()
		deps.SessionRecorder = rec
		fmt.Fprintf(os.Stderr, "Recording session to: %s\n", rec.Path())
	} else {
		deps.SessionRecorder = embedded.NewNoopRecorder()
	}

	model := tui.NewModel(deps)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
