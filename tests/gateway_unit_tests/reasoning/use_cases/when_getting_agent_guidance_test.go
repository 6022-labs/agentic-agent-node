package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentGuidanceTestingSuite struct {
	sut *use_cases.GetAgentGuidance

	mockAgentGuidanceClient *interfaces_mocks.MockAgentGuidanceClientInterface
}

func WhenGettingAgentGuidanceBeforeEach(t *testing.T) *WhenGettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentGuidanceClient := interfaces_mocks.NewMockAgentGuidanceClientInterface(mockController)

	return &WhenGettingAgentGuidanceTestingSuite{
		sut:                     use_cases.NewGetAgentGuidance(mockAgentGuidanceClient),
		mockAgentGuidanceClient: mockAgentGuidanceClient,
	}
}

func TestWhenGettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the guidance", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			suite.mockAgentGuidanceClient.EXPECT().
				Get(gomock.Any()).
				Return(&models.ExternalAgentGuidancePayload{Content: "answer concisely"}, nil)
		}

		t.Run("Should return the content", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "answer concisely", resp.Content)
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			suite.mockAgentGuidanceClient.EXPECT().Get(gomock.Any()).Return(nil, errors.New("connection refused"))
		}

		t.Run("Should return runtime_unavailable", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})
}
