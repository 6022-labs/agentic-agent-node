package middlewares

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/observability"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

// AgentTelemetryContextMiddleware puts the agent identity on the ctx — the one place resolving "which agent" (swap for multi-agent). Runs first.
type AgentTelemetryContextMiddleware struct {
	agentRepository repositories.AgentRepositoryInterface
	cachedIdentity  atomic.Pointer[observability.AgentIdentity]
}

func NewAgentTelemetryContextMiddleware(
	agentRepository repositories.AgentRepositoryInterface,
) *AgentTelemetryContextMiddleware {
	return &AgentTelemetryContextMiddleware{
		agentRepository: agentRepository,
	}
}

func (m *AgentTelemetryContextMiddleware) Handle(ctx *fiber.Ctx) error {
	// Bridge the cancellable request ctx into UserContext (handlers are synchronous) + attach identity, so use cases get both.
	var userCtx context.Context = ctx.Context()
	if identity, ok := m.resolveAgentIdentity(userCtx); ok {
		userCtx = observability.ContextWithAgentIdentity(userCtx, identity)
	}
	ctx.SetUserContext(userCtx)
	return ctx.Next()
}

// resolveAgentIdentity caches the single agent's identity once known (lazy until minted).
func (m *AgentTelemetryContextMiddleware) resolveAgentIdentity(reqCtx context.Context) (observability.AgentIdentity, bool) {
	if cached := m.cachedIdentity.Load(); cached != nil {
		return *cached, true
	}
	agent, err := m.agentRepository.Get(reqCtx)
	if err != nil || agent == nil || strings.TrimSpace(lo.FromPtr(agent.EnsDomain)) == "" {
		return observability.AgentIdentity{}, false
	}
	identity := observability.AgentIdentity{
		EnsDomain: lo.FromPtr(agent.EnsDomain),
	}
	m.cachedIdentity.Store(&identity)
	return identity, true
}
