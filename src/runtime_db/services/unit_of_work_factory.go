package services

import (
	"context"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/utils"
	"gorm.io/gorm"
)

// UnitOfWorkFactory is the persistence side of runtime_common's atomic-scope port.
type UnitOfWorkFactory struct {
	client *runtime_db.DatabaseClient
}

func NewUnitOfWorkFactory(client *runtime_db.DatabaseClient) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{client: client}
}

func (f *UnitOfWorkFactory) Start(ctx context.Context) (runtime_common_interfaces.UnitOfWorkInterface, error) {
	// Already inside one: hand back a handle that cannot end the caller's scope.
	if _, ok := utils.UnitOfWorkFrom(ctx); ok {
		return &joinedUnitOfWork{ctx: ctx}, nil
	}

	tx := f.client.Begin(ctx)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &unitOfWork{tx: tx, ctx: utils.ContextWithUnitOfWork(ctx, tx)}, nil
}

type unitOfWork struct {
	tx     *gorm.DB
	ctx    context.Context
	closed bool
}

func (u *unitOfWork) Context() context.Context {
	return u.ctx
}

func (u *unitOfWork) Commit() error {
	if u.closed {
		return nil
	}
	u.closed = true
	return u.tx.Commit().Error
}

func (u *unitOfWork) Rollback() error {
	if u.closed {
		return nil
	}
	u.closed = true
	return u.tx.Rollback().Error
}

// joinedUnitOfWork defers to the scope already in flight, so a nested caller cannot
// commit or discard work it does not own.
type joinedUnitOfWork struct {
	ctx context.Context
}

func (u *joinedUnitOfWork) Context() context.Context { return u.ctx }
func (u *joinedUnitOfWork) Commit() error            { return nil }
func (u *joinedUnitOfWork) Rollback() error          { return nil }
