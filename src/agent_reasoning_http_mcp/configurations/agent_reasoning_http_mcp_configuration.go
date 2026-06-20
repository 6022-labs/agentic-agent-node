package configurations

import (
	"net/http"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/services"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
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
		return &http.Client{
			Transport: common_metrics.NewHttpMetricsRoundTripper(
				http.DefaultTransport,
				common_metrics.NewExternalHttpMetrics("agentic_runtime_http_mcp"),
			)}
	}, dig.Name("mcp-http-client"))
	if err != nil {
		panic(err)
	}
}
