package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentReceivePolicyTestingSuite struct {
	sut *services.HttpPaymentReceivePolicyClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingPaymentReceivePolicyBeforeEach(t *testing.T) *WhenGettingPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingPaymentReceivePolicyTestingSuite{
		sut:        services.NewHttpPaymentReceivePolicyClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentReceivePolicyWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentReceivePolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentReceivePolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.Nil(t, payload)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentReceivePolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentReceivePolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"get_failed"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.Nil(t, payload)
			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the policy head", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentReceivePolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentReceivePolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"id":"a","agentId":"b","agentWalletAddress":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}`,
					),
				}, nil)
		}

		t.Run("Should decode the head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "a", payload.Id)
			assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", payload.AgentWalletAddress)
		})
	})
}
