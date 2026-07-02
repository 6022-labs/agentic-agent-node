package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingX402ConfigTestingSuite struct {
	sut *services.X402PaymentMethodStrategy

	mockConfigRepository *payment_repositories_mocks.MockPaymentRuleX402ConfigRepositoryInterface
}

func WhenSavingX402ConfigBeforeEach(t *testing.T) *WhenSavingX402ConfigTestingSuite {
	mockController := gomock.NewController(t)

	facilitators := &runtime_common_settings.PaymentFacilitatorsSettings{}
	mockErc20Validator := payment_interfaces_mocks.NewMockErc20ValidatorInterface(mockController)
	mockConfigRepository := payment_repositories_mocks.NewMockPaymentRuleX402ConfigRepositoryInterface(mockController)

	return &WhenSavingX402ConfigTestingSuite{
		sut:                  services.NewX402PaymentMethodStrategy(facilitators, &runtime_common_settings.EvmChainsSettings{}, mockErc20Validator, mockConfigRepository),
		mockConfigRepository: mockConfigRepository,
	}
}

func TestWhenSavingX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid config blob", func(t *testing.T) {
		t.Parallel()

		ruleId := uuid.New()

		initSuite := func(suite *WhenSavingX402ConfigTestingSuite) {
			suite.mockConfigRepository.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, config *models.PaymentRuleX402Config) error {
					assert.Equal(t, ruleId, config.PaymentRuleId)
					assert.Equal(t, "eip155:80002", config.Network)
					assert.Equal(t, []string{".*"}, config.AddressPatterns)
					return nil
				}).
				Times(1)
		}

		t.Run("Should upsert the typed config keyed by the rule id", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402ConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.SaveConfig(context.Background(), ruleId, json.RawMessage(`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`))

			assert.NoError(t, err)
		})
	})

	t.Run("Given a malformed config blob", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402ConfigBeforeEach(t)

			err := suite.sut.SaveConfig(context.Background(), uuid.New(), json.RawMessage("{not-json"))

			assert.Error(t, err)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingX402ConfigTestingSuite) {
			suite.mockConfigRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(assert.AnError).Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingX402ConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.SaveConfig(context.Background(), uuid.New(), json.RawMessage(`{"network":"eip155:80002","addressPatterns":[".*"]}`))

			assert.Error(t, err)
		})
	})
}
