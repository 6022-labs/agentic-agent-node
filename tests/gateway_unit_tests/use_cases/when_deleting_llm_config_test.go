package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
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

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeNoContent, res.Outcome)
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

			res := suite.sut.Execute(context.Background())

			assert.Equal(t, usecase_result.OutcomeUnavailable, res.Outcome)
			assert.Contains(t, res.Message, "transport down")
		})
	})
}
