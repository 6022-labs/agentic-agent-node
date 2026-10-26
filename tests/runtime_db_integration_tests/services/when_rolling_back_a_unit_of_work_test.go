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

			t.Run("Given the scope is already closed", func(t *testing.T) {
				t.Parallel()

				// `defer Rollback()` sits on the committed path too.
				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenRollingBackAUnitOfWorkBeforeEach(tc)
					defer WhenRollingBackAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Commit())

					assert.NoError(t, unitOfWork.Rollback())
					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the scope is open", func(t *testing.T) {
				t.Parallel()

				t.Run("Should discard every write made through the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenRollingBackAUnitOfWorkBeforeEach(tc)
					defer WhenRollingBackAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "discarded"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})
		})
	}
}
