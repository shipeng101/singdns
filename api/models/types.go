package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringArray represents a string array that can be stored as JSON in the database
type StringArray []string

// Scan implements the sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray value")
	}

	return json.Unmarshal(bytes, sa)
}

// Value implements the driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}
