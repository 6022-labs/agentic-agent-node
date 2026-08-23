package middlewares

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/observability"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/gofiber/fiber/v3"
)

const agentIdentityResolveInterval = 30 * time.Second

// AgentTelemetryContextMiddleware puts the cached agent identity on the ctx so gateway signals carry it. Runs first.
type AgentTelemetryContextMiddleware struct {
	agentRegistrationClient identity_interfaces.AgentRegistrationClientInterface

	mu          sync.Mutex
	identity    *observability.AgentIdentity
	lastAttempt time.Time
}

func NewAgentTelemetryContextMiddleware(
	agentRegistrationClient identity_interfaces.AgentRegistrationClientInterface,
) *AgentTelemetryContextMiddleware {
	return &AgentTelemetryContextMiddleware{agentRegistrationClient: agentRegistrationClient}
}

func (m *AgentTelemetryContextMiddleware) Handle(ctx fiber.Ctx) error {
	var userCtx context.Context = ctx.RequestCtx()
	if identity, ok := m.resolveAgentIdentity(userCtx); ok {
		userCtx = observability.ContextWithAgentIdentity(userCtx, identity)
	}
	ctx.SetContext(userCtx)
	return ctx.Next()
}

// resolveAgentIdentity fetches the identity once, caching and throttling retries while unknown (no lock during I/O).
func (m *AgentTelemetryContextMiddleware) resolveAgentIdentity(
	reqCtx context.Context,
) (observability.AgentIdentity, bool) {
	m.mu.Lock()
	if m.identity != nil {
		identity := *m.identity
		m.mu.Unlock()
		return identity, true
	}
	if !m.lastAttempt.IsZero() && time.Since(m.lastAttempt) < agentIdentityResolveInterval {
		m.mu.Unlock()
		return observability.AgentIdentity{}, false
	}
	m.lastAttempt = time.Now()
	m.mu.Unlock()

	payload, err := m.agentRegistrationClient.Get(reqCtx)
	if err != nil || payload == nil || strings.TrimSpace(payload.EnsDomain) == "" {
		return observability.AgentIdentity{}, false
	}

	identity := observability.AgentIdentity{
		EnsDomain: payload.EnsDomain,
	}
	m.mu.Lock()
	m.identity = &identity
	m.mu.Unlock()
	return identity, true
}
