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

type WhenUpdatingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentReceiveRule

	mockPaymentReceiveRuleClient *interfaces_mocks.MockPaymentReceiveRuleClientInterface
}

func WhenUpdatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleClient := interfaces_mocks.NewMockPaymentReceiveRuleClientInterface(mockController)

	return &WhenUpdatingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentReceiveRule(mockPaymentReceiveRuleClient),

		mockPaymentReceiveRuleClient: mockPaymentReceiveRuleClient,
	}
}

func TestWhenUpdatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	ruleId := uuid.New()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentReceiveRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewValidationError("name", "name is invalid"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentReceiveRuleRequest{Id: ruleId})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("rule_not_found", "no such rule")).
				Times(1)
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentReceiveRuleRequest{Id: ruleId})

			assert.Nil(t, rule)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewConflictError("asset_already_has_rule", "another rule already prices this token"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentReceiveRuleRequest{Id: ruleId})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentReceiveRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime accepts the updated rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Update(gomock.Any(), ruleId.String(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, rule *models.ExternalPaymentReceiveRule) (*models.ExternalPaymentReceiveRule, error) {
					assert.Equal(t, "Renamed", rule.Name)
					assert.Equal(t, []models.ExternalPaymentReceiveRulePrice{
						{Price: "0", AddressPatterns: []string{"0xowner"}},
						{Price: "1000", AddressPatterns: []string{".*"}},
					}, rule.Prices)
					rule.Id = id
					return rule, nil
				}).
				Times(1)
		}

		t.Run("Should forward the rule and return it", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(
				context.Background(),
				&requests.UpdatePaymentReceiveRuleRequest{
					Id: ruleId,
					SavePaymentReceiveRuleRequest: requests.SavePaymentReceiveRuleRequest{
						Name:   "Renamed",
						Method: "x402",
						Prices: []requests.SavePaymentReceiveRulePriceRequest{
							{Price: "0", AddressPatterns: []string{"0xowner"}},
							{Price: "1000", AddressPatterns: []string{".*"}},
						},
						Config: json.RawMessage(`{"network":"eip155:80002"}`),
					},
				},
			)

			assert.NoError(t, err)
			assert.Equal(t, ruleId.String(), rule.Id)
		})
	})
}
