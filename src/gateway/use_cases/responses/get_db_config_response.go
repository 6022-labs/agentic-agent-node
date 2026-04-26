package responses

import "github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"

type GetDbConfigResponse struct {
	Driver    enums.DbDriver       `json:"driver"`
	Host      string               `json:"host"`
	Port      int                  `json:"port"`
	Database  string               `json:"database"`
	Username  string               `json:"username"`
	Status    enums.DbSetupStatus  `json:"status"`
	LastError string               `json:"lastError,omitempty"`
}
