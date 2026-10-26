package utils

import (
	"context"

	"gorm.io/gorm"
)

// unitOfWorkKey carries the open transaction so repositories join it without knowing it exists.
type unitOfWorkKey struct{}

// ContextWithUnitOfWork returns a context whose repository calls join tx.
func ContextWithUnitOfWork(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, unitOfWorkKey{}, tx)
}

// UnitOfWorkFrom returns the transaction the context carries, if any.
func UnitOfWorkFrom(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(unitOfWorkKey{}).(*gorm.DB)
	return tx, ok
}
