package gateway_mvc

import (
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	"github.com/gofiber/fiber/v2"
)

const BootstrapStatusEndpoint = "/bootstrap/status"

type BootstrapController struct {
	stateReader      common_interfaces.BootstrapStateReaderInterface
	identityResolver common_interfaces.AgentIdentityResolverInterface
}

func NewBootstrapController(
	stateReader common_interfaces.BootstrapStateReaderInterface,
	identityResolver common_interfaces.AgentIdentityResolverInterface,
) *BootstrapController {
	return &BootstrapController{
		stateReader:      stateReader,
		identityResolver: identityResolver,
	}
}

func (c *BootstrapController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(BootstrapStatusEndpoint, c.GetStatus)
}

func (c *BootstrapController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

func (c *BootstrapController) GetStatus(ctx *fiber.Ctx) error {
	state := c.stateReader.Get()

	resp := &responses.GetBootstrapStatusResponse{
		Status:       state.Status,
		AgentAddress: state.AgentAddress,
		LastError:    state.LastError,
	}

	// Best-effort identity resolution. Only attempted once the wallet is loaded;
	// failures are surfaced inline so the operator can see what's missing.
	if state.Status == enums.BootstrapStatusReady {
		identity, err := c.identityResolver.Resolve(ctx.Context())
		if err != nil {
			resp.Identity = &responses.AgentIdentitySnapshot{
				Resolved: false,
				Error:    err.Error(),
			}
		} else {
			resp.Identity = &responses.AgentIdentitySnapshot{
				Resolved:          true,
				CollectionAddress: identity.CollectionAddress,
				TokenId:           identity.TokenId,
				Name:              identity.Name,
			}
		}
	}

	return ctx.JSON(resp)
}

