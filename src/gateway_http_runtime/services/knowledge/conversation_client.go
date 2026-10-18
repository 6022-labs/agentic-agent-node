package knowledge

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/samber/lo"
)

type HttpConversationClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpConversationClient(client clients.ClientWithResponsesInterface) *HttpConversationClient {
	return &HttpConversationClient{client: client}
}

func (c *HttpConversationClient) ListConversations(
	ctx context.Context,
	page uint,
	elementsPerPage uint,
) (*models.ExternalListConversationsPayload, error) {
	resp, err := c.client.GetInternalConversationsWithResponse(ctx, &clients.GetInternalConversationsParams{
		Page:            lo.ToPtr(int(page)),
		ElementsPerPage: lo.ToPtr(int(elementsPerPage)),
	})
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalListConversationsPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpConversationClient) CreateConversation(
	ctx context.Context,
	body *models.ExternalCreateConversationPayload,
) (*models.ExternalConversationPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConversationsWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalConversationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpConversationClient) UpdateConversation(
	ctx context.Context,
	conversationId string,
	body *models.ExternalUpdateConversationPayload,
) (*models.ExternalConversationPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConversationsConversationIdWithBodyWithResponse(
		ctx,
		conversationId,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalConversationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpConversationClient) ListMessages(
	ctx context.Context,
	conversationId string,
) (*models.ExternalListConversationMessagesPayload, error) {
	resp, err := c.client.GetInternalConversationsConversationIdMessagesWithResponse(ctx, conversationId)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalListConversationMessagesPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpConversationClient) AppendMessages(
	ctx context.Context,
	conversationId string,
	body *models.ExternalAppendConversationMessagesPayload,
) (*models.ExternalAppendConversationMessagesResultPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
		ctx,
		conversationId,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAppendConversationMessagesResultPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}
