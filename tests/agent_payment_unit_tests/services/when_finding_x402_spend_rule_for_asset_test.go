package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenFindingX402SpendRuleForAssetTestingSuite struct {
	sut *services.X402PaymentSpendMethodStrategy

	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
}

func WhenFindingX402SpendRuleForAssetBeforeEach(t *testing.T) *WhenFindingX402SpendRuleForAssetTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController)

	return &WhenFindingX402SpendRuleForAssetTestingSuite{
		sut: services.NewX402PaymentSpendMethodStrategy(mockPaymentSpendRuleX402ConfigRepository),

		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
	}
}

func TestWhenFindingX402SpendRuleForAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a malformed config blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402SpendRuleForAssetBeforeEach(t)

			_, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), json.RawMessage("{not-json"))

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config", validationError.Field)
		})
	})

	t.Run("Given reading the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFindingX402SpendRuleForAssetTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402SpendRuleForAssetBeforeEach(t)
			initSuite(suite)

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
			_, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), spendConfigJson(t, config))

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given no rule spends the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFindingX402SpendRuleForAssetTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return no rule id", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402SpendRuleForAssetBeforeEach(t)
			initSuite(suite)

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
			ruleId, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), spendConfigJson(t, config))

			assert.NoError(t, err)
			assert.Equal(t, uuid.Nil, ruleId)
		})
	})

	t.Run("Given a rule already spends the token", func(t *testing.T) {
		t.Parallel()

		policyId := uuid.New()
		existing := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPolicyId(policyId).Build()

		initSuite := func(suite *WhenFindingX402SpendRuleForAssetTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), policyId, existing.Network, existing.Asset).
				Return(existing, nil).
				Times(1)
		}

		t.Run("Should return that rule's id", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402SpendRuleForAssetBeforeEach(t)
			initSuite(suite)

			ruleId, err := suite.sut.RuleIdForAsset(ctx, policyId, spendConfigJson(t, existing))

			assert.NoError(t, err)
			assert.Equal(t, existing.PaymentSpendRuleId, ruleId)
		})
	})
}
