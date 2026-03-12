// Package valkey implements the PolicyReader port using Redis-compatible commands.
package valkey

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/underpass-ai/underpass-demo/internal/domain"
)

// PolicyReader reads tool policies from a Valkey (Redis-compatible) store.
type PolicyReader struct {
	client    *redis.Client
	keyPrefix string
}

// NewPolicyReader creates a reader that scans keys with the given prefix.
func NewPolicyReader(addr, password string, db int, keyPrefix string) (*PolicyReader, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("valkey ping: %w", err)
	}
	return &PolicyReader{client: client, keyPrefix: keyPrefix}, nil
}

// ReadAll scans all keys matching the prefix and returns their policies.
func (r *PolicyReader) ReadAll(ctx context.Context) ([]domain.ToolPolicy, error) {
	return r.scanKeys(ctx, r.keyPrefix+":*")
}

// ReadByContext returns policies for a specific context signature.
func (r *PolicyReader) ReadByContext(ctx context.Context, contextSig string) ([]domain.ToolPolicy, error) {
	return r.scanKeys(ctx, r.keyPrefix+":"+contextSig+":*")
}

func (r *PolicyReader) scanKeys(ctx context.Context, pattern string) ([]domain.ToolPolicy, error) {
	var policies []domain.ToolPolicy
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		data, err := r.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var p domain.ToolPolicy
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		policies = append(policies, p)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return policies, nil
}

// Close releases the Redis connection.
func (r *PolicyReader) Close() error {
	return r.client.Close()
}
