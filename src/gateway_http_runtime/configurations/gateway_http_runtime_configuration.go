package configurations

import (
	"net/http"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	core_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	core_services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/core"
	identity_services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/identity"
	knowledge_services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/knowledge"
	payment_services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	reasoning_services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/dig"
)

// newRuntimeClient builds the generated runtime client with a metrics-wrapped transport.
func newRuntimeClient(s *settings.RuntimeClientSettings) (clients.ClientWithResponsesInterface, error) {
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(common_metrics.NewHttpMetricsRoundTripper(
			http.DefaultTransport,
			common_metrics.NewExternalHttpMetrics("agentic_gateway_http_runtime"),
		),
			otelhttp.WithMeterProvider(metricnoop.NewMeterProvider()),
		),
	}
	return clients.NewClientWithResponses(s.BaseUrl, clients.WithHTTPClient(httpClient))
}

// AddGatewayHttpRuntimeConfiguration wires the shared runtime client plus each
// per-domain set of HTTP client adapters (core, identity, knowledge, payment, reasoning).
func AddGatewayHttpRuntimeConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewRuntimeClientSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(newRuntimeClient); err != nil {
		panic(err)
	}

	addCoreConfiguration(container)
	addIdentityConfiguration(container)
	addKnowledgeConfiguration(container)
	addPaymentConfiguration(container)
	addReasoningConfiguration(container)
}

func addCoreConfiguration(container *dig.Container) {
	if err := container.Provide(
		core_services.NewHttpAuthorizer,
		dig.As(new(common_interfaces.AuthorizerInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		core_services.NewHttpBlockchainConfigClient,
		dig.As(new(core_interfaces.BlockchainConfigClientInterface)),
	); err != nil {
		panic(err)
	}
}

func addIdentityConfiguration(container *dig.Container) {
	if err := container.Provide(
		identity_services.NewHttpAgentRegistrationClient,
		dig.As(new(identity_interfaces.AgentRegistrationClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		identity_services.NewHttpAgentCollectionsClient,
		dig.As(new(identity_interfaces.AgentCollectionsClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		identity_services.NewHttpAgentSoulClient,
		dig.As(new(identity_interfaces.AgentSoulClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		identity_services.NewHttpAgentCardSignatureClient,
		dig.As(new(identity_interfaces.AgentCardSignatureClientInterface)),
	); err != nil {
		panic(err)
	}
}

func addKnowledgeConfiguration(container *dig.Container) {
	if err := container.Provide(
		knowledge_services.NewHttpConversationClient,
		dig.As(new(knowledge_interfaces.ConversationClientInterface)),
	); err != nil {
		panic(err)
	}
}

func addPaymentConfiguration(container *dig.Container) {
	if err := container.Provide(
		payment_services.NewHttpPaymentPolicyClient,
		dig.As(new(payment_interfaces.PaymentPolicyClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		payment_services.NewHttpPaymentRuleClient,
		dig.As(new(payment_interfaces.PaymentRuleClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		payment_services.NewHttpPaymentFacilitatorClient,
		dig.As(new(payment_interfaces.PaymentFacilitatorClientInterface)),
	); err != nil {
		panic(err)
	}
}

func addReasoningConfiguration(container *dig.Container) {
	if err := container.Provide(
		reasoning_services.NewHttpLLMConfigClient,
		dig.As(new(reasoning_interfaces.LLMConfigClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		reasoning_services.NewHttpMCPConfigClient,
		dig.As(new(reasoning_interfaces.MCPConfigClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		reasoning_services.NewHttpSelfToolConfigClient,
		dig.As(new(reasoning_interfaces.SelfToolConfigClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		reasoning_services.NewHttpResponseClient,
		dig.As(new(reasoning_interfaces.ResponseClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		reasoning_services.NewHttpAgentGuidanceClient,
		dig.As(new(reasoning_interfaces.AgentGuidanceClientInterface)),
	); err != nil {
		panic(err)
	}
}
