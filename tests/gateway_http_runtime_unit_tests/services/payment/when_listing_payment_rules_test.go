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

type WhenListingPaymentRulesTestingSuite struct {
	sut *services.HttpPaymentRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingPaymentRulesBeforeEach(t *testing.T) *WhenListingPaymentRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingPaymentRulesTestingSuite{
		sut:        services.NewHttpPaymentRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingPaymentRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime lists rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"rules":[{"id":"r1","name":"Public price","method":"x402",` +
							`"config":{"network":"eip155:80002","addressPatterns":[".*"]}}]}`,
					),
				}, nil)
		}

		t.Run("Should decode them", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, rules, 1) {
				assert.Equal(t, "r1", rules[0].Id)
				assert.Equal(t, "x402", rules[0].Method)
				assert.JSONEq(t, `{"network":"eip155:80002","addressPatterns":[".*"]}`, string(rules[0].Config))
			}
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should return a conflict error carrying the runtime code", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.Nil(t, rules)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the listing fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentPolicyRulesWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			rules, err := suite.sut.List(context.Background())

			assert.Nil(t, rules)
			assert.Error(t, err)
		})
	})
}
