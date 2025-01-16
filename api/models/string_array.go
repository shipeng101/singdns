package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringArray is a custom type for handling string arrays in the database
type StringArray []string

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal StringArray value: %v", value)
	}

	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(a)
}

// MarshalJSON implements the json.Marshaler interface
func (a StringArray) MarshalJSON() ([]byte, error) {
	if a == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal([]string(a))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var s []string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = StringArray(s)
	return nil
}
