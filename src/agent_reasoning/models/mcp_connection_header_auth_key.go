package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MCPConnectionHeaderAuthKey string

func (k *MCPConnectionHeaderAuthKey) String() string {
	return string(*k)
}

// MCPConnectionHeaderAuthKeys is the persisted list of auth header names,
// stored as a JSON string array. Matching values live in Vault.
type MCPConnectionHeaderAuthKeys []MCPConnectionHeaderAuthKey

func (k *MCPConnectionHeaderAuthKeys) FromArray(values []string) {
	*k = make(MCPConnectionHeaderAuthKeys, len(values))
	for i, v := range values {
		(*k)[i] = MCPConnectionHeaderAuthKey(v)
	}
}

func (k MCPConnectionHeaderAuthKeys) Value() (driver.Value, error) {
	out := make([]string, len(k))
	for i, v := range k {
		out[i] = string(v)
	}
	return json.Marshal(out)
}

// GormDBDataType picks the per-dialect column type (`jsonb` on postgres,
// `json` on mysql/mariadb) instead of the `bytea`/`BLOB` that `[]byte` reflects to.
func (MCPConnectionHeaderAuthKeys) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "jsonb"
	case "mysql":
		return "json"
	case "sqlite":
		return "text"
	default:
		return ""
	}
}

func (k *MCPConnectionHeaderAuthKeys) Scan(value any) error {
	if value == nil {
		*k = MCPConnectionHeaderAuthKeys{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string: // sqlite returns text columns as string
		bytes = []byte(v)
	default:
		return errors.New("MCPConnectionHeaderAuthKeys.Scan: expected []byte or string")
	}
	var values []string
	if err := json.Unmarshal(bytes, &values); err != nil {
		return err
	}
	k.FromArray(values)
	return nil
}
