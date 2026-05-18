package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
)

type GetBootstrapStatus struct {
	stateReader common_interfaces.BootstrapStateReaderInterface
}

func NewGetBootstrapStatus(stateReader common_interfaces.BootstrapStateReaderInterface) *GetBootstrapStatus {
	return &GetBootstrapStatus{stateReader: stateReader}
}

func (u *GetBootstrapStatus) Execute() *usecase_result.Result {
	state := u.stateReader.Get()
	return usecase_result.Ok(&responses.GetBootstrapStatusResponse{
		Status:       state.Status,
		AgentAddress: state.AgentAddress,
		LastError:    state.LastError,
	})
}
