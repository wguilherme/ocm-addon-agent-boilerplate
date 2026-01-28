package contracts

import "context"

// SyncService defines the interface for agent synchronization services.
// Each service implements a specific synchronization operation (inventory, metrics, events).
type SyncService interface {
	// Name returns the service name for logging.
	Name() string

	// Sync executes the synchronization operation.
	Sync(ctx context.Context, config *SyncConfig) error
}
