package domain

// ContextRequest describes what rehydration context to build.
type ContextRequest struct {
	RootNodeID  string
	Role        string
	Phase       string
	TokenBudget uint32
	Scopes      []string
}

// ContextResult is the kernel's response — a scoped rehydration bundle.
type ContextResult struct {
	RootNodeID    string
	Nodes         []GraphNode
	Relationships []GraphRelationship
	Details       []NodeDetail
	Sections      []BundleSection
	TokenCount    uint32
	ContentHash   string
	SnapshotID    string
}

// GraphNode represents a node in the task/knowledge graph.
type GraphNode struct {
	ID       string
	Kind     string // e.g. "mission", "task", "decision"
	Label    string
	Status   string // e.g. "done", "active", "pending", "abandoned"
	Metadata map[string]string
}

// GraphRelationship represents an edge between two graph nodes.
type GraphRelationship struct {
	SourceID string
	TargetID string
	Kind     string // e.g. "has_subtask", "branched_from", "depends_on"
	Label    string
}

// NodeDetail provides extended information for a graph node.
type NodeDetail struct {
	NodeID      string
	Title       string
	Description string
	Properties  map[string]string
}

// BundleSection is a rendered section of a rehydration bundle.
type BundleSection struct {
	Heading    string
	Content    string
	TokenCount uint32
}

// GraphResult is the response from a graph traversal query.
type GraphResult struct {
	Root          GraphNode
	Neighbors     []GraphNode
	Relationships []GraphRelationship
}
