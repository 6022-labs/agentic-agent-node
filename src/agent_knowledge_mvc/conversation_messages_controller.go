package agent_knowledge_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const InternalConversationMessagesEndpoint = "/internal/conversations/:conversationId/messages"

type ConversationMessagesController struct {
	appendConversationMessages *use_cases.AppendConversationMessages
	listConversationMessages   *use_cases.ListConversationMessages
}

func NewConversationMessagesController(
	appendConversationMessages *use_cases.AppendConversationMessages,
	listConversationMessages *use_cases.ListConversationMessages,
) *ConversationMessagesController {
	return &ConversationMessagesController{
		appendConversationMessages: appendConversationMessages,
		listConversationMessages:   listConversationMessages,
	}
}

func (c *ConversationMessagesController) RegisterRoutes(app fiber.Router) {
	app.Post(InternalConversationMessagesEndpoint, c.Append)
	app.Get(InternalConversationMessagesEndpoint, c.List)
}

// Append adds messages to a conversation transcript.
// @Summary Internal append messages
// @Description Idempotent batch append to the transcript. Dedup by SHA-512 hash of (role, content, conversationId).
// @Tags Internal
// @Accept json
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Param request body requests.AppendConversationMessagesRequest true "Messages"
// @Success 200 {object} responses.AppendConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError
// @Failure 404 {object} errors.NotFoundError "conversation_not_found"
// @Failure 409 {object} errors.ConflictError "agent_not_minted"
// @Failure 500 {object} errors.InternalError "append_failed"
// @Security CallerId
// @Router /internal/conversations/{conversationId}/messages [post]
func (c *ConversationMessagesController) Append(ctx fiber.Ctx) error {
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
	return ctx.JSON(resp)
}

// List returns the conversation transcript ordered by timestamp ascending.
// @Summary Internal list messages
// @Tags Internal
// @Produce json
// @Param conversationId path string true "Conversation id"
// @Success 200 {object} responses.ListConversationMessagesResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 401 {object} errors.UnauthorizedError
// @Failure 404 {object} errors.NotFoundError "conversation_not_found"
// @Failure 500 {object} errors.InternalError "list_failed"
// @Security CallerId
// @Router /internal/conversations/{conversationId}/messages [get]
func (c *ConversationMessagesController) List(ctx fiber.Ctx) error {
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
