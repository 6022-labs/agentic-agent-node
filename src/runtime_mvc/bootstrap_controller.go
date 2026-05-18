package runtime_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/gofiber/fiber/v2"
)

const InternalBootstrapStatusEndpoint = "/internal/bootstrap/status"

type BootstrapController struct {
	getStatus *use_cases.GetBootstrapStatus
}

func NewBootstrapController(getStatus *use_cases.GetBootstrapStatus) *BootstrapController {
	return &BootstrapController{getStatus: getStatus}
}

func (c *BootstrapController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalBootstrapStatusEndpoint, c.GetStatus)
}

// GetStatus returns the current wallet bootstrap state.
// @Summary Internal bootstrap status
// @Description Wallet lifecycle snapshot for the gateway. Internal-only endpoint.
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.GetBootstrapStatusResponse
// @Router /internal/bootstrap/status [get]
func (c *BootstrapController) GetStatus(ctx *fiber.Ctx) error {
	return mvc_result.Write(ctx, c.getStatus.Execute())
}
