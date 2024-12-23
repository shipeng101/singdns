package models

import "time"

// HealthData represents the health status of a node or service
type HealthData struct {
	Status    string    `json:"status"`    // online/offline
	Latency   int64     `json:"latency"`   // in milliseconds
	LastCheck time.Time `json:"lastCheck"` // last check time
	Error     string    `json:"error"`     // error message if any
}
