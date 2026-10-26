package interfaces

import (
	"context"
	"errors"
)

// ErrUnitOfWorkRolledBack is returned by Commit when a joined scope rolled the work back.
var ErrUnitOfWorkRolledBack = errors.New("unit of work rolled back by a nested scope")

// UnitOfWorkInterface is one atomic scope, owned by a single goroutine. It tracks no changes:
// repositories called with Context() join it, and the caller decides whether it is kept.
type UnitOfWorkInterface interface {
	// Context carries the scope; repository calls made with it join the transaction.
	Context() context.Context
	// Commit keeps the work, or fails with ErrUnitOfWorkRolledBack if a joined scope rolled back.
	Commit() error
	// Rollback discards the work, including the outer scope's when called on a joined one.
	Rollback() error
}
