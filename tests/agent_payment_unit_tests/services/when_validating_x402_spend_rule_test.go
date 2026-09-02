package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
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

	t.Run("Given every facet carries a valid pattern", func(t *testing.T) {
		t.Parallel()

		initRequest := func(t *testing.T) *requests.SavePaymentSpendRuleRequest {
			t.Helper()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{"  eip155:80002  "}).
				WithAssetTransferMethodPatterns([]string{"eip3009"}).
				Build()
			return &requests.SavePaymentSpendRuleRequest{Config: spendConfigJson(t, config)}
		}

		t.Run("Should accept it and trim each pattern", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			request := initRequest(t)

			assert.NoError(t, suite.sut.Validate(ctx, request))

			var sanitized models.PaymentSpendRuleX402Config
			assert.NoError(t, sanitized.Parse(request.Config))
			assert.Equal(t, []string{"eip155:80002"}, sanitized.NetworkPatterns)
		})

		// The receive strategy probes the chain for the asset's EIP-712 domain;
		// a matcher has no concrete asset, so this one must stay offline.
		t.Run("Should not persist anything while validating", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(0)

			assert.NoError(t, suite.sut.Validate(ctx, initRequest(t)))
		})
	})

	t.Run("Given a facet is empty", func(t *testing.T) {
		t.Parallel()

		// An empty array matches nothing, so it would deny the facet outright
		// rather than open it — refuse instead of storing a silent deny.
		emptyFacets := []struct {
			field  string
			mutate func(*models.PaymentSpendRuleX402Config)
		}{
			{"config.networkPatterns", func(c *models.PaymentSpendRuleX402Config) { c.NetworkPatterns = nil }},
			{"config.assetPatterns", func(c *models.PaymentSpendRuleX402Config) { c.AssetPatterns = nil }},
			{"config.payToPatterns", func(c *models.PaymentSpendRuleX402Config) { c.PayToPatterns = nil }},
			{
				"config.assetTransferMethodPatterns",
				func(c *models.PaymentSpendRuleX402Config) { c.AssetTransferMethodPatterns = nil },
			},
		}

		for _, emptyFacet := range emptyFacets {
			t.Run("Should refuse when "+emptyFacet.field+" is missing", func(t *testing.T) {
				t.Parallel()

				suite := WhenValidatingX402SpendRuleBeforeEach(t)
				config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()
				emptyFacet.mutate(config)

				err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
					Config: spendConfigJson(t, config),
				})

				var validationError *apperrors.ValidationError
				assert.ErrorAs(t, err, &validationError)
				assert.Equal(t, emptyFacet.field, validationError.Field)
			})
		}
	})

	t.Run("Given a pattern that cannot compile", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it rather than store a rule that never matches", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{"0x[("}).
				Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "config.networkPatterns", validationError.Field)
		})
	})

	t.Run("Given a blank pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402SpendRuleBeforeEach(t)
			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetPatterns([]string{"   "}).
				Build()

			err := suite.sut.Validate(ctx, &requests.SavePaymentSpendRuleRequest{
				Config: spendConfigJson(t, config),
			})

			assert.ErrorContains(t, err, "cannot be empty")
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
