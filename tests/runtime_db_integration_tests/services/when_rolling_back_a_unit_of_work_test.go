package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

type WhenRollingBackAUnitOfWorkTestingSuite struct {
	*unitOfWorkTestingSuite
}

func WhenRollingBackAUnitOfWorkBeforeEach(tc *common_tests.DatabaseTestCase) *WhenRollingBackAUnitOfWorkTestingSuite {
	return &WhenRollingBackAUnitOfWorkTestingSuite{unitOfWorkTestingSuite: newUnitOfWorkTestingSuite(tc)}
}

func WhenRollingBackAUnitOfWorkAfterEach(suite *WhenRollingBackAUnitOfWorkTestingSuite) {
	suite.close()
}

func TestWhenRollingBackAUnitOfWork(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a rollback follows a commit", func(t *testing.T) {
				t.Parallel()

				// `defer Rollback()` sits on the committed path too, so it must be a no-op.
				t.Run("Should be a no-op that keeps the committed work", func(t *testing.T) {
					t.Parallel()
					suite := WhenRollingBackAUnitOfWorkBeforeEach(tc)
					defer WhenRollingBackAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)

					assert.NoError(t, unitOfWork.Commit())
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the scope is rolled back", func(t *testing.T) {
				t.Parallel()

				t.Run("Should discard every write made through it", func(t *testing.T) {
					t.Parallel()
					suite := WhenRollingBackAUnitOfWorkBeforeEach(tc)
					defer WhenRollingBackAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "discarded"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the caller never commits", func(t *testing.T) {
				t.Parallel()

				// The pattern every caller uses: defer Rollback, return early on refusal.
				t.Run("Should discard the work when the deferred rollback runs", func(t *testing.T) {
					t.Parallel()
					suite := WhenRollingBackAUnitOfWorkBeforeEach(tc)
					defer WhenRollingBackAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					func() {
						unitOfWork, err := suite.sut.Start(context.Background())
						assert.NoError(t, err)
						defer func() { _ = unitOfWork.Rollback() }()

						connection.Name = "abandoned"
						_, err = suite.connections.Update(unitOfWork.Context(), connection)
						assert.NoError(t, err)
					}()

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})
		})
	}
}
