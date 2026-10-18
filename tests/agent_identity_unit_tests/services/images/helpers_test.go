package images_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
)

// sampleImageParams builds the shared image params with a real PNG base.
func sampleImageParams(t *testing.T) identity_interfaces.GenerateImageParams {
	t.Helper()

	return identity_interfaces.GenerateImageParams{
		Name:              "my-agent",
		EnsDomain:         "my-agent.agent.eth",
		Role:              "assistant",
		Creator:           "0xabc",
		BaseImageBytes:    sampleBasePNG(t),
		BaseImageMimeType: "image/png",
		Wallets:           []string{"0x1111", "0x2222"},
		CreatedAt:         time.Unix(0, 0),
		Stars:             3,
	}
}

// sampleBasePNG builds a throwaway gradient PNG, avoiding a binary fixture.
func sampleBasePNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 472, 728))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 160, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode sample png: %v", err)
	}
	return buf.Bytes()
}
