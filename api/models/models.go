package models

// Stats represents traffic statistics
type Stats struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Latency  int64 `json:"latency"`
}

// TrafficStats represents overall traffic statistics
type TrafficStats struct {
	Upload    int64            `json:"upload"`
	Download  int64            `json:"download"`
	NodeStats map[string]Stats `json:"node_stats"`
}
