package responses

// WellKnownAgent6022 is the 6022-protocol discovery document; payment terms come from the HTTP 402 handshake on /responses.
type WellKnownAgent6022 struct {
	Protocol        string             `json:"protocol"`
	ProtocolVersion string             `json:"protocolVersion"`
	Name            string             `json:"name"`
	EnsDomain       string             `json:"ensDomain,omitempty"`
	Endpoints       Agent6022Endpoints `json:"endpoints"`
	PaymentMethods  []string           `json:"paymentMethods,omitempty"`
}

// Agent6022Endpoints maps each API dialect the agent speaks to its URL; an agent publishes only the ones it serves.
type Agent6022Endpoints struct {
	Responses       string `json:"responses,omitempty"`
	ChatCompletions string `json:"chatCompletions,omitempty"`
}
