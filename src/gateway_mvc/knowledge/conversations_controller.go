package knowledge

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const (
	ConversationsEndpoint        = "/conversations"
	ConversationByIdEndpoint     = "/conversations/:conversationId"
	ConversationMessagesEndpoint = "/conversations/:conversationId/messages"
)

type ConversationsController struct {
	listConversations          *use_cases.ListConversations
	createConversation         *use_cases.CreateConversation
	updateConversation         *use_cases.UpdateConversation
	listConversationMessages   *use_cases.ListConversationMessages
	appendConversationMessages *use_cases.AppendConversationMessages
}

func NewConversationsController(
	listConversations *use_cases.ListConversations,
	createConversation *use_cases.CreateConversation,
	updateConversation *use_cases.UpdateConversation,
	listConversationMessages *use_cases.ListConversationMessages,
	appendConversationMessages *use_cases.AppendConversationMessages,
) *ConversationsController {
	return &ConversationsController{
		listConversations:          listConversations,
		createConversation:         createConversation,
		updateConversation:         updateConversation,
		listConversationMessages:   listConversationMessages,
		appendConversationMessages: appendConversationMessages,
	}
}

func (c *ConversationsController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *ConversationsController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(ConversationsEndpoint, c.List)
	app.Post(ConversationsEndpoint, c.Create)
	app.Put(ConversationByIdEndpoint, c.Update)
	app.Get(ConversationMessagesEndpoint, c.ListMessages)
	app.Post(ConversationMessagesEndpoint, c.AppendMessages)
}

// List returns one page of the agent's memory sessions (most recent activity first).
// @Summary List conversations
// @Tags Conversations
// @Produce json
// @Param page query int false "Page (default 1)"
// @Param elementsPerPage query int false "Elements per page (default 10, max 100)"
// @Success 200 {object} responses.ListConversationsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations [get]
func (c *ConversationsController) List(ctx fiber.Ctx) error {
	req := &requests.ListConversationsRequest{}
	if err := ctx.Bind().Query(req); err != nil {
		return mvc.WriteError(ctx, err)
	}

	resp, err := c.listConversations.Execute(ctx.Context(), req)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// Create opens an empty named memory session.
// @Summary Create conversation
// @Tags Conversations
// @Accept json
// @Produce json
// @Param request body requests.CreateConversationRequest true "Conversation name"
// @Success 201 {object} responses.ConversationResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 409 {object} errors.ConflictError "conversation_name_taken or agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations [post]
func (c *ConversationsController) Create(ctx fiber.Ctx) error {
	body := new(requests.CreateConversationRequest)
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}

	resp, err := c.createConversation.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}

// Update relabels a memory session.
// @Summary Update conversation
// @Tags Conversations
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.UpdateConversationRequest true "Conversation name"
// @Success 200 {object} responses.ConversationResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "conversation_not_found"
// @Failure 409 {object} errors.ConflictError "conversation_name_taken"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId} [put]
func (c *ConversationsController) Update(ctx fiber.Ctx) error {
	request := new(requests.UpdateConversationRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}

	resp, err := c.updateConversation.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// ListMessages returns the transcript for one conversation (timestamp ascending).
// @Summary List conversation messages
// @Tags Conversations
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Success 200 {object} responses.ListConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "conversation_not_found"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId}/messages [get]
func (c *ConversationsController) ListMessages(ctx fiber.Ctx) error {
	request := new(requests.ListConversationMessagesRequest)
	if err := ctx.Bind().URI(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}

	resp, err := c.listConversationMessages.Execute(ctx.Context(), request)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.JSON(resp)
}

// AppendMessages injects messages into a conversation transcript (out-of-band,
// e.g. seeding "memory"). Idempotent on the runtime via content hashing.
// @Summary Append conversation messages
// @Tags Conversations
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.AppendConversationMessagesRequest true "Messages"
// @Success 201 {object} responses.AppendConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError "missing_authorization or invalid_token"
// @Failure 404 {object} errors.NotFoundError "conversation_not_found"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 502 {object} errors.UnavailableError "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId}/messages [post]
func (c *ConversationsController) AppendMessages(ctx fiber.Ctx) error {
	body := new(requests.AppendConversationMessagesRequest)
	if err := ctx.Bind().URI(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_params", "message": err.Error()})
	}
	if err := ctx.Bind().Body(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}

	resp, err := c.appendConversationMessages.Execute(ctx.Context(), body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}
