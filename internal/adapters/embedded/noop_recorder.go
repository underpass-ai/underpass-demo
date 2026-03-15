package embedded

import "github.com/underpass-ai/underpass-demo/internal/domain"

// NoopRecorder implements ports.SessionRecorder by discarding all entries.
// Used when --record-session is not passed.
type NoopRecorder struct{}

func NewNoopRecorder() *NoopRecorder { return &NoopRecorder{} }

func (r *NoopRecorder) Record(_ domain.SessionRecord) error { return nil }
func (r *NoopRecorder) Close() error                        { return nil }
