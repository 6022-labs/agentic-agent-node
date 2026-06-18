package identity

import (
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	"github.com/gofiber/fiber/v2"
)

const (
	AgentRegistrationEndpoint          = "/agent/registration"
	AgentRegistrationMintInfoEndpoint  = "/agent/registration/mint-info"
	AgentRegistrationMintCheckEndpoint = "/agent/registration/mint/check"
	AgentRegistrationEnsInfoEndpoint   = "/agent/registration/ens-info"
	AgentRegistrationEnsCheckEndpoint  = "/agent/registration/ens/check"
	AgentRegistrationSelfEndpoint      = "/agent/registration/self"
)

// AgentRegistrationController is the owner-facing mint + ENS registration API.
// The connected wallet (= owner) is taken from the JWT-authenticated session.
type AgentRegistrationController struct {
	saveAgentRegistration *use_cases.SaveAgentRegistration
	getAgentRegistration  *use_cases.GetAgentRegistration
	getAgentMintInfo      *use_cases.GetAgentMintInfo
	checkAgentMint        *use_cases.CheckAgentMint
	getAgentEnsInfo       *use_cases.GetAgentEnsInfo
	checkAgentEns         *use_cases.CheckAgentEns
	selfRegisterAgent     *use_cases.SelfRegisterAgent
}

func NewAgentRegistrationController(
	saveAgentRegistration *use_cases.SaveAgentRegistration,
	getAgentRegistration *use_cases.GetAgentRegistration,
	getAgentMintInfo *use_cases.GetAgentMintInfo,
	checkAgentMint *use_cases.CheckAgentMint,
	getAgentEnsInfo *use_cases.GetAgentEnsInfo,
	checkAgentEns *use_cases.CheckAgentEns,
	selfRegisterAgent *use_cases.SelfRegisterAgent,
) *AgentRegistrationController {
	return &AgentRegistrationController{
		saveAgentRegistration: saveAgentRegistration,
		getAgentRegistration:  getAgentRegistration,
		getAgentMintInfo:      getAgentMintInfo,
		checkAgentMint:        checkAgentMint,
		getAgentEnsInfo:       getAgentEnsInfo,
		checkAgentEns:         checkAgentEns,
		selfRegisterAgent:     selfRegisterAgent,
	}
}

func (c *AgentRegistrationController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *AgentRegistrationController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Post(AgentRegistrationEndpoint, c.Save)
	app.Get(AgentRegistrationEndpoint, c.Get)
	app.Get(AgentRegistrationMintInfoEndpoint, c.MintInfo)
	app.Post(AgentRegistrationMintCheckEndpoint, c.CheckMint)
	app.Get(AgentRegistrationEnsInfoEndpoint, c.EnsInfo)
	app.Post(AgentRegistrationEnsCheckEndpoint, c.CheckEns)
	app.Post(AgentRegistrationSelfEndpoint, c.Self)
}

// Save persists the agent draft filled by the connected wallet.
// @Summary Save agent registration draft
// @Tags AgentRegistration
// @Accept multipart/form-data
// @Produce json
// @Param collectionAddress formData string true "Collection address"
// @Param name formData string true "Agent name"
// @Param role formData string true "Agent role"
// @Param cloneOf formData int false "Token id this agent clones"
// @Param attributes formData string false "Attributes as a JSON object string"
// @Param image formData file true "Base image to frame"
// @Success 200 {object} responses.AgentRegistrationResponse
// @Failure 400 {object} errors.ValidationError
// @Security BearerAuth
// @Router /agent/registration [post]
func (c *AgentRegistrationController) Save(ctx *fiber.Ctx) error {
	body := new(requests.SaveAgentRegistrationRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if image, err := ctx.FormFile("image"); err == nil {
		body.Image = image
	}
	owner, ok := ctx.Locals(middlewares.AuthAddressLocalKey).(string)
	if !ok || owner == "" {
		return mvc.WriteError(ctx, errors.NewUnauthorizedError("missing_owner", "authenticated address is required"))
	}
	resp, err := c.saveAgentRegistration.Execute(ctx.UserContext(), body, owner)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Get returns the stored draft to pre-fill the form on retry.
// @Summary Get agent registration draft
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.AgentRegistrationResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Security BearerAuth
// @Router /agent/registration [get]
func (c *AgentRegistrationController) Get(ctx *fiber.Ctx) error {
	resp, err := c.getAgentRegistration.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// MintInfo returns the mint() calldata + moderator state.
// @Summary Get mint calldata
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.MintInfoResponse
// @Security BearerAuth
// @Router /agent/registration/mint-info [get]
func (c *AgentRegistrationController) MintInfo(ctx *fiber.Ctx) error {
	resp, err := c.getAgentMintInfo.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// CheckMint reconciles the mint status from chain after a browser-wallet tx.
// @Summary Check mint status
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Security BearerAuth
// @Router /agent/registration/mint/check [post]
func (c *AgentRegistrationController) CheckMint(ctx *fiber.Ctx) error {
	resp, err := c.checkAgentMint.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// EnsInfo returns the initAgentTexts calldata.
// @Summary Get ENS calldata
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.EnsInfoResponse
// @Failure 409 {object} fiber.Map "not_minted"
// @Security BearerAuth
// @Router /agent/registration/ens-info [get]
func (c *AgentRegistrationController) EnsInfo(ctx *fiber.Ctx) error {
	resp, err := c.getAgentEnsInfo.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// CheckEns reconciles ENS provisioning from chain after a browser-wallet tx.
// @Summary Check ENS status
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Security BearerAuth
// @Router /agent/registration/ens/check [post]
func (c *AgentRegistrationController) CheckEns(ctx *fiber.Ctx) error {
	resp, err := c.checkAgentEns.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Self mints + ENS-registers with the node wallet after a gas donation.
// @Summary Self mint + ENS register
// @Tags AgentRegistration
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Security BearerAuth
// @Router /agent/registration/self [post]
func (c *AgentRegistrationController) Self(ctx *fiber.Ctx) error {
	resp, err := c.selfRegisterAgent.Execute(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
