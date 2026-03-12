// Package domain holds the core value objects for the demo client.
package domain

import "time"

// ToolPolicy is a computed policy for a (context, tool) pair,
// mirroring the tool-learning service's domain model.
type ToolPolicy struct {
	ContextSignature string    `json:"context_signature"`
	ToolID           string    `json:"tool_id"`
	Alpha            float64   `json:"alpha"`
	Beta             float64   `json:"beta"`
	P95LatencyMs     int64     `json:"p95_latency_ms"`
	P95Cost          float64   `json:"p95_cost"`
	ErrorRate        float64   `json:"error_rate"`
	NSamples         int64     `json:"n_samples"`
	FreshnessTs      time.Time `json:"freshness_ts"`
	Confidence       float64   `json:"confidence"`
}

// PolicyUpdateEvent is the NATS event published by tool-learning.
type PolicyUpdateEvent struct {
	Event            string `json:"event"`
	Ts               string `json:"ts"`
	Schedule         string `json:"schedule"`
	PoliciesWritten  int    `json:"policies_written"`
	PoliciesFiltered int    `json:"policies_filtered"`
}
