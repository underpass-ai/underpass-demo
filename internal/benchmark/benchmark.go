// Package benchmark measures context precision savings between traditional
// agent frameworks (stuff everything in context) and Underpass's surgical
// approach (graph-based, token-budgeted bundles from the rehydration kernel).
package benchmark

import "fmt"

// EstimateTokens approximates token count for a text string.
// Standard approximation: ~4 characters per token for English/code mix.
// Gives ±10% vs tiktoken — acceptable for comparative benchmarking.
// The comparison uses the SAME estimator for both approaches, so the
// ratio is accurate even if absolute counts have small error.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}

// ContextSection is a named segment of an agent's context window.
type ContextSection struct {
	Name    string
	Content string
}

// Scenario defines a task with both traditional and surgical context approaches.
type Scenario struct {
	Name        string
	Description string
	Traditional []ContextSection
	Surgical    []ContextSection
}

// TotalTokens returns the sum of estimated tokens across all sections.
func TotalTokens(sections []ContextSection) int {
	total := 0
	for _, s := range sections {
		total += EstimateTokens(s.Content)
	}
	return total
}

// SectionBreakdown returns a formatted breakdown of token counts per section.
func SectionBreakdown(sections []ContextSection) string {
	result := ""
	for i, s := range sections {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%s:%d", s.Name, EstimateTokens(s.Content))
	}
	return result
}

// Provider represents an LLM provider's pricing.
type Provider struct {
	Name      string
	InputPerM float64 // USD per million input tokens
}

// Providers returns current public pricing as of 2026-03.
func Providers() []Provider {
	return []Provider{
		{Name: "Claude Sonnet 4", InputPerM: 3.00},
		{Name: "Claude Opus 4", InputPerM: 15.00},
		{Name: "GPT-4o", InputPerM: 2.50},
		{Name: "Claude Haiku 4.5", InputPerM: 0.80},
		{Name: "Gemini 2.5 Pro", InputPerM: 1.25},
		{Name: "Local 8B (amortized)", InputPerM: 0.10},
	}
}

// CostPerCall calculates input cost for a given token count and provider.
func CostPerCall(tokens int, pricePerM float64) float64 {
	return float64(tokens) * pricePerM / 1_000_000
}
