package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingAgentGuidanceTestingSuite struct {
	sut *use_cases.SetAgentGuidance

	mockAgentGuidanceClient *interfaces_mocks.MockAgentGuidanceClientInterface
}

func WhenSettingAgentGuidanceBeforeEach(t *testing.T) *WhenSettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentGuidanceClient := interfaces_mocks.NewMockAgentGuidanceClientInterface(mockController)

	return &WhenSettingAgentGuidanceTestingSuite{
		sut:                     use_cases.NewSetAgentGuidance(mockAgentGuidanceClient),
		mockAgentGuidanceClient: mockAgentGuidanceClient,
	}
}

func TestWhenSettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			suite.mockAgentGuidanceClient.EXPECT().
				Set(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("connection refused"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SetAgentGuidanceRequest{Content: "x"})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			suite.mockAgentGuidanceClient.EXPECT().
				Set(gomock.Any(), &models.ExternalAgentGuidancePayload{Content: "be concise"}).
				Return(&models.ExternalAgentGuidancePayload{Content: "be concise"}, nil)
		}

		t.Run("Should forward the content and echo the runtime's value", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.SetAgentGuidanceRequest{Content: "be concise"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "be concise", resp.Content)
		})
	})
}
