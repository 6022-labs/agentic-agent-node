package services_test

import (
	"context"
	"testing"

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

			t.Run("Given the work is committed", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep every write made through the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenCommittingAUnitOfWorkBeforeEach(tc)
					defer WhenCommittingAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Commit())

					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})
		})
	}
}
