package gateway_mvc

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	"github.com/gofiber/fiber/v2"
)

const HealthEndpoint = "/health"

type HealthController struct {
	startedAt time.Time
}

func NewHealthController() *HealthController {
	return &HealthController{startedAt: time.Now()}
}

func (c *HealthController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(HealthEndpoint, c.Health)
}

func (c *HealthController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Health returns liveness + uptime.
// @Summary Health
// @Description Liveness probe; serves before bootstrap completes.
// @Tags Health
// @Produce json
// @Success 200 {object} responses.GetHealthResponse
// @Router /health [get]
func (c *HealthController) Health(ctx *fiber.Ctx) error {
	return ctx.JSON(&responses.GetHealthResponse{
		Status:        "ok",
		UptimeSeconds: int64(time.Since(c.startedAt).Seconds()),
	})
}
