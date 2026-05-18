package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc_result"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/gofiber/fiber/v2"
)

const BootstrapStatusEndpoint = "/bootstrap/status"

type BootstrapController struct {
	getStatus *use_cases.GetBootstrapStatus
}

func NewBootstrapController(getStatus *use_cases.GetBootstrapStatus) *BootstrapController {
	return &BootstrapController{getStatus: getStatus}
}

func (c *BootstrapController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(BootstrapStatusEndpoint, c.GetStatus)
}

func (c *BootstrapController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// GetStatus reports wallet bootstrap state and a best-effort identity snapshot.
// @Summary Bootstrap status
// @Description Wallet lifecycle (`awaiting-vault` | `ready`) plus a best-effort NFT identity snapshot once the wallet is loaded.
// @Tags Bootstrap
// @Produce json
// @Success 200 {object} responses.GetBootstrapStatusResponse
// @Router /bootstrap/status [get]
func (c *BootstrapController) GetStatus(ctx *fiber.Ctx) error {
	return mvc_result.Write(ctx, c.getStatus.Execute(ctx.Context()))
}
