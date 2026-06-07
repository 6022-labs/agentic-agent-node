package configurations

import (
	"net/http"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/settings"
	"go.uber.org/dig"
)

func AddAgentIdentityHttpIpfsConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewIPFSSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(func() *http.Client {
		return http.DefaultClient
	}, dig.Name("ipfs-http-client")); err != nil {
		panic(err)
	}

	if err := container.Provide(
		clients.NewIPFSClient,
		dig.As(new(clients.IPFSClientInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewIPFSFileStorage,
		dig.As(new(identity_interfaces.IPFSFileStorageInterface)),
	); err != nil {
		panic(err)
	}
}
