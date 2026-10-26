package repositories_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertUniqueViolation matches the driver-specific duplicate-key wording of postgres, mysql/mariadb and sqlite.
func assertUniqueViolation(t *testing.T, err error) {
	t.Helper()
	if assert.Error(t, err) {
		msg := strings.ToLower(err.Error())
		assert.True(
			t,
			strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique"),
			"not a unique violation: %v",
			err,
		)
	}
}
