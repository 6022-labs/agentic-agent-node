package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenRoutingPaymentSpendMethodStrategiesTestingSuite struct {
	sut *services.PaymentSpendMethodStrategies

	mockStrategy *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
}

func WhenRoutingPaymentSpendMethodStrategiesBeforeEach(
	t *testing.T,
) *WhenRoutingPaymentSpendMethodStrategiesTestingSuite {
	mockController := gomock.NewController(t)

	mockStrategy := payment_interfaces_mocks.NewMockPaymentSpendMethodStrategyInterface(mockController)
	mockStrategy.EXPECT().Method().Return(enums.PaymentMethodX402).AnyTimes()

	return &WhenRoutingPaymentSpendMethodStrategiesTestingSuite{
		sut: services.NewPaymentSpendMethodStrategies(services.NewPaymentSpendMethodStrategiesParams{
			Strategies: []payment_interfaces.PaymentSpendMethodStrategyInterface{mockStrategy},
		}),
		mockStrategy: mockStrategy,
	}
}

func TestWhenRoutingPaymentSpendMethodStrategies(t *testing.T) {
	t.Parallel()

	t.Run("Given a method with a registered strategy", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return that strategy", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoutingPaymentSpendMethodStrategiesBeforeEach(t)

			strategy, ok := suite.sut.For(enums.PaymentMethodX402)

			assert.True(t, ok)
			assert.Equal(t, suite.mockStrategy, strategy)
		})
	})

	t.Run("Given a method with no registered strategy", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it unsupported", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoutingPaymentSpendMethodStrategiesBeforeEach(t)

			strategy, ok := suite.sut.For(enums.PaymentMethod("lightning"))

			assert.False(t, ok)
			assert.Nil(t, strategy)
		})
	})
}
