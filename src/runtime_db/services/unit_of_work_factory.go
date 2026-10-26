package services

import (
	"context"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
)

// UnitOfWorkFactory is the persistence side of runtime_common's atomic-scope port.
type UnitOfWorkFactory struct {
	client *runtime_db.DatabaseClient
}

func NewUnitOfWorkFactory(client *runtime_db.DatabaseClient) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{client: client}
}

func (f *UnitOfWorkFactory) Start(ctx context.Context) (runtime_common_interfaces.UnitOfWorkInterface, error) {
	if transaction, ok := f.client.TransactionFrom(ctx); ok {
		return &joinedUnitOfWork{ctx: ctx, transaction: transaction}, nil
	}

	scopedCtx, transaction, err := f.client.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return &unitOfWork{ctx: scopedCtx, transaction: transaction}, nil
}

type unitOfWork struct {
	ctx         context.Context
	transaction *runtime_db.Transaction
}

func (u *unitOfWork) Context() context.Context { return u.ctx }
func (u *unitOfWork) Commit() error            { return u.transaction.Commit() }
func (u *unitOfWork) Rollback() error          { return u.transaction.Rollback() }

// joinedUnitOfWork cannot end the caller's scope; its Rollback only dooms the outer Commit.
type joinedUnitOfWork struct {
	ctx         context.Context
	transaction *runtime_db.Transaction
}

func (u *joinedUnitOfWork) Context() context.Context { return u.ctx }
func (u *joinedUnitOfWork) Commit() error            { return nil }

func (u *joinedUnitOfWork) Rollback() error {
	u.transaction.MarkRollbackOnly()
	return nil
}
