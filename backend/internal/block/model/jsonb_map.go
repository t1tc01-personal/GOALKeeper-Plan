package model

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONBMap is a custom type for handling JSONB fields in PostgreSQL
type JSONBMap map[string]any

// Value implements the driver.Valuer interface for JSONBMap
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONBMap
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONBMap)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		bytes = []byte("{}")
	}

	if len(bytes) == 0 {
		*j = make(JSONBMap)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

