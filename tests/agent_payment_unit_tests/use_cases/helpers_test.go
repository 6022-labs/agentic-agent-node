package use_cases_test

import (
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"go.uber.org/mock/gomock"
)

// newMockStrategies returns a registry resolving every method to one mock strategy.
func newMockStrategies(
	mockController *gomock.Controller,
) (*payment_interfaces_mocks.MockPaymentReceiveMethodStrategiesInterface, *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface) {
	mockStrategy := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategyInterface(mockController)
	mockStrategies := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategiesInterface(mockController)
	mockStrategies.EXPECT().For(gomock.Any()).Return(mockStrategy, true).AnyTimes()
	return mockStrategies, mockStrategy
}

func newUnsupportedMethodStrategies(
	mockController *gomock.Controller,
) *payment_interfaces_mocks.MockPaymentReceiveMethodStrategiesInterface {
	mockStrategies := payment_interfaces_mocks.NewMockPaymentReceiveMethodStrategiesInterface(mockController)
	mockStrategies.EXPECT().For(gomock.Any()).Return(nil, false).Times(1)
	return mockStrategies
}
