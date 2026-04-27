package runtime_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/gofiber/fiber/v2"
)

const InternalBootstrapStatusEndpoint = "/internal/bootstrap/status"

type BootstrapController struct {
	stateReader common_interfaces.BootstrapStateReaderInterface
}

func NewBootstrapController(stateReader common_interfaces.BootstrapStateReaderInterface) *BootstrapController {
	return &BootstrapController{stateReader: stateReader}
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
	state := c.stateReader.Get()
	return ctx.JSON(&responses.GetBootstrapStatusResponse{
		Status:       state.Status,
		AgentAddress: state.AgentAddress,
		LastError:    state.LastError,
	})
}
