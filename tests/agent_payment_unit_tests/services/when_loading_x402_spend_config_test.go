package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenLoadingX402SpendConfigTestingSuite struct {
	sut *services.X402PaymentSpendMethodStrategy

	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
}

func WhenLoadingX402SpendConfigBeforeEach(t *testing.T) *WhenLoadingX402SpendConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController)

	return &WhenLoadingX402SpendConfigTestingSuite{
		sut: services.NewX402PaymentSpendMethodStrategy(mockPaymentSpendRuleX402ConfigRepository),

		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
	}
}

func TestWhenLoadingX402SpendConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the config lookup fails", func(t *testing.T) {
		t.Parallel()

		ruleId := uuid.New()
		initSuite := func(suite *WhenLoadingX402SpendConfigTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, ruleId).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402SpendConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(ctx, ruleId)

			assert.Nil(t, raw)
			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given the rule has no config", func(t *testing.T) {
		t.Parallel()

		ruleId := uuid.New()
		initSuite := func(suite *WhenLoadingX402SpendConfigTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, ruleId).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402SpendConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(ctx, ruleId)

			assert.NoError(t, err)
			assert.Nil(t, raw)
		})
	})

	t.Run("Given the rule has a config", func(t *testing.T) {
		t.Parallel()

		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
		initSuite := func(suite *WhenLoadingX402SpendConfigTestingSuite) {
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, config.PaymentSpendRuleId).
				Return(config, nil).
				Times(1)
		}

		t.Run("Should return it as the wire blob", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402SpendConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(ctx, config.PaymentSpendRuleId)

			assert.NoError(t, err)
			loaded := &models.PaymentSpendRuleX402Config{}
			assert.NoError(t, loaded.Parse(raw))
			assert.Equal(t, config.Network, loaded.Network)
			assert.Equal(t, config.Asset, loaded.Asset)
		})
	})
}
