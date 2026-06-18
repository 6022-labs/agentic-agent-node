package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingLlmConfigTestingSuite struct {
	sut *use_cases.DeleteLlmConfig

	mockClient *interfaces_mocks.MockLLMConfigClientInterface
}

func WhenDeletingLlmConfigBeforeEach(t *testing.T) *WhenDeletingLlmConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockLLMConfigClientInterface(mockController)

	sut := use_cases.NewDeleteLlmConfig(mockClient)

	return &WhenDeletingLlmConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenDeletingLlmConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the client succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any()).Return(nil)
		}

		t.Run("Should return NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any()).Return(errors.New("transport down"))
		}

		t.Run("Should return Unavailable carrying the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "transport down")
		})
	})
}
