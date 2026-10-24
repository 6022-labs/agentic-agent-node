package enums_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestWhenSettingTransactionStatus(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown status", func(t *testing.T) {
		t.Parallel()
		var status enums.TransactionStatus

		t.Run("Should reject with an error mentioning the value", func(t *testing.T) {
			t.Parallel()

			err := status.Set("published")
			assert.ErrorContains(t, err, "invalid TransactionStatus: published")
		})
	})

	t.Run("Given a known status in mixed casing", func(t *testing.T) {
		t.Parallel()
		var status enums.TransactionStatus

		t.Run("Should normalize and assign it", func(t *testing.T) {
			t.Parallel()

			err := status.Set("Confirmed")
			assert.NoError(t, err)
			assert.Equal(t, enums.TransactionStatusConfirmed, status)
		})
	})

	t.Run("Given the full list of statuses", func(t *testing.T) {
		t.Parallel()
		statuses := enums.GetAllTransactionStatuses()

		t.Run("Should contain every known status and exclude unknowns", func(t *testing.T) {
			t.Parallel()

			assert.True(t, statuses.Contains(enums.TransactionStatusPending))
			assert.True(t, statuses.Contains(enums.TransactionStatusSubmitted))
			assert.True(t, statuses.Contains(enums.TransactionStatusConfirmed))
			assert.True(t, statuses.Contains(enums.TransactionStatusFailed))
			assert.False(t, statuses.Contains(enums.TransactionStatus("published")))
		})
	})
}
