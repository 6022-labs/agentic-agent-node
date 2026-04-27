package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func validSetDbConfigRequest() *requests.SetDbConfigRequest {
	return &requests.SetDbConfigRequest{
		Driver:   enums.DbDriverPostgres,
		Host:     "db.example.com",
		Port:     5432,
		Database: "agent",
		Username: "agent",
		Password: "secret",
	}
}

func TestWhenValidatingSetDbConfigRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a fully populated valid request", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()

		t.Run("Should pass validation", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given an unknown driver", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Driver = "oracle"

		t.Run("Should reject with a validation error on driver", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "driver")
		})
	})

	t.Run("Given the host is whitespace", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Host = "   "

		t.Run("Should reject with a validation error on host", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "host")
		})
	})

	t.Run("Given the port is zero", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Port = 0

		t.Run("Should reject with a validation error on port", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "port")
		})
	})

	t.Run("Given the port exceeds the TCP range", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Port = 70000

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "port")
		})
	})

	t.Run("Given the database name is empty", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Database = ""

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "database")
		})
	})

	t.Run("Given the username is whitespace", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Username = "  "

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "username")
		})
	})

	t.Run("Given the password is empty", func(t *testing.T) {
		t.Parallel()
		req := validSetDbConfigRequest()
		req.Password = ""

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "password")
		})
	})
}
