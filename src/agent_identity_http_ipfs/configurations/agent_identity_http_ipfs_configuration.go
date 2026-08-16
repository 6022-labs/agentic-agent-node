package configurations

import (
	"net"
	"net/http"
	"time"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_http_ipfs/settings"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/dig"
)

func AddAgentIdentityHttpIpfsConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewIPFSSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(func() *http.Client {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext
		transport.TLSHandshakeTimeout = 10 * time.Second
		transport.IdleConnTimeout = 90 * time.Second
		return &http.Client{
			Transport: otelhttp.NewTransport(common_metrics.NewHttpMetricsRoundTripper(
				transport,
				common_metrics.NewExternalHttpMetrics("agentic_runtime_http_ipfs"),
			))}
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
