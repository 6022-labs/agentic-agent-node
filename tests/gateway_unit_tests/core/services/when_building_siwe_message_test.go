package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type WhenBuildingSiweMessageTestingSuite struct {
	sut *services.SiweMessageBuilder
}

func WhenBuildingSiweMessageBeforeEach() *WhenBuildingSiweMessageTestingSuite {
	return &WhenBuildingSiweMessageTestingSuite{
		sut: services.NewSiweMessageBuilder(&settings.SiweSettings{
			Domain:    "agent.example.com",
			Uri:       "https://agent.example.com",
			Statement: "Sign in to your 6022 agent node.",
		}),
	}
}

func TestWhenBuildingSiweMessage(t *testing.T) {
	t.Parallel()

	t.Run("Given a challenge context", func(t *testing.T) {
		t.Parallel()

		address := common.HexToAddress("0xd389d997df5654b87914492e1e0c6861bfe835de")
		issuedAt := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
		expiresAt := issuedAt.Add(5 * time.Minute)

		t.Run("Should render the canonical EIP-4361 message", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingSiweMessageBeforeEach()

			message := suite.sut.Build(address, 80002, "nonce-abc", issuedAt, expiresAt)

			assert.Equal(t,
				"agent.example.com wants you to sign in with your Ethereum account:\n"+
					"0xd389d997Df5654b87914492E1e0C6861bfe835DE\n"+
					"\n"+
					"Sign in to your 6022 agent node.\n"+
					"\n"+
					"URI: https://agent.example.com\n"+
					"Version: 1\n"+
					"Chain ID: 80002\n"+
					"Nonce: nonce-abc\n"+
					"Issued At: 2026-07-17T10:00:00Z\n"+
					"Expiration Time: 2026-07-17T10:05:00Z",
				message,
			)
		})

		t.Run("Should render times in UTC", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingSiweMessageBeforeEach()
			paris := time.FixedZone("CEST", 2*3600)

			message := suite.sut.Build(address, 80002, "nonce-abc", issuedAt.In(paris), expiresAt.In(paris))

			assert.Contains(t, message, "Issued At: 2026-07-17T10:00:00Z")
			assert.Contains(t, message, "Expiration Time: 2026-07-17T10:05:00Z")
		})
	})
}
