package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringMap represents a map[string]string that can be stored as JSON in the database
type StringMap map[string]string

// Scan implements the sql.Scanner interface
func (sm *StringMap) Scan(value interface{}) error {
	if value == nil {
		*sm = StringMap{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringMap value")
	}

	return json.Unmarshal(bytes, sm)
}

// Value implements the driver.Valuer interface
func (sm StringMap) Value() (driver.Value, error) {
	if sm == nil {
		return nil, nil
	}
	return json.Marshal(sm)
}

// Config represents the application configuration
type Config struct {
	Settings *Settings `json:"settings"`
}
