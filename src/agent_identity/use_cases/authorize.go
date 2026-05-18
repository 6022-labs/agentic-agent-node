package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/ethereum/go-ethereum/common"
)

type Authorize struct {
	authorizer common_interfaces.AuthorizerInterface
}

func NewAuthorize(authorizer common_interfaces.AuthorizerInterface) *Authorize {
	return &Authorize{authorizer: authorizer}
}

func (u *Authorize) Execute(ctx context.Context, req *requests.AuthorizeRequest) *usecase_result.Result {
	authorized, err := u.authorizer.IsAuthorized(ctx, common.HexToAddress(req.Address))
	if err != nil {
		return usecase_result.InternalError("authorize_failed", err.Error())
	}
	return usecase_result.Ok(&responses.AuthorizeResponse{Authorized: authorized})
}
