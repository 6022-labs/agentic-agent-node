package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenFindingX402ReceiveRuleForAssetTestingSuite struct {
	sut *services.X402PaymentReceiveMethodStrategy

	mockPaymentReceiveRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentReceiveRuleX402ConfigRepositoryInterface
}

func WhenFindingX402ReceiveRuleForAssetBeforeEach(t *testing.T) *WhenFindingX402ReceiveRuleForAssetTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentReceiveRuleX402ConfigRepositoryInterface(mockController)

	return &WhenFindingX402ReceiveRuleForAssetTestingSuite{
		sut: services.NewX402PaymentReceiveMethodStrategy(
			&runtime_common_settings.PaymentReceiveSettings{},
			&runtime_common_settings.EvmChainsSettings{},
			payment_interfaces_mocks.NewMockErc20ValidatorInterface(mockController),
			mockPaymentReceiveRuleX402ConfigRepository,
		),

		mockPaymentReceiveRuleX402ConfigRepository: mockPaymentReceiveRuleX402ConfigRepository,
	}
}

func receiveConfigJson(t *testing.T, config *models.PaymentReceiveRuleX402Config) json.RawMessage {
	t.Helper()

	raw, err := config.Marshal()
	assert.NoError(t, err)
	return raw
}

func TestWhenFindingX402ReceiveRuleForAsset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an invalid config", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402ReceiveRuleForAssetBeforeEach(t)

			_, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), json.RawMessage("{not-json"))

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config", validationError.Field)
		})
	})

	t.Run("Given reading the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFindingX402ReceiveRuleForAssetTestingSuite) {
			suite.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402ReceiveRuleForAssetBeforeEach(t)
			initSuite(suite)

			config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().Build()
			_, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), receiveConfigJson(t, config))

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given no rule prices the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFindingX402ReceiveRuleForAssetTestingSuite) {
			suite.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return no rule id", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402ReceiveRuleForAssetBeforeEach(t)
			initSuite(suite)

			config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().Build()
			ruleId, err := suite.sut.RuleIdForAsset(ctx, uuid.New(), receiveConfigJson(t, config))

			assert.NoError(t, err)
			assert.Equal(t, uuid.Nil, ruleId)
		})
	})

	t.Run("Given a rule already prices the token", func(t *testing.T) {
		t.Parallel()

		policyId := uuid.New()
		existing := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().WithPolicyId(policyId).Build()

		initSuite := func(suite *WhenFindingX402ReceiveRuleForAssetTestingSuite) {
			suite.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(gomock.Any(), policyId, existing.Network, existing.Asset).
				Return(existing, nil).
				Times(1)
		}

		t.Run("Should return that rule's id", func(t *testing.T) {
			t.Parallel()

			suite := WhenFindingX402ReceiveRuleForAssetBeforeEach(t)
			initSuite(suite)

			ruleId, err := suite.sut.RuleIdForAsset(ctx, policyId, receiveConfigJson(t, existing))

			assert.NoError(t, err)
			assert.Equal(t, existing.PaymentReceiveRuleId, ruleId)
		})
	})
}
