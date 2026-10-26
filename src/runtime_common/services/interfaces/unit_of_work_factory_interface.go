package interfaces

import "context"

// UnitOfWorkFactoryInterface opens an atomic scope for any domain that needs one.
type UnitOfWorkFactoryInterface interface {
	// Start joins the open scope the context carries, else opens a new one.
	Start(ctx context.Context) (UnitOfWorkInterface, error)
}
