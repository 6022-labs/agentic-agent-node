package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
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

// spendConfigJson renders a matcher as the wire blob a request carries.
func spendConfigJson(t *testing.T, config *models.PaymentSpendRuleX402Config) json.RawMessage {
	t.Helper()

	raw, err := config.Marshal()
	assert.NoError(t, err)
	return raw
}

func TestWhenValidatingX402SpendRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a concrete network and asset with a payee pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it and lowercase the asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAsset("0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582").
				Build()

			request := &requests.SavePaymentSpendRuleRequest{Config: spendConfigJson(t, config)}
			assert.NoError(t, suite.sut.Validate(ctx, request))

			var sanitized models.PaymentSpendRuleX402Config
			assert.NoError(t, sanitized.Parse(request.Config))
			assert.Equal(t, "0x41e94eb019c0762f9bfcf9fb1e58725bfb0e7582", sanitized.Asset)
			assert.Equal(t, "eip155:80002", sanitized.Network)
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

	t.Run("Given the network is not a CAIP-2 id", func(t *testing.T) {
		t.Parallel()

		// The amounts are atomic units of one token on one chain, so neither may be a pattern.
		t.Run("Should refuse it", func(t *testing.T) {
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

	t.Run("Given the asset is not a contract address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
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

	t.Run("Given no payee pattern", func(t *testing.T) {
		t.Parallel()

		// An empty array matches nobody, so the rule could never authorize anything.
		t.Run("Should refuse rather than store a rule that never fires", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPayToPatterns(nil).Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.payToPatterns", validationError.Field)
		})
	})

	t.Run("Given a payee pattern that cannot compile", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0x[("}).
				Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.payToPatterns", validationError.Field)
		})
	})

	t.Run("Given a payee pattern spelled as a checksummed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be stored lowercased", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}).
				Build()

			request := &requests.SavePaymentSpendRuleRequest{Config: spendConfigJson(t, config)}
			assert.NoError(t, suite.sut.Validate(ctx, request))

			var sanitized models.PaymentSpendRuleX402Config
			assert.NoError(t, sanitized.Parse(request.Config))
			assert.Equal(t, []string{"0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"}, sanitized.PayToPatterns)
		})
	})

	t.Run("Given an unsupported transfer-method constraint", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
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

	t.Run("Given no config at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
		})
	})
}
