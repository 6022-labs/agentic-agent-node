package configurations

import (
	"net/http"
	"time"

	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/transports"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/dig"
)

func AddAgentReasoningHttpA2AConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewA2ASettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewA2AMetrics,
		dig.As(new(metrics_interfaces.A2AMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewA2ARequester,
		dig.As(new(reasoning_interfaces.A2ARequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		clients.NewA2AClient,
		dig.As(new(clients.A2AClientInterface)),
	); err != nil {
		panic(err)
	}

	// Http client — transport wrapped so raw calls also emit http.client.* metrics.
	if err := container.Provide(func(a2aSettings *settings.A2ASettings) *http.Client {
		return &http.Client{
			Timeout: time.Duration(a2aSettings.RequestTimeoutSeconds) * time.Second,
			Transport: otelhttp.NewTransport(common_metrics.NewHttpMetricsRoundTripper(
				transports.NewA2ARoundTripper(a2aSettings),
				common_metrics.NewExternalHttpMetrics("agentic_runtime_http_a2a"),
			),
				otelhttp.WithMeterProvider(metricnoop.NewMeterProvider()),
			),
			// PAYMENT-SIGNATURE must never follow a redirect to another origin.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
	}, dig.Name("a2a-http-client")); err != nil {
		panic(err)
	}
}
