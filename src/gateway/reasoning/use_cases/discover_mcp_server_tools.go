package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
	"github.com/samber/lo"
)

type DiscoverMcpServerTools struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewDiscoverMcpServerTools(client gateway_interfaces.MCPConfigClientInterface) *DiscoverMcpServerTools {
	return &DiscoverMcpServerTools{client: client}
}

func (u *DiscoverMcpServerTools) Execute(
	ctx context.Context,
	request *requests.DiscoverMcpServerToolsRequest,
) (*responses.McpServerToolsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	var connectionId *string
	if request.ConnectionId != nil {
		connectionId = lo.ToPtr(request.ConnectionId.String())
	}

	payload, err := u.client.DiscoverTools(ctx, &models.ExternalDiscoverMCPServerToolsPayload{
		Url:          request.Url,
		HeaderAuth:   request.HeaderAuth,
		ConnectionId: connectionId,
	})
	if err != nil {
		return nil, err
	}
	return responses.ToMcpServerToolsResponse(payload), nil
}
