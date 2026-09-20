package images_test

import (
	"encoding/xml"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services/images"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenGeneratingDefaultImage(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent with a base image, wallets and stars", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an animated SVG embedding the image, gradient and stars", func(t *testing.T) {
			t.Parallel()

			imageGenerationSettings := &settings.ImageGenerationSettings{ImageMarkPath: nil}
			sut := images.NewDefaultImageGenerator(imageGenerationSettings)

			result, err := sut.Generate(sampleImageParams(t))

			assert.NoError(t, err)
			svg := string(result)
			assert.Contains(t, svg, "<svg")
			assert.Contains(t, svg, "data:image/png;base64,")
			assert.Contains(t, svg, "url(#borderGradient)")
			assert.Contains(t, svg, "★★★")
			assert.Contains(t, svg, "<animate")
		})
	})

	t.Run("Given text that carries XML-sensitive characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should escape them so the SVG stays well-formed", func(t *testing.T) {
			t.Parallel()

			sut := images.NewDefaultImageGenerator(&settings.ImageGenerationSettings{})
			params := sampleImageParams(t)
			params.Name = "tom&jerry"
			params.Role = "<expert>"
			params.Wallets = []string{`0x"quoted"`}

			result, err := sut.Generate(params)

			assert.NoError(t, err)
			svg := string(result)
			assert.Contains(t, svg, "tom&amp;jerry")
			assert.Contains(t, svg, "Role: &lt;expert&gt;")
			assert.NoError(t, xml.Unmarshal(result, new(struct{})))
		})
	})

	t.Run("Given an origin domain without a generation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should render the origin badge with generation 0 instead of panicking", func(t *testing.T) {
			t.Parallel()

			sut := images.NewDefaultImageGenerator(&settings.ImageGenerationSettings{})
			params := sampleImageParams(t)
			params.OriginEnsDomain = lo.ToPtr("origin.agent.eth")

			result, err := sut.Generate(params)

			assert.NoError(t, err)
			assert.Contains(t, string(result), "Origin: origin.agent.eth (0)")
		})
	})
}
