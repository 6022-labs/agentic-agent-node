package agent_knowledge_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/gofiber/fiber/v3"
)

const (
	InternalConversationsEndpoint    = "/internal/conversations"
	InternalConversationByIdEndpoint = "/internal/conversations/:conversationId"
)

type ConversationsController struct {
	listConversations  *use_cases.ListConversations
	createConversation *use_cases.CreateConversation
	updateConversation *use_cases.UpdateConversation
}

func NewConversationsController(
	listConversations *use_cases.ListConversations,
	createConversation *use_cases.CreateConversation,
	updateConversation *use_cases.UpdateConversation,
) *ConversationsController {
	return &ConversationsController{
		listConversations:  listConversations,
		createConversation: createConversation,
		updateConversation: updateConversation,
	}
}

func (c *ConversationsController) RegisterRoutes(app fiber.Router) {
	app.Get(InternalConversationsEndpoint, c.List)
	app.Post(InternalConversationsEndpoint, c.Create)
	app.Put(InternalConversationByIdEndpoint, c.Update)
}

// List returns one page of the agent's memory sessions, newest activity first.
// @Summary Internal list conversations
// @Tags Internal
// @Produce json
// @Param page query int false "Page (default 1)"
// @Param elementsPerPage query int false "Elements per page (default 10, max 100)"
// @Success 200 {object} responses.ListConversationsResponse
// @Failure 500 {object} fiber.Map "list_failed"
// @Router /internal/conversations [get]
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
// @Summary Internal create conversation
// @Tags Internal
// @Accept json
// @Produce json
// @Param request body requests.CreateConversationRequest true "Conversation name"
// @Success 201 {object} responses.ConversationResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 409 {object} fiber.Map "conversation_name_taken"
// @Failure 500 {object} fiber.Map "create_failed"
// @Router /internal/conversations [post]
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
// @Summary Internal update conversation
// @Tags Internal
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.UpdateConversationRequest true "Conversation name"
// @Success 200 {object} responses.ConversationResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 404 {object} fiber.Map "conversation_not_found"
// @Failure 409 {object} fiber.Map "conversation_name_taken"
// @Failure 500 {object} fiber.Map "update_failed"
// @Router /internal/conversations/{conversationId} [put]
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
