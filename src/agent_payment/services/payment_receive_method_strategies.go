package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"go.uber.org/dig"
)

// PaymentReceiveMethodStrategies routes a rule to its method's strategy. Adding a
// method means registering its strategy in the dig group, nothing here.
type PaymentReceiveMethodStrategies struct {
	byMethod map[enums.PaymentMethod]payment_interfaces.PaymentReceiveMethodStrategyInterface
}

type NewPaymentReceiveMethodStrategiesParams struct {
	dig.In

	Strategies []payment_interfaces.PaymentReceiveMethodStrategyInterface `group:"payment_receive_method_strategies"`
}

func NewPaymentReceiveMethodStrategies(params NewPaymentReceiveMethodStrategiesParams) *PaymentReceiveMethodStrategies {
	byMethod := map[enums.PaymentMethod]payment_interfaces.PaymentReceiveMethodStrategyInterface{}
	for _, strategy := range params.Strategies {
		byMethod[strategy.Method()] = strategy
	}
	return &PaymentReceiveMethodStrategies{byMethod: byMethod}
}

func (r *PaymentReceiveMethodStrategies) For(
	method enums.PaymentMethod,
) (payment_interfaces.PaymentReceiveMethodStrategyInterface, bool) {
	strategy, ok := r.byMethod[method]
	return strategy, ok
}
