package runtime_db

import (
	"context"
	"sync"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"gorm.io/gorm"
)

// transactionKey carries the open Transaction so repositories join it without knowing it exists.
type transactionKey struct{}

// Transaction is one open database transaction; the context carries it until it ends.
type Transaction struct {
	mutex        sync.Mutex
	tx           *gorm.DB
	closed       bool
	rollbackOnly bool
}

func (t *Transaction) Commit() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	if t.rollbackOnly {
		if err := t.tx.Rollback().Error; err != nil {
			return err
		}
		return runtime_common_interfaces.ErrUnitOfWorkRolledBack
	}
	return t.tx.Commit().Error
}

func (t *Transaction) Rollback() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	return t.tx.Rollback().Error
}

// MarkRollbackOnly makes the next Commit roll back instead.
func (t *Transaction) MarkRollbackOnly() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.rollbackOnly = true
}

// open returns the live handle, or false once the transaction has ended.
func (t *Transaction) open() (*gorm.DB, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.tx, !t.closed
}

// BeginTransaction opens a transaction and returns a context whose repository calls join it.
func (c *DatabaseClient) BeginTransaction(ctx context.Context) (context.Context, *Transaction, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	transaction := &Transaction{tx: tx}
	return context.WithValue(ctx, transactionKey{}, transaction), transaction, nil
}

// TransactionFrom returns the still-open transaction the context carries.
func (c *DatabaseClient) TransactionFrom(ctx context.Context) (*Transaction, bool) {
	transaction, ok := ctx.Value(transactionKey{}).(*Transaction)
	if !ok {
		return nil, false
	}
	if _, open := transaction.open(); !open {
		return nil, false
	}
	return transaction, true
}
