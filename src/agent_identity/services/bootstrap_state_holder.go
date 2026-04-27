package services

import (
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
)

type BootstrapStateHolder struct {
	mu    sync.RWMutex
	state models.BootstrapState
}

func NewBootstrapStateHolder() *BootstrapStateHolder {
	return &BootstrapStateHolder{
		state: models.BootstrapState{
			Status: enums.BootstrapStatusAwaitingVault,
		},
	}
}

func (h *BootstrapStateHolder) Get() models.BootstrapState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

func (h *BootstrapStateHolder) SetStatus(status enums.BootstrapStatus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state.Status = status
}

func (h *BootstrapStateHolder) SetError(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state.LastError = message
}

func (h *BootstrapStateHolder) SetReady(address string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state.Status = enums.BootstrapStatusReady
	h.state.AgentAddress = address
	h.state.LastError = ""
}
