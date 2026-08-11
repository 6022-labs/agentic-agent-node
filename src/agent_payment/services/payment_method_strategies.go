package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"go.uber.org/dig"
)

// PaymentMethodStrategies routes a rule to its method's strategy. Adding a
// method means registering its strategy in the dig group, nothing here.
type PaymentMethodStrategies struct {
	byMethod map[enums.PaymentMethod]payment_interfaces.PaymentMethodStrategyInterface
}

type NewPaymentMethodStrategiesParams struct {
	dig.In

	Strategies []payment_interfaces.PaymentMethodStrategyInterface `group:"payment_method_strategies"`
}

func NewPaymentMethodStrategies(params NewPaymentMethodStrategiesParams) *PaymentMethodStrategies {
	byMethod := map[enums.PaymentMethod]payment_interfaces.PaymentMethodStrategyInterface{}
	for _, strategy := range params.Strategies {
		byMethod[strategy.Method()] = strategy
	}
	return &PaymentMethodStrategies{byMethod: byMethod}
}

func (r *PaymentMethodStrategies) For(
	method enums.PaymentMethod,
) (payment_interfaces.PaymentMethodStrategyInterface, bool) {
	strategy, ok := r.byMethod[method]
	return strategy, ok
}
