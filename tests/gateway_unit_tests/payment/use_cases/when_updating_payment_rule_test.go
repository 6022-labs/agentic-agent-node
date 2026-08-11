package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentRule

	mockPaymentRuleClient *interfaces_mocks.MockPaymentRuleClientInterface
}

func WhenUpdatingPaymentRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleClient := interfaces_mocks.NewMockPaymentRuleClientInterface(mockController)

	return &WhenUpdatingPaymentRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentRule(mockPaymentRuleClient),

		mockPaymentRuleClient: mockPaymentRuleClient,
	}
}

func TestWhenUpdatingPaymentRule(t *testing.T) {
	t.Parallel()

	ruleId := uuid.New()

	t.Run("Given the runtime accepts the updated rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, rule *models.ExternalPaymentRule) (*models.ExternalPaymentRule, error) {
					assert.Equal(t, "Renamed", rule.Name)
					rule.Id = id
					return rule, nil
				}).
				Times(1)
		}

		t.Run("Should forward the rule and return it", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentRuleRequest{Id: ruleId, SavePaymentRuleRequest: requests.SavePaymentRuleRequest{
					Name:   "Renamed",
					Method: "x402",
					Config: json.RawMessage(`{"network":"eip155:80002"}`),
				}},
			)

			assert.NoError(t, err)
			assert.Equal(t, ruleId.String(), rule.Id)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("rule_not_found", "no such rule")).
				Times(1)
		}

		t.Run("Should pass the not found error through", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentRuleRequest{Id: ruleId})

			assert.Nil(t, rule)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})
}
