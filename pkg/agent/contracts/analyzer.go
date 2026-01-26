package contracts

import (
	"context"
)

// Collector coleta recursos Kubernetes do tipo T do spoke cluster.
type Collector[T any] interface {
	Collect(ctx context.Context, config *SyncConfig) ([]T, error)
	Name() string
}

// Processor transforma dados T em relatório R (agregações, métricas, filtros).
type Processor[T any, R any] interface {
	Process(ctx context.Context, data []T, clusterName string) (R, error)
	Name() string
}

// Analyzer combina Collector + Processor para análise completa.
type Analyzer[T any, R any] interface {
	Analyze(ctx context.Context, config *SyncConfig) (R, error)
	Name() string
}
