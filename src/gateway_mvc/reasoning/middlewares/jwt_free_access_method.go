package middlewares

import (
	"strings"

	core_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	core_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// jwtFreeAccessPriority runs this link before the payment methods.
const jwtFreeAccessPriority = 10

// JwtFreeAccessMethod grants a JWT-proven caller access when any matcher deems
// it free-entitled; anything else defers to the next link.
type JwtFreeAccessMethod struct {
	jwtValidator     core_interfaces.JwtValidatorInterface
	freeRuleMatchers []middleware_interfaces.FreeRuleMatcherInterface
}

type NewJwtFreeAccessMethodParams struct {
	dig.In

	JwtValidator     core_interfaces.JwtValidatorInterface
	FreeRuleMatchers []middleware_interfaces.FreeRuleMatcherInterface `group:"free_rule_matchers"`
}

func NewJwtFreeAccessMethod(params NewJwtFreeAccessMethodParams) *JwtFreeAccessMethod {
	return &JwtFreeAccessMethod{
		jwtValidator:     params.JwtValidator,
		freeRuleMatchers: params.FreeRuleMatchers,
	}
}

func (m *JwtFreeAccessMethod) Priority() int {
	return jwtFreeAccessPriority
}

func (m *JwtFreeAccessMethod) TryHandle(ctx *fiber.Ctx, rules []models.ExternalPaymentRule) (bool, error) {
	address, ok := m.freeAccessAddress(ctx, rules)
	if !ok {
		return false, nil
	}
	ctx.Locals(core_middlewares.AuthAddressLocalKey, address.Hex())
	return true, ctx.Next()
}

// freeAccessAddress returns the bearer's free-entitled address; false on any failure.
func (m *JwtFreeAccessMethod) freeAccessAddress(
	ctx *fiber.Ctx,
	rules []models.ExternalPaymentRule,
) (common.Address, bool) {
	const prefix = "Bearer "
	header := ctx.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(header, prefix) {
		return common.Address{}, false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return common.Address{}, false
	}

	address, err := m.jwtValidator.Verify(token)
	if err != nil {
		return common.Address{}, false
	}

	for _, matcher := range m.freeRuleMatchers {
		if matcher.MatchesFreeRule(rules, address) {
			return address, true
		}
	}
	return common.Address{}, false
}
