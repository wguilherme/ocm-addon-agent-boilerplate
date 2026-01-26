package strategies

import (
	"context"

	"k8s.io/klog/v2"
)

// HelloStrategy implements a simple strategy that just logs hello.
// This is a minimal example to demonstrate the Strategy Pattern.
type HelloStrategy struct{}

// NewHelloStrategy creates a new Hello strategy.
func NewHelloStrategy() *HelloStrategy {
	return &HelloStrategy{}
}

// Name returns the strategy name.
func (s *HelloStrategy) Name() string {
	return "Hello"
}

// Sync logs a hello message.
func (s *HelloStrategy) Sync(ctx context.Context, config *SyncConfig) error {
	klog.Infof("Hello from cluster %s!", config.SpokeClusterName)
	return nil
}
