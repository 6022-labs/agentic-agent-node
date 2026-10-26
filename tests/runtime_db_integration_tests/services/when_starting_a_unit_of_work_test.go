package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

type WhenStartingAUnitOfWorkTestingSuite struct {
	*unitOfWorkTestingSuite
}

func WhenStartingAUnitOfWorkBeforeEach(tc *common_tests.DatabaseTestCase) *WhenStartingAUnitOfWorkTestingSuite {
	return &WhenStartingAUnitOfWorkTestingSuite{unitOfWorkTestingSuite: newUnitOfWorkTestingSuite(tc)}
}

func WhenStartingAUnitOfWorkAfterEach(suite *WhenStartingAUnitOfWorkTestingSuite) {
	suite.close()
}

func TestWhenStartingAUnitOfWork(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a scope is already in flight", func(t *testing.T) {
				t.Parallel()

				// A nested caller must not keep work it does not own.
				t.Run("Should join the outer scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenStartingAUnitOfWorkBeforeEach(tc)
					defer WhenStartingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					outer, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					inner, err := suite.sut.Start(outer.Context())
					assert.NoError(t, err)
					connection.Name = "written by the inner scope"
					_, err = suite.mcpConnectionRepository.Update(inner.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, inner.Commit())

					assert.NoError(t, outer.Rollback())
					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the context carries a scope that already ended", func(t *testing.T) {
				t.Parallel()

				// Joining the ended scope would make this Commit a silent no-op.
				t.Run("Should open a new scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenStartingAUnitOfWorkBeforeEach(tc)
					defer WhenStartingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					ended, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					assert.NoError(t, ended.Commit())

					fresh, err := suite.sut.Start(ended.Context())
					assert.NoError(t, err)
					connection.Name = "kept by the new scope"
					_, err = suite.mcpConnectionRepository.Update(fresh.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, fresh.Commit())

					assert.Equal(t, "kept by the new scope", suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given no scope is in flight", func(t *testing.T) {
				t.Parallel()

				t.Run("Should open a scope repositories write through", func(t *testing.T) {
					t.Parallel()
					suite := WhenStartingAUnitOfWorkBeforeEach(tc)
					defer WhenStartingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "written through the scope"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})
		})
	}
}
