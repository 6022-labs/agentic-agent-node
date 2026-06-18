package services_test

import (
	"context"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentRuleTestingSuite struct {
	sut *services.HttpPaymentRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenUpdatingPaymentRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenUpdatingPaymentRuleTestingSuite{
		sut:        services.NewHttpPaymentRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the updated rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentPolicyRulesIdWithBodyWithResponse(gomock.Any(), "r1", "application/json", gomock.Any()).
				Return(&clients.PutInternalConfigPaymentPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"id":"r1","name":"Renamed"}`),
				}, nil)
		}

		t.Run("Should decode the updated rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Update(context.Background(), "r1", &models.ExternalPaymentRule{Name: "Renamed"})

			assert.NoError(t, err)
			assert.Equal(t, "Renamed", rule.Name)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigPaymentPolicyRulesIdWithBodyWithResponse(gomock.Any(), "r1", "application/json", gomock.Any()).
				Return(&clients.PutInternalConfigPaymentPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"rule_not_found"}`),
				}, nil)
		}

		t.Run("Should map the 404 to a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Update(context.Background(), "r1", &models.ExternalPaymentRule{})

			assert.Nil(t, rule)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})
}
