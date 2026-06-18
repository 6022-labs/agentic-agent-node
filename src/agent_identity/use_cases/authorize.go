package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

type Authorize struct {
	authorizer common_interfaces.AuthorizerInterface
}

func NewAuthorize(authorizer common_interfaces.AuthorizerInterface) *Authorize {
	return &Authorize{authorizer: authorizer}
}

func (u *Authorize) Execute(ctx context.Context, req *requests.AuthorizeRequest) (*responses.AuthorizeResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	authorized, err := u.authorizer.IsAuthorized(ctx, common.HexToAddress(req.Address))
	if err != nil {
		return nil, errors.NewInternalError("authorize_failed", err.Error())
	}
	return &responses.AuthorizeResponse{Authorized: authorized}, nil
}
