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

// MCPConnectionHeaderAuthKeys is the persisted list of header names the
// MCPConnection authenticates with. Stored as a JSON string array (jsonb on
// postgres, JSON on mysql/mariadb). The matching values live in Vault.
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

// GormDBDataType emits the right column type per dialect so gormschema /
// atlas generate `jsonb` on postgres and `json` on mysql/mariadb instead of
// the default `bytea`/`BLOB` that `[]byte` would otherwise reflect to.
func (MCPConnectionHeaderAuthKeys) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "jsonb"
	case "mysql":
		return "json"
	default:
		return ""
	}
}

func (k *MCPConnectionHeaderAuthKeys) Scan(value any) error {
	if value == nil {
		*k = MCPConnectionHeaderAuthKeys{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("MCPConnectionHeaderAuthKeys.Scan: expected []byte")
	}
	var values []string
	if err := json.Unmarshal(bytes, &values); err != nil {
		return err
	}
	k.FromArray(values)
	return nil
}
