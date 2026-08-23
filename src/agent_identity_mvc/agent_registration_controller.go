package agent_identity_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const (
	InternalAgentRegistrationEndpoint          = "/internal/agent/registration"
	InternalAgentRegistrationMintInfoEndpoint  = "/internal/agent/registration/mint-info"
	InternalAgentRegistrationMintCheckEndpoint = "/internal/agent/registration/mint/check"
	InternalAgentRegistrationEnsInfoEndpoint   = "/internal/agent/registration/ens-info"
	InternalAgentRegistrationEnsCheckEndpoint  = "/internal/agent/registration/ens/check"
	InternalAgentRegistrationSelfEndpoint      = "/internal/agent/registration/self"
)

// AgentRegistrationController drives the mint + ENS registration lifecycle; the
// gateway proxies the owner's authenticated calls here.
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

func (c *AgentRegistrationController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalAgentRegistrationEndpoint, c.Save)
	app.Get(InternalAgentRegistrationEndpoint, c.Get)
	app.Get(InternalAgentRegistrationMintInfoEndpoint, c.MintInfo)
	app.Post(InternalAgentRegistrationMintCheckEndpoint, c.CheckMint)
	app.Get(InternalAgentRegistrationEnsInfoEndpoint, c.EnsInfo)
	app.Post(InternalAgentRegistrationEnsCheckEndpoint, c.CheckEns)
	app.Post(InternalAgentRegistrationSelfEndpoint, c.Self)
}

// Save upserts the agent draft filled by the connected wallet.
// @Summary Save agent registration draft
// @Description Validates the collection + form and binds the draft to the node wallet. Returns stored draft + status.
// @Tags Internal
// @Accept multipart/form-data
// @Produce json
// @Param collectionAddress formData string true "Collection address"
// @Param owner formData string true "Owner address"
// @Param name formData string true "Agent name"
// @Param cloneOf formData int false "Token id this agent clones"
// @Param attributes formData string true "Attributes as a JSON object string; must include the required attributes"
// @Param image formData file true "Base image to frame"
// @Success 200 {object} responses.AgentRegistrationResponse
// @Failure 400 {object} errors.ValidationError
// @Router /internal/agent/registration [post]
func (c *AgentRegistrationController) Save(ctx fiber.Ctx) error {
	body := new(requests.SaveAgentRegistrationRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}
	if image, err := ctx.FormFile("image"); err == nil {
		body.Image = image
	}
	resp, err := c.saveAgentRegistration.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Get returns the stored draft to pre-fill the form on retry.
// @Summary Get agent registration draft
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.AgentRegistrationResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Router /internal/agent/registration [get]
func (c *AgentRegistrationController) Get(ctx fiber.Ctx) error {
	resp, err := c.getAgentRegistration.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// MintInfo returns the mint() calldata + moderator state.
// @Summary Get mint calldata
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.MintInfoResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Router /internal/agent/registration/mint-info [get]
func (c *AgentRegistrationController) MintInfo(ctx fiber.Ctx) error {
	resp, err := c.getAgentMintInfo.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// CheckMint reconciles the mint status from chain.
// @Summary Check mint status
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Router /internal/agent/registration/mint/check [post]
func (c *AgentRegistrationController) CheckMint(ctx fiber.Ctx) error {
	resp, err := c.checkAgentMint.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// EnsInfo returns the initAgentTexts calldata.
// @Summary Get ENS calldata
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.EnsInfoResponse
// @Failure 409 {object} fiber.Map "not_minted"
// @Router /internal/agent/registration/ens-info [get]
func (c *AgentRegistrationController) EnsInfo(ctx fiber.Ctx) error {
	resp, err := c.getAgentEnsInfo.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// CheckEns reconciles ENS provisioning from chain.
// @Summary Check ENS status
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Failure 409 {object} fiber.Map "not_minted"
// @Router /internal/agent/registration/ens/check [post]
func (c *AgentRegistrationController) CheckEns(ctx fiber.Ctx) error {
	resp, err := c.checkAgentEns.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Self mints + ENS-registers with the node wallet (after a gas donation).
// @Summary Self mint + ENS register
// @Description Uses the node's bootstrap wallet to mint (or create a mint proposal) and ENS-register the stored draft.
// @Tags Internal
// @Produce json
// @Success 200 {object} responses.RegistrationStatusResponse
// @Failure 404 {object} fiber.Map "draft_not_found"
// @Router /internal/agent/registration/self [post]
func (c *AgentRegistrationController) Self(ctx fiber.Ctx) error {
	resp, err := c.selfRegisterAgent.Execute(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}
