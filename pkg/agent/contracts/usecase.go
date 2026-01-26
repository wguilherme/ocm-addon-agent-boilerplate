package contracts

import "context"

// UseCase → operação de negócio que recebe input I e retorna output O.
// Separa orquestração de infraestrutura (Clean Architecture).
type UseCase[I any, O any] interface {
	Perform(ctx context.Context, input I) (O, error)
}
