package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/identity_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentSoulTestingSuite struct {
	sut *use_cases.GetAgentSoul

	mockAgentSoulClient *interfaces_mocks.MockAgentSoulClientInterface
}

func WhenGettingAgentSoulBeforeEach(t *testing.T) *WhenGettingAgentSoulTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentSoulClient := interfaces_mocks.NewMockAgentSoulClientInterface(mockController)

	return &WhenGettingAgentSoulTestingSuite{
		sut:                 use_cases.NewGetAgentSoul(mockAgentSoulClient),
		mockAgentSoulClient: mockAgentSoulClient,
	}
}

func TestWhenGettingAgentSoul(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime reports no identity", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentSoulClient.EXPECT().Get(gomock.Any()).Return(nil, identity_interfaces.ErrAgentSoulNotFound)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "identity_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewConflictError("conflict", "resource conflict")
		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentSoulClient.EXPECT().Get(gomock.Any()).Return(nil, runtimeErr)
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewUnavailableError("runtime_unavailable", "the runtime is unavailable", nil)
		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentSoulClient.EXPECT().Get(gomock.Any()).Return(nil, runtimeErr)
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime returns the soul", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentSoulClient.EXPECT().
				Get(gomock.Any()).
				Return(&models.ExternalAgentSoulPayload{Soul: "persona"}, nil)
		}

		t.Run("Should return the soul", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "persona", resp.Soul)
		})
	})
}
