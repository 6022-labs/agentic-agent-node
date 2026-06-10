package gateway_mvc

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/gofiber/fiber/v2"
)

const (
	ConversationsEndpoint        = "/conversations"
	ConversationMessagesEndpoint = "/conversations/:conversationId/messages"
)

type ConversationsController struct {
	listConversations          *use_cases.ListConversations
	listConversationMessages   *use_cases.ListConversationMessages
	appendConversationMessages *use_cases.AppendConversationMessages
}

func NewConversationsController(
	listConversations *use_cases.ListConversations,
	listConversationMessages *use_cases.ListConversationMessages,
	appendConversationMessages *use_cases.AppendConversationMessages,
) *ConversationsController {
	return &ConversationsController{
		listConversations:          listConversations,
		listConversationMessages:   listConversationMessages,
		appendConversationMessages: appendConversationMessages,
	}
}

func (c *ConversationsController) RegisterUnauthenticatedRoutes(_ fiber.Router) {}

func (c *ConversationsController) RegisterAuthenticatedRoutes(app fiber.Router) {
	app.Get(ConversationsEndpoint, c.List)
	app.Get(ConversationMessagesEndpoint, c.ListMessages)
	app.Post(ConversationMessagesEndpoint, c.AppendMessages)
}

// List returns the agent's past conversations (most recent first).
// @Summary List conversations
// @Tags Conversations
// @Produce json
// @Success 200 {object} responses.ListConversationsResponse
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations [get]
func (c *ConversationsController) List(ctx *fiber.Ctx) error {
	resp, err := c.listConversations.Execute(ctx.UserContext())
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
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId}/messages [get]
func (c *ConversationsController) ListMessages(ctx *fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	resp, err := c.listConversationMessages.Execute(ctx.UserContext(), conversationId)
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
// @Failure 502 {object} fiber.Map "runtime_unavailable"
// @Security BearerAuth
// @Router /conversations/{conversationId}/messages [post]
func (c *ConversationsController) AppendMessages(ctx *fiber.Ctx) error {
	conversationId := strings.TrimSpace(ctx.Params("conversationId"))
	if len(conversationId) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(errors.NewValidationError("conversationId", "cannot be empty"))
	}

	body := new(requests.AppendConversationMessagesRequest)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": err.Error()})
	}

	resp, err := c.appendConversationMessages.Execute(ctx.UserContext(), conversationId, body)
	if err != nil {
		return mvc.WriteError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(resp)
}
