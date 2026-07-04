package responses

// WellKnownAgentCard is the A2A v1.0 discovery document (ProtoJSON shape).
type WellKnownAgentCard struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	SupportedInterfaces []AgentInterface  `json:"supportedInterfaces"`
	Version             string            `json:"version"`
	Capabilities        AgentCapabilities `json:"capabilities"`
	DefaultInputModes   []string          `json:"defaultInputModes"`
	DefaultOutputModes  []string          `json:"defaultOutputModes"`
	Skills              []AgentSkill      `json:"skills"`
}

type AgentInterface struct {
	Url             string `json:"url"`
	ProtocolBinding string `json:"protocolBinding"`
	ProtocolVersion string `json:"protocolVersion"`
}

type AgentCapabilities struct {
	Streaming         bool `json:"streaming"`
	PushNotifications bool `json:"pushNotifications"`
}

type AgentSkill struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	InputModes  []string `json:"inputModes,omitempty"`
	OutputModes []string `json:"outputModes,omitempty"`
}
