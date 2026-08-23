package configurations

import (
	"net/http"
	"time"

	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_x402/clients"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/dig"
)

// facilitatorHttpTimeout must exceed an accept option's max_timeout_seconds:
// /settle blocks until the payment transaction lands on-chain.
const facilitatorHttpTimeout = 75 * time.Second

func AddGatewayHttpX402Configuration(container *dig.Container) {
	if err := container.Provide(
		func() *http.Client {
			return &http.Client{
				Timeout: facilitatorHttpTimeout,
				Transport: otelhttp.NewTransport(common_metrics.NewHttpMetricsRoundTripper(
					http.DefaultTransport,
					common_metrics.NewExternalHttpMetrics("agentic_gateway_http_x402"),
				),
					otelhttp.WithMeterProvider(metricnoop.NewMeterProvider()),
				),
			}
		},
		dig.Name("x402-facilitator-http-client"),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		clients.NewFacilitatorClient,
		dig.As(new(gateway_interfaces.X402FacilitatorClientInterface)),
	); err != nil {
		panic(err)
	}
}
