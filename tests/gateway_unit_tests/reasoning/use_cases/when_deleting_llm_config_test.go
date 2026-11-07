package use_cases_test

import (
	"context"
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

	t.Run("Given the runtime answers 404", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewNotFoundError("not_found", "resource not found")
		initSuite := func(suite *WhenDeletingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any()).Return(runtimeErr)
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewUnavailableError("runtime_unavailable", "the runtime is unavailable", nil)
		initSuite := func(suite *WhenDeletingLlmConfigTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any()).Return(runtimeErr)
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLlmConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
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
}
