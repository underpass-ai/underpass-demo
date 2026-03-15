// Package session provides session recording adapters.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// NDJSONRecorder writes session entries as newline-delimited JSON to
// ~/.config/tlctl/sessions/session_<timestamp>.ndjson.
type NDJSONRecorder struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
}

// NewNDJSONRecorder creates the sessions directory and opens a new NDJSON file.
func NewNDJSONRecorder() (*NDJSONRecorder, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("session recorder: %w", err)
	}
	dir := filepath.Join(home, ".config", "tlctl", "sessions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("session dir: %w", err)
	}
	name := fmt.Sprintf("session_%s.ndjson", time.Now().Format("20060102T150405Z"))
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("session file: %w", err)
	}
	return &NDJSONRecorder{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

func (r *NDJSONRecorder) Record(entry domain.SessionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enc.Encode(entry)
}

func (r *NDJSONRecorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}

// Path returns the file path for the current session log.
func (r *NDJSONRecorder) Path() string {
	return r.file.Name()
}
