package knowledge

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpConversationHistoryClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpConversationHistoryClient(client clients.ClientWithResponsesInterface) *HttpConversationHistoryClient {
	return &HttpConversationHistoryClient{client: client}
}

func (c *HttpConversationHistoryClient) ListConversations(ctx context.Context) (*models.ExternalListConversationsPayload, error) {
	resp, err := c.client.GetInternalConversationsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalListConversationsPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpConversationHistoryClient) ListMessages(ctx context.Context, conversationId string) (*models.ExternalListConversationMessagesPayload, error) {
	resp, err := c.client.GetInternalConversationsConversationIdMessagesWithResponse(ctx, conversationId)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalListConversationMessagesPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpConversationHistoryClient) AppendMessages(ctx context.Context, conversationId string, body *models.ExternalAppendConversationMessagesPayload) (*models.ExternalAppendConversationMessagesResultPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConversationsConversationIdMessagesWithBodyWithResponse(ctx, conversationId, helpers.JsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	var payload models.ExternalAppendConversationMessagesResultPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}
