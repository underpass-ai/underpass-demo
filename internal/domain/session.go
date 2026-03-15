package domain

import "time"

// SessionRecord is a single entry in the session recording log.
type SessionRecord struct {
	Kind  string    `json:"kind"` // phase, event, dispatch, bundle, kernel
	Ts    time.Time `json:"ts"`
	Phase int       `json:"phase"`
	Data  any       `json:"data"`
}
