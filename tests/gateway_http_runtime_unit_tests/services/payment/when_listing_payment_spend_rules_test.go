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

type WhenListingPaymentSpendRulesTestingSuite struct {
	sut *services.HttpPaymentSpendRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingPaymentSpendRulesBeforeEach(t *testing.T) *WhenListingPaymentSpendRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingPaymentSpendRulesTestingSuite{
		sut:        services.NewHttpPaymentSpendRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingPaymentSpendRulesOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime lists spend rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"rules":[{"id":"r1","name":"Anything, unbounded","maxAmountPerPayment":"1000",` +
							`"windowSeconds":86400,"maxAmountPerWindow":"50000","windowSpent":"4000",` +
							`"windowResetsAt":"2026-09-10T12:00:00Z","method":"x402",` +
							`"config":{"networkPatterns":[".*"],"assetPatterns":[".*"]}}]}`,
					),
				}, nil)
		}

		t.Run("Should decode them with their matcher", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, rules, 1) {
				assert.Equal(t, "r1", rules[0].Id)
				assert.Equal(t, "1000", rules[0].MaxAmountPerPayment)
				assert.Equal(t, 86400, rules[0].WindowSeconds)
				assert.Equal(t, "50000", rules[0].MaxAmountPerWindow)
				assert.Equal(t, "4000", rules[0].WindowSpent)
				if assert.NotNil(t, rules[0].WindowResetsAt) {
					assert.Equal(t, 2026, rules[0].WindowResetsAt.Year())
				}
				assert.Contains(t, string(rules[0].Config), "networkPatterns")
			}
		})
	})

	t.Run("Given the runtime rejects the request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should map the status onto a domain error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.Nil(t, rules)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
		})
	})
}
