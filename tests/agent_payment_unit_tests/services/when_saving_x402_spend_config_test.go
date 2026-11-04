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

type WhenSavingX402SpendConfigTestingSuite struct {
	sut *services.X402PaymentSpendMethodStrategy

	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
}

func WhenSavingX402SpendConfigBeforeEach(t *testing.T) *WhenSavingX402SpendConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController)

	return &WhenSavingX402SpendConfigTestingSuite{
		sut: services.NewX402PaymentSpendMethodStrategy(mockPaymentSpendRuleX402ConfigRepository),

		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
	}
}

func TestWhenSavingX402SpendConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a malformed config blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402SpendConfigBeforeEach(t)

			err := suite.sut.SaveConfig(ctx, uuid.New(), uuid.New(), json.RawMessage("{not-json"))

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config", validationError.Field)
		})
	})

	t.Run("Given saving the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingX402SpendConfigTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				Return(assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402SpendConfigBeforeEach(t)
			initSuite(suite)

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
			err := suite.sut.SaveConfig(ctx, uuid.New(), uuid.New(), spendConfigJson(t, config))

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given a valid config blob", func(t *testing.T) {
		t.Parallel()

		policyId := uuid.New()
		ruleId := uuid.New()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

		initSuite := func(suite *WhenSavingX402SpendConfigTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				Upsert(gomock.Any(), model_builders.NewPaymentSpendRuleX402ConfigBuilder().
					WithPaymentSpendRuleId(ruleId).
					WithPolicyId(policyId).
					WithNetwork(config.Network).
					WithAsset(config.Asset).
					Build()).
				Return(nil).
				Times(1)
		}

		// The policy id is copied so the database can hold one rule per token.
		t.Run("Should upsert the token keyed by the rule and its policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402SpendConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.SaveConfig(ctx, policyId, ruleId, spendConfigJson(t, config))

			assert.NoError(t, err)
		})
	})
}
