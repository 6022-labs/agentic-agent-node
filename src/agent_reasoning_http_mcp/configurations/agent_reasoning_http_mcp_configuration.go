package configurations

import (
	"net"
	"net/http"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/services"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/dig"
)

func AddAgentReasoningHttpMcpConfiguration(container *dig.Container) {
	// Metrics
	err := container.Provide(
		metrics.NewMcpMetrics,
		dig.As(new(metrics_interfaces.McpMetricsInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Services
	err = container.Provide(
		services.NewMCPRequester,
		dig.As(new(interfaces.McpRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Clients
	err = container.Provide(
		clients.NewMcpClient,
		dig.As(new(clients.McpClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Http clients — transport wrapped so raw calls also emit http.client.* metrics.
	err = container.Provide(func() *http.Client {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DialContext = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext
		transport.TLSHandshakeTimeout = 10 * time.Second
		transport.IdleConnTimeout = 90 * time.Second
		return &http.Client{
			Transport: otelhttp.NewTransport(common_metrics.NewHttpMetricsRoundTripper(
				transport,
				common_metrics.NewExternalHttpMetrics("agentic_runtime_http_mcp"),
			),
				otelhttp.WithMeterProvider(metricnoop.NewMeterProvider()),
			)}
	}, dig.Name("mcp-http-client"))
	if err != nil {
		panic(err)
	}
}
