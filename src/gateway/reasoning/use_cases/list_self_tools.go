package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type ListSelfTools struct {
	selfToolConfigClient gateway_interfaces.SelfToolConfigClientInterface
}

func NewListSelfTools(selfToolConfigClient gateway_interfaces.SelfToolConfigClientInterface) *ListSelfTools {
	return &ListSelfTools{selfToolConfigClient: selfToolConfigClient}
}

func (u *ListSelfTools) Execute(ctx context.Context) (*responses.ListSelfToolsResponse, error) {
	payload, err := u.selfToolConfigClient.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]responses.SelfToolResponse, 0, len(payload.Tools))
	for i := range payload.Tools {
		out = append(out, *responses.ToSelfToolResponse(&payload.Tools[i]))
	}
	return &responses.ListSelfToolsResponse{Tools: out}, nil
}
