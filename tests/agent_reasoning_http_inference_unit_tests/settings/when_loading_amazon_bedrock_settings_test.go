package settings_test

import (
	"testing"

	settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	"github.com/stretchr/testify/assert"
)

func TestWhenLoadingAmazonBedrockSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build the settings", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAmazonBedrockSettings()

			assert.NotNil(t, result)
		})
	})
}
