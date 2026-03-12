// tlctl is the USS Underpass Demo TUI — a Bubble Tea client that visualizes
// the tool-learning Thompson Sampling pipeline through a spaceship narrative.
//
// Usage:
//
//	tlctl --embedded                          # zero-infra demo mode
//	tlctl --valkey-addr=localhost:6379        # live Valkey data
//	tlctl --valkey-addr=localhost:6379 --nats-url=nats://localhost:4222
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/underpass-ai/underpass-demo/internal/adapters/embedded"
	natsadapter "github.com/underpass-ai/underpass-demo/internal/adapters/nats"
	valkeyadapter "github.com/underpass-ai/underpass-demo/internal/adapters/valkey"
	"github.com/underpass-ai/underpass-demo/internal/tui"
)

func main() {
	embeddedMode := flag.Bool("embedded", false, "Run with in-memory ship data (zero infrastructure)")
	valkeyAddr := flag.String("valkey-addr", "localhost:6379", "Valkey address")
	valkeyPass := flag.String("valkey-pass", "", "Valkey password")
	valkeyDB := flag.Int("valkey-db", 0, "Valkey database")
	valkeyPrefix := flag.String("valkey-prefix", "tool_policy", "Valkey key prefix")
	natsURL := flag.String("nats-url", "", "NATS URL (optional, for event stream)")
	flag.Parse()

	if err := run(*embeddedMode, *valkeyAddr, *valkeyPass, *valkeyDB, *valkeyPrefix, *natsURL); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(embeddedMode bool, valkeyAddr, valkeyPass string, valkeyDB int, valkeyPrefix, natsURL string) error {
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

	model := tui.NewModel(deps)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
