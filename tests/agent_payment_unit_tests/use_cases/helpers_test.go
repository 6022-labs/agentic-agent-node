package use_cases_test

import (
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"go.uber.org/mock/gomock"
)

// newMockStrategies returns a mock registry that resolves every method to a
// single mock strategy — the seam every payment use case dispatches config
// work through.
func newMockStrategies(mockController *gomock.Controller) (*payment_interfaces_mocks.MockPaymentMethodStrategiesInterface, *payment_interfaces_mocks.MockPaymentMethodStrategyInterface) {
	mockStrategy := payment_interfaces_mocks.NewMockPaymentMethodStrategyInterface(mockController)
	mockStrategies := payment_interfaces_mocks.NewMockPaymentMethodStrategiesInterface(mockController)
	mockStrategies.EXPECT().For(gomock.Any()).Return(mockStrategy, true).AnyTimes()
	return mockStrategies, mockStrategy
}
