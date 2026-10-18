package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingListConversationsRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given no pagination at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fall back to the first page and the default size", func(t *testing.T) {
			t.Parallel()

			req := &requests.ListConversationsRequest{}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, uint(1), *req.Page)
			assert.Equal(t, uint(10), *req.ElementsPerPage)
		})
	})

	t.Run("Given out-of-range pagination", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clamp rather than refuse", func(t *testing.T) {
			t.Parallel()

			req := &requests.ListConversationsRequest{
				Page:            lo.ToPtr(uint(0)),
				ElementsPerPage: lo.ToPtr(uint(5000)),
			}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, uint(1), *req.Page)
			assert.Equal(t, uint(100), *req.ElementsPerPage)
		})
	})

	t.Run("Given a page within range", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep it untouched", func(t *testing.T) {
			t.Parallel()

			req := &requests.ListConversationsRequest{
				Page:            lo.ToPtr(uint(3)),
				ElementsPerPage: lo.ToPtr(uint(25)),
			}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, uint(3), *req.Page)
			assert.Equal(t, uint(25), *req.ElementsPerPage)
		})
	})
}
