package api

import (
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// MonitorStats represents system and process statistics
type MonitorStats struct {
	Time time.Time `json:"time"`

	// System stats
	SystemCPU    float64 `json:"system_cpu"`
	SystemMemory float64 `json:"system_memory"`

	// Process stats
	ProcessCPU    float64 `json:"process_cpu"`
	ProcessMemory float64 `json:"process_memory"`
	NumGoroutine  int     `json:"num_goroutine"`
	NumCPU        int     `json:"num_cpu"`

	// Memory stats
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// Monitor represents a performance monitor
type Monitor struct {
	stats    *MonitorStats
	process  *process.Process
	interval time.Duration
	mu       sync.RWMutex
	done     chan struct{}
}

// NewMonitor creates a new performance monitor
func NewMonitor(interval time.Duration) (*Monitor, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	m := &Monitor{
		stats:    &MonitorStats{},
		process:  proc,
		interval: interval,
		done:     make(chan struct{}),
	}

	// Start monitoring
	go m.monitor()

	return m, nil
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.done)
}

// GetStats returns the current stats
func (m *Monitor) GetStats() *MonitorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// monitor periodically updates stats
func (m *Monitor) monitor() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.updateStats()
		}
	}
}

// updateStats updates all statistics
func (m *Monitor) updateStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.Time = time.Now()

	// Update system stats
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		m.stats.SystemCPU = cpuPercent[0]
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.stats.SystemMemory = memInfo.UsedPercent
	}

	// Update process stats
	if cpuPercent, err := m.process.CPUPercent(); err == nil {
		m.stats.ProcessCPU = cpuPercent
	}

	if memInfo, err := m.process.MemoryInfo(); err == nil {
		m.stats.ProcessMemory = float64(memInfo.RSS) / 1024 / 1024 // Convert to MB
	}

	m.stats.NumGoroutine = runtime.NumGoroutine()
	m.stats.NumCPU = runtime.NumCPU()

	// Update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.stats.Alloc = memStats.Alloc
	m.stats.TotalAlloc = memStats.TotalAlloc
	m.stats.Sys = memStats.Sys
	m.stats.NumGC = memStats.NumGC
}

// StartProfiling starts the pprof HTTP server
func StartProfiling(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			LogError(err, "Failed to start pprof server")
		}
	}()

	return nil
}

// StartMetrics starts the Prometheus metrics server
func StartMetrics(addr string) error {
	// Create a new registry
	registry := prometheus.NewRegistry()

	// Register default Go metrics
	registry.MustRegister(prometheus.NewBuildInfoCollector())
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	// Register custom metrics
	registry.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "singdns_goroutines",
				Help: "Number of goroutines",
			},
			func() float64 {
				return float64(runtime.NumGoroutine())
			},
		),
	)

	// Start metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			LogError(err, "Failed to start metrics server")
		}
	}()

	return nil
}
