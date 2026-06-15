package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenLoadingX402ConfigTestingSuite struct {
	sut *services.X402PaymentMethodStrategy

	mockConfigRepository *payment_repositories_mocks.MockPaymentRuleX402ConfigRepositoryInterface
}

func WhenLoadingX402ConfigBeforeEach(t *testing.T) *WhenLoadingX402ConfigTestingSuite {
	mockController := gomock.NewController(t)

	facilitators := &runtime_common_settings.PaymentFacilitatorsSettings{}
	mockErc20Validator := payment_interfaces_mocks.NewMockErc20ValidatorInterface(mockController)
	mockConfigRepository := payment_repositories_mocks.NewMockPaymentRuleX402ConfigRepositoryInterface(mockController)

	return &WhenLoadingX402ConfigTestingSuite{
		sut:                  services.NewX402PaymentMethodStrategy(facilitators, &runtime_common_settings.ChainsSettings{}, mockErc20Validator, mockConfigRepository),
		mockConfigRepository: mockConfigRepository,
	}
}

func TestWhenLoadingX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule has a config", func(t *testing.T) {
		t.Parallel()

		row := model_builders.NewPaymentRuleX402ConfigBuilder().
			WithAddressPatterns([]string{".*"}).
			Build()

		initSuite := func(suite *WhenLoadingX402ConfigTestingSuite) {
			suite.mockConfigRepository.EXPECT().GetByRuleId(gomock.Any(), row.PaymentRuleId).Return(row, nil).Times(1)
		}

		t.Run("Should return it as the wire blob", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402ConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(context.Background(), row.PaymentRuleId)

			assert.NoError(t, err)
			config := &models.PaymentRuleX402Config{}
			assert.NoError(t, config.Parse(raw))
			assert.Equal(t, row.Network, config.Network)
			assert.Equal(t, []string{".*"}, config.AddressPatterns)
		})
	})

	t.Run("Given the rule has no config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenLoadingX402ConfigTestingSuite) {
			suite.mockConfigRepository.EXPECT().GetByRuleId(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402ConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(context.Background(), uuid.New())

			assert.NoError(t, err)
			assert.Nil(t, raw)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenLoadingX402ConfigTestingSuite) {
			suite.mockConfigRepository.EXPECT().GetByRuleId(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingX402ConfigBeforeEach(t)
			initSuite(suite)

			raw, err := suite.sut.LoadConfig(context.Background(), uuid.New())

			assert.Error(t, err)
			assert.Nil(t, raw)
		})
	})
}
