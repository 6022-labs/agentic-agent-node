package services_test

import (
	"context"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentSpendPolicyTestingSuite struct {
	sut *services.HttpPaymentSpendPolicyClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenGettingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingPaymentSpendPolicyTestingSuite{
		sut:        services.NewHttpPaymentSpendPolicyClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the policy head", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"id":"a","agentId":"b","agentWalletAddress":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}`,
					),
				}, nil)
		}

		t.Run("Should decode the head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "a", payload.Id)
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should return a conflict error carrying the runtime code", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.Nil(t, payload)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"get_failed"}`),
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.Nil(t, payload)
			assert.Error(t, err)
		})
	})
}
