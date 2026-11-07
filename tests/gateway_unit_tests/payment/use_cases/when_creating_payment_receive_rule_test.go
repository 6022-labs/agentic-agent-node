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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.CreatePaymentReceiveRule

	mockPaymentReceiveRuleClient *interfaces_mocks.MockPaymentReceiveRuleClientInterface
}

func WhenCreatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenCreatingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleClient := interfaces_mocks.NewMockPaymentReceiveRuleClientInterface(mockController)

	return &WhenCreatingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewCreatePaymentReceiveRule(mockPaymentReceiveRuleClient),

		mockPaymentReceiveRuleClient: mockPaymentReceiveRuleClient,
	}
}

func TestWhenCreatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SavePaymentReceiveRuleRequest{})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewValidationError("asset", "no ERC-20 token answers")).
				Times(1)
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentReceiveRuleRequest{})

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewConflictError("asset_already_has_rule", "another rule already prices this token"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SavePaymentReceiveRuleRequest{})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SavePaymentReceiveRuleRequest{})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime accepts the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.ExternalPaymentReceiveRule) (*models.ExternalPaymentReceiveRule, error) {
					assert.Equal(t, "Public price", rule.Name)
					assert.Equal(t, []models.ExternalPaymentReceiveRulePrice{
						{Price: "0", AddressPatterns: []string{"0xowner"}},
						{Price: "1000", AddressPatterns: []string{".*"}},
					}, rule.Prices)
					assert.Equal(t, "x402", rule.Method)
					assert.JSONEq(t, `{"network":"eip155:80002"}`, string(rule.Config))
					rule.Id = "rule-id"
					return rule, nil
				}).
				Times(1)
		}

		t.Run("Should forward the rule and return it", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentReceiveRuleRequest{
				Name:   "Public price",
				Method: "x402",
				Prices: []requests.SavePaymentReceiveRulePriceRequest{
					{Price: "0", AddressPatterns: []string{"0xowner"}},
					{Price: "1000", AddressPatterns: []string{".*"}},
				},
				Config: json.RawMessage(`{"network":"eip155:80002"}`),
			})

			assert.NoError(t, err)
			assert.Equal(t, "rule-id", rule.Id)
		})
	})
}
