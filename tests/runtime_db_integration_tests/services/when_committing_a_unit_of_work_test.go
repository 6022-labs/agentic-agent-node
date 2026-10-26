package services_test

import (
	"context"
	"testing"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

type WhenCommittingAUnitOfWorkTestingSuite struct {
	*unitOfWorkTestingSuite
}

func WhenCommittingAUnitOfWorkBeforeEach(tc *common_tests.DatabaseTestCase) *WhenCommittingAUnitOfWorkTestingSuite {
	return &WhenCommittingAUnitOfWorkTestingSuite{unitOfWorkTestingSuite: newUnitOfWorkTestingSuite(tc)}
}

func WhenCommittingAUnitOfWorkAfterEach(suite *WhenCommittingAUnitOfWorkTestingSuite) {
	suite.close()
}

func TestWhenCommittingAUnitOfWork(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the scope is already closed", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenCommittingAUnitOfWorkBeforeEach(tc)
					defer WhenCommittingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "discarded"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Rollback())

					assert.NoError(t, unitOfWork.Commit())
					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given a joined scope rolled back", func(t *testing.T) {
				t.Parallel()

				t.Run("Should fail with ErrUnitOfWorkRolledBack", func(t *testing.T) {
					t.Parallel()
					suite := WhenCommittingAUnitOfWorkBeforeEach(tc)
					defer WhenCommittingAUnitOfWorkAfterEach(suite)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					inner, err := suite.sut.Start(unitOfWork.Context())
					assert.NoError(t, err)
					assert.NoError(t, inner.Rollback())

					assert.ErrorIs(t, unitOfWork.Commit(), runtime_common_interfaces.ErrUnitOfWorkRolledBack)
				})

				t.Run("Should discard every write made through the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenCommittingAUnitOfWorkBeforeEach(tc)
					defer WhenCommittingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "discarded"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					inner, err := suite.sut.Start(unitOfWork.Context())
					assert.NoError(t, err)
					assert.NoError(t, inner.Rollback())
					_ = unitOfWork.Commit()

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the scope is committed", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep every write made through the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenCommittingAUnitOfWorkBeforeEach(tc)
					defer WhenCommittingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.mcpConnectionRepository.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Commit())

					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})
		})
	}
}
