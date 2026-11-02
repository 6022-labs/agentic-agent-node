package services_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingX402SpendRuleTestingSuite struct {
	sut *services.X402PaymentSpendMethodStrategy

	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
}

func WhenValidatingX402SpendRuleBeforeEach(t *testing.T) *WhenValidatingX402SpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController)

	return &WhenValidatingX402SpendRuleTestingSuite{
		sut: services.NewX402PaymentSpendMethodStrategy(
			mockPaymentSpendRuleX402ConfigRepository,
		),

		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
	}
}

// spendConfigJson renders a token config as the wire blob a request carries.
func spendConfigJson(t *testing.T, config *models.PaymentSpendRuleX402Config) json.RawMessage {
	t.Helper()

	raw, err := config.Marshal()
	assert.NoError(t, err)
	return raw
}

func TestWhenValidatingX402SpendRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given no config at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config", validationError.Field)
		})
	})

	t.Run("Given the network is not eip155", func(t *testing.T) {
		t.Parallel()

		// The amounts are atomic units of one token on one chain, so neither may be a pattern.
		t.Run("Should return a validation error on config.network", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithNetwork(".*").Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.network", validationError.Field)
		})
	})

	t.Run("Given the asset is not a hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithAsset(".*").Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.asset", validationError.Field)
		})
	})

	t.Run("Given an unsupported transfer-method constraint", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.assetTransferMethod", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			unsupported := enums.AssetTransferMethod("swap")
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
			config.AssetTransferMethod = &unsupported

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.assetTransferMethod", validationError.Field)
		})
	})

	// The network is part of the one-rule-per-token key, so a padded chain id must not make a second key.
	t.Run("Given a network whose chain id has leading zeros", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store the canonical network", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithNetwork("eip155:08453").Build()

			request := &requests.SavePaymentSpendRuleRequest{Config: spendConfigJson(t, config)}
			assert.NoError(t, suite.sut.Validate(ctx, request))

			var sanitized models.PaymentSpendRuleX402Config
			assert.NoError(t, sanitized.Parse(request.Config))
			assert.Equal(t, "eip155:8453", sanitized.Network)
		})
	})

	t.Run("Given a concrete network and asset", func(t *testing.T) {
		t.Parallel()

		// The asset is the lookup key, so every pasted case must store alike.
		t.Run("Should store the asset lowercased", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			asset := common_tests.RandomEthAddress().Hex()
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithAsset(asset).Build()

			request := &requests.SavePaymentSpendRuleRequest{Config: spendConfigJson(t, config)}
			assert.NoError(t, suite.sut.Validate(ctx, request))

			var sanitized models.PaymentSpendRuleX402Config
			assert.NoError(t, sanitized.Parse(request.Config))
			assert.Equal(t, strings.ToLower(asset), sanitized.Asset)
		})

		// Unlike receive, a spend rule only names the token, so validation never probes the chain.
		t.Run("Should not persist anything while validating", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(0)

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
			assert.NoError(t, suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			}))
		})
	})
}
