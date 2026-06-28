package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenExtractingA2AMessageText(t *testing.T) {
	t.Parallel()

	t.Run("Given multiple text parts", func(t *testing.T) {
		t.Parallel()

		t.Run("Should concatenate them in order", func(t *testing.T) {
			t.Parallel()

			message := requests.A2AMessage{Parts: []requests.A2APart{
				{Text: "foo"},
				{Text: "bar"},
			}}

			assert.Equal(t, "foobar", message.Text())
		})
	})

	t.Run("Given parts without text (non-text variants)", func(t *testing.T) {
		t.Parallel()

		t.Run("Should ignore them and keep only text", func(t *testing.T) {
			t.Parallel()

			message := requests.A2AMessage{Parts: []requests.A2APart{
				{Text: "keep"},
				{Text: ""},
			}}

			assert.Equal(t, "keep", message.Text())
		})
	})
}
