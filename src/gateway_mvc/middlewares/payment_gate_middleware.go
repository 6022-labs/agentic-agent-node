package middlewares

import (
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/middlewares/interfaces"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// defaultPaymentMethod advertises the 402 when a caller sends no payment proof.
const defaultPaymentMethod = methodX402

// PaymentGateMiddleware fetches the rules once and dispatches the request to
// the payment method that recognizes its proof, or to the default method's 402
// advertisement when no proof is present.
type PaymentGateMiddleware struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
	byMethod          map[string]middleware_interfaces.PaymentMethodInterface
}

type NewPaymentGateMiddlewareParams struct {
	dig.In

	PaymentRuleClient gateway_interfaces.PaymentRuleClientInterface
	PaymentMethods    []middleware_interfaces.PaymentMethodInterface `group:"payment_methods"`
}

func NewPaymentGateMiddleware(params NewPaymentGateMiddlewareParams) *PaymentGateMiddleware {
	byMethod := map[string]middleware_interfaces.PaymentMethodInterface{}
	for _, method := range params.PaymentMethods {
		byMethod[method.Method()] = method
	}
	return &PaymentGateMiddleware{
		paymentRuleClient: params.PaymentRuleClient,
		byMethod:          byMethod,
	}
}

func (m *PaymentGateMiddleware) Handle(ctx *fiber.Ctx) error {
	rules, err := m.paymentRuleClient.List(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_rules_unavailable", err.Error()))
	}

	for _, method := range m.byMethod {
		if method.HasProof(ctx) {
			return method.Handle(ctx, rules)
		}
	}

	// No payment proof: the default method advertises its 402. Aggregating
	// multiple methods into one 402 is a follow-up when a 2nd method exists.
	defaultMethod, ok := m.byMethod[defaultPaymentMethod]
	if !ok {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_methods_unavailable", "no payment method configured"))
	}
	return defaultMethod.Handle(ctx, rules)
}
