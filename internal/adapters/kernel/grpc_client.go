// Package kernel provides a gRPC adapter for the rehydration kernel service.
package kernel

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/underpass-ai/underpass-demo/internal/gen/underpass/rehydration/kernel/v1alpha1"

	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// GRPCClient implements ports.ContextProvider by calling the real kernel.
type GRPCClient struct {
	conn  *grpc.ClientConn
	query pb.QueryServiceClient
}

// NewGRPCClient connects to the kernel at addr (host:port).
// Uses insecure credentials for demo; mTLS will be added later.
func NewGRPCClient(addr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("kernel dial: %w", err)
	}
	return &GRPCClient{
		conn:  conn,
		query: pb.NewQueryServiceClient(conn),
	}, nil
}

// GetContext calls the kernel's GetContext RPC and maps the response to domain types.
func (c *GRPCClient) GetContext(ctx context.Context, req domain.ContextRequest) (*domain.ContextResult, error) {
	resp, err := c.query.GetContext(ctx, &pb.GetContextRequest{
		RootNodeId:  req.RootNodeID,
		Role:        req.Role,
		Phase:       req.Phase,
		TokenBudget: req.TokenBudget,
		Scopes:      req.Scopes,
	})
	if err != nil {
		return nil, fmt.Errorf("kernel GetContext: %w", err)
	}
	return mapContextResponse(resp), nil
}

// GetGraphRelationships calls the kernel's GetGraphRelationships RPC.
func (c *GRPCClient) GetGraphRelationships(ctx context.Context, nodeID, nodeKind string, depth uint32) (*domain.GraphResult, error) {
	resp, err := c.query.GetGraphRelationships(ctx, &pb.GetGraphRelationshipsRequest{
		NodeId:   nodeID,
		NodeKind: nodeKind,
		Depth:    depth,
	})
	if err != nil {
		return nil, fmt.Errorf("kernel GetGraphRelationships: %w", err)
	}
	return mapGraphResponse(resp), nil
}

// Close tears down the gRPC connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// ── Anti-corruption layer: proto → domain ───────────────────────────────────

func mapContextResponse(r *pb.GetContextResponse) *domain.ContextResult {
	res := &domain.ContextResult{
		RootNodeID:  r.GetRootNodeId(),
		TokenCount:  r.GetTokenCount(),
		ContentHash: r.GetContentHash(),
		SnapshotID:  r.GetSnapshotId(),
	}
	for _, n := range r.GetNodes() {
		res.Nodes = append(res.Nodes, mapNode(n))
	}
	for _, rel := range r.GetRelationships() {
		res.Relationships = append(res.Relationships, mapRel(rel))
	}
	for _, d := range r.GetDetails() {
		res.Details = append(res.Details, mapDetail(d))
	}
	for _, s := range r.GetSections() {
		res.Sections = append(res.Sections, domain.BundleSection{
			Heading:    s.GetHeading(),
			Content:    s.GetContent(),
			TokenCount: s.GetTokenCount(),
		})
	}
	return res
}

func mapGraphResponse(r *pb.GetGraphRelationshipsResponse) *domain.GraphResult {
	res := &domain.GraphResult{
		Root: mapNode(r.GetRoot()),
	}
	for _, n := range r.GetNeighbors() {
		res.Neighbors = append(res.Neighbors, mapNode(n))
	}
	for _, rel := range r.GetRelationships() {
		res.Relationships = append(res.Relationships, mapRel(rel))
	}
	return res
}

func mapNode(n *pb.GraphNode) domain.GraphNode {
	if n == nil {
		return domain.GraphNode{}
	}
	return domain.GraphNode{
		ID:       n.GetId(),
		Kind:     n.GetKind(),
		Label:    n.GetLabel(),
		Status:   n.GetStatus(),
		Metadata: n.GetMetadata(),
	}
}

func mapRel(r *pb.GraphRelationship) domain.GraphRelationship {
	if r == nil {
		return domain.GraphRelationship{}
	}
	return domain.GraphRelationship{
		SourceID: r.GetSourceId(),
		TargetID: r.GetTargetId(),
		Kind:     r.GetKind(),
		Label:    r.GetLabel(),
	}
}

func mapDetail(d *pb.NodeDetail) domain.NodeDetail {
	if d == nil {
		return domain.NodeDetail{}
	}
	return domain.NodeDetail{
		NodeID:      d.GetNodeId(),
		Title:       d.GetTitle(),
		Description: d.GetDescription(),
		Properties:  d.GetProperties(),
	}
}
