package core

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	"github.com/gofiber/fiber/v3"
)

const ConfigEndpoint = "/config"

// ConfigController serves the public node config the owner UI boots from —
// unauthenticated because the UI needs the chain id before wallet login.
type ConfigController struct {
	getConfig *use_cases.GetConfig
}

func NewConfigController(
	getConfig *use_cases.GetConfig,
) *ConfigController {
	return &ConfigController{getConfig: getConfig}
}

func (c *ConfigController) RegisterUnauthenticatedRoutes(app fiber.Router) {
	app.Get(ConfigEndpoint, c.Get)
}

func (c *ConfigController) RegisterAuthenticatedRoutes(_ fiber.Router) {}

// Get returns the chain config the node runs against and the IPFS gateways to use.
// @Summary Public node config
// @Tags Config
// @Produce json
// @Success 200 {object} responses.ConfigResponse
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Router /config [get]
func (c *ConfigController) Get(ctx fiber.Ctx) error {
	resp, err := c.getConfig.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
