package images

import (
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
)

type GenerateAgentImages struct {
	defaultImageGenerator identity_interfaces.DefaultImageGeneratorInterface
	iconImageGenerator    identity_interfaces.IconImageGeneratorInterface
}

func NewGenerateAgentImages(
	defaultImageGenerator identity_interfaces.DefaultImageGeneratorInterface,
	iconImageGenerator identity_interfaces.IconImageGeneratorInterface,
) *GenerateAgentImages {
	return &GenerateAgentImages{
		defaultImageGenerator: defaultImageGenerator,
		iconImageGenerator:    iconImageGenerator,
	}
}

func (g *GenerateAgentImages) Generate(params identity_interfaces.GenerateImageParams) (map[string][]byte, error) {
	defaultBytes, err := g.defaultImageGenerator.Generate(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate default image: %w", err)
	}

	iconBytes, err := g.iconImageGenerator.Generate(params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate icon image: %w", err)
	}

	return map[string][]byte{
		"default": defaultBytes,
		"icon":    iconBytes,
	}, nil
}
