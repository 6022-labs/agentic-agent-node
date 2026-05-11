package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

const bootstrapStatusPath = "/internal/bootstrap/status"

// HttpBootstrapStateReader implements common.BootstrapStateReaderInterface
// against runtime_host. The interface is sync; gateway pays a localhost RTT
// per Get() call.
type HttpBootstrapStateReader struct {
	client *clients.RuntimeHttpClient
}

func NewHttpBootstrapStateReader(client *clients.RuntimeHttpClient) *HttpBootstrapStateReader {
	return &HttpBootstrapStateReader{client: client}
}

func (r *HttpBootstrapStateReader) Get() models.BootstrapState {
	var payload runtime_models.ExternalBootstrapStatusPayload
	if err := r.client.Get(context.Background(), bootstrapStatusPath, &payload, nil); err != nil {
		return models.BootstrapState{LastError: err.Error()}
	}
	return models.BootstrapState{
		Status:       payload.Status,
		AgentAddress: payload.AgentAddress,
		LastError:    payload.LastError,
	}
}
