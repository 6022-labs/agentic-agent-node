package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// MCPHeaderAuthSecretStorageInterface persists header-auth values per MCP
// connection. Each (connectionId, key) pair maps to a separate Vault secret
// path; the persisted MCPConnection row holds only the key names.
type MCPHeaderAuthSecretStorageInterface interface {
	// GetValue returns the header value, or ("", false, nil) if none is set.
	GetValue(ctx context.Context, connectionId uuid.UUID, key string) (string, bool, error)
	SaveValue(ctx context.Context, connectionId uuid.UUID, key string, value string) error
	// DeleteValue clears the header value. No-op if none is set.
	DeleteValue(ctx context.Context, connectionId uuid.UUID, key string) error
}
