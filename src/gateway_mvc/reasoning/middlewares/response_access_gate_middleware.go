package middlewares

import (
	"errors"
	"slices"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// ResponseAccessGateMiddleware runs the /responses access chain in priority order.
type ResponseAccessGateMiddleware struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
	chain             []middleware_interfaces.AccessMethodInterface
}

type NewResponseAccessGateMiddlewareParams struct {
	dig.In

	PaymentRuleClient gateway_interfaces.PaymentRuleClientInterface
	AccessMethods     []middleware_interfaces.AccessMethodInterface `group:"access_methods"`
}

func NewResponseAccessGateMiddleware(params NewResponseAccessGateMiddlewareParams) *ResponseAccessGateMiddleware {
	chain := slices.Clone(params.AccessMethods)

	slices.SortStableFunc(chain, func(a, b middleware_interfaces.AccessMethodInterface) int {
		return a.Priority() - b.Priority()
	})

	return &ResponseAccessGateMiddleware{
		paymentRuleClient: params.PaymentRuleClient,
		chain:             chain,
	}
}

func (m *ResponseAccessGateMiddleware) Handle(ctx *fiber.Ctx) error {
	rules, err := m.paymentRuleClient.List(ctx.UserContext())
	if err != nil {
		// Pre-mint the runtime refuses to materialize rules; run the chain ruleless (x402 advertises a 402).
		var conflictErr *apperrors.ConflictError
		if !errors.As(err, &conflictErr) {
			return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_rules_unavailable", err.Error()))
		}
		rules = nil
	}

	for _, link := range m.chain {
		if handled, err := link.TryHandle(ctx, rules); handled {
			return err
		}
	}

	// Unreachable while a terminal link (x402) is configured.
	return mvc.WriteError(
		ctx,
		apperrors.NewUnavailableError("access_methods_unavailable", "no access method handled the request"),
	)
}
