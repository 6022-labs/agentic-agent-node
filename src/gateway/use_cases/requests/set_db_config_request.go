package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/gateway/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"
)

type SetDbConfigRequest struct {
	Driver   enums.DbDriver `json:"driver"`
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	Database string         `json:"database"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Setup    bool           `json:"setup"`
}

func (r *SetDbConfigRequest) ValidateAndSanitize() error {
	if !enums.GetAllDbDrivers().Contains(r.Driver) {
		return errors.NewValidationError("driver", "must be one of: "+enums.GetAllDbDrivers().String())
	}
	if len(strings.TrimSpace(r.Host)) == 0 {
		return errors.NewValidationError("host", "cannot be empty")
	}
	if r.Port <= 0 || r.Port > 65535 {
		return errors.NewValidationError("port", "must be between 1 and 65535")
	}
	if len(strings.TrimSpace(r.Database)) == 0 {
		return errors.NewValidationError("database", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Username)) == 0 {
		return errors.NewValidationError("username", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Password)) == 0 {
		return errors.NewValidationError("password", "cannot be empty")
	}
	return nil
}
