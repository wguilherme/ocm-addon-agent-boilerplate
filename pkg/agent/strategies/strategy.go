// Package strategies implements the Strategy Pattern for agent sync operations.
// Each strategy represents a different way to communicate data from spoke to hub.
package strategies

import (
	"context"

	"github.com/totvs/addon-framework-basic/pkg/agent/contracts"
)

// SyncStrategy defines the interface for all sync strategies.
// Each strategy implements a different method to communicate spoke data to the hub.
type SyncStrategy interface {
	// Name returns the strategy name for logging purposes.
	Name() string

	// Sync executes the synchronization operation.
	// It receives both spoke and hub clients to perform the necessary operations.
	Sync(ctx context.Context, config *contracts.SyncConfig) error
}

// SyncConfig is an alias to contracts.SyncConfig for backward compatibility.
type SyncConfig = contracts.SyncConfig
