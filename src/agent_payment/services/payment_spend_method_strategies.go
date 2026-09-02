package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"go.uber.org/dig"
)

// PaymentSpendMethodStrategies routes a spend rule to its method strategy. Adding a
// method means registering its strategy in the dig group, nothing here.
type PaymentSpendMethodStrategies struct {
	byMethod map[enums.PaymentMethod]payment_interfaces.PaymentSpendMethodStrategyInterface
}

type NewPaymentSpendMethodStrategiesParams struct {
	dig.In

	Strategies []payment_interfaces.PaymentSpendMethodStrategyInterface `group:"payment_spend_method_strategies"`
}

func NewPaymentSpendMethodStrategies(params NewPaymentSpendMethodStrategiesParams) *PaymentSpendMethodStrategies {
	byMethod := map[enums.PaymentMethod]payment_interfaces.PaymentSpendMethodStrategyInterface{}
	for _, strategy := range params.Strategies {
		byMethod[strategy.Method()] = strategy
	}
	return &PaymentSpendMethodStrategies{byMethod: byMethod}
}

func (r *PaymentSpendMethodStrategies) For(
	method enums.PaymentMethod,
) (payment_interfaces.PaymentSpendMethodStrategyInterface, bool) {
	strategy, ok := r.byMethod[method]
	return strategy, ok
}
