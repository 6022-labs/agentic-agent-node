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

func TestWhenListingPaymentSpendRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.Nil(t, rules)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.Nil(t, rules)
			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime lists rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentSpendPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"rules":[{"id":"r1","name":"USDC on Base","maxAmountPerPayment":"1000",` +
							`"windows":[{"windowSeconds":86400,"maxAmount":"50000","spent":"4000",` +
							`"resetsAt":"2026-09-10T12:00:00Z"}],"method":"x402",` +
							`"config":{"network":"eip155:8453"}}]}`,
					),
				}, nil)
		}

		t.Run("Should decode them with their windows and token", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, rules, 1) {
				assert.Equal(t, "r1", rules[0].Id)
				assert.Equal(t, "1000", *rules[0].MaxAmountPerPayment)
				if assert.Len(t, rules[0].Windows, 1) {
					assert.Equal(t, 86400, rules[0].Windows[0].WindowSeconds)
					assert.Equal(t, "50000", rules[0].Windows[0].MaxAmount)
					assert.Equal(t, "4000", rules[0].Windows[0].Spent)
					if assert.NotNil(t, rules[0].Windows[0].ResetsAt) {
						assert.Equal(t, 2026, rules[0].Windows[0].ResetsAt.Year())
					}
				}
				assert.Contains(t, string(rules[0].Config), "eip155:8453")
			}
		})
	})
}
