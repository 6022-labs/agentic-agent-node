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

			t.Run("Given a nested scope is started", func(t *testing.T) {
				t.Parallel()

				// A nested caller must not be able to keep or discard work it does not own.
				t.Run("Should join the outer scope rather than commit on its own", func(t *testing.T) {
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
					_, err = suite.connections.Update(inner.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, inner.Commit())

					assert.NoError(t, outer.Rollback())
					assert.Equal(t, original, suite.nameOf(t, connection.Id),
						"the inner commit must not have escaped the outer rollback")
				})
			})
		})
	}
}
