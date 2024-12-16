package system

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// Service 系统服务接口
type Service interface {
	// GetStatus 获取系统状态
	GetStatus() (*Status, error)
	// GetProcessInfo 获取进程信息
	GetProcessInfo() (*ProcessInfo, error)
	// Restart 重启服务
	Restart() error
	// Start 启动服务
	Start() error
	// Stop 停止服务
	Stop() error
}

// Status 系统状态
type Status struct {
	CPU     float64 `json:"cpu"`     // CPU使用率
	Memory  float64 `json:"memory"`  // 内存使用率
	Uptime  int64   `json:"uptime"`  // 运行时间(秒)
	NumGo   int     `json:"numGo"`   // Go协程数量
	Version string  `json:"version"` // 版本号
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID        int32     `json:"pid"`        // 进程ID
	CPUPercent float64   `json:"cpuPercent"` // CPU使用率
	MemPercent float32   `json:"memPercent"` // 内存使用率
	RSS        uint64    `json:"rss"`        // 物理内存使用
	VMS        uint64    `json:"vms"`        // 虚拟内存使用
	StartTime  time.Time `json:"startTime"`  // 启动时间
}

type service struct {
	startTime time.Time
	version   string
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewService 创建系统服务
func NewService() Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		startTime: time.Now(),
		version:   "1.0.0",
		running:   false,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// GetStatus 获取系统状态
func (s *service) GetStatus() (*Status, error) {
	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	// 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &Status{
		CPU:     cpuPercent[0],
		Memory:  memInfo.UsedPercent,
		Uptime:  int64(time.Since(s.startTime).Seconds()),
		NumGo:   runtime.NumGoroutine(),
		Version: s.version,
	}, nil
}

// GetProcessInfo 获取进程信息
func (s *service) GetProcessInfo() (*ProcessInfo, error) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	cpuPercent, err := p.CPUPercent()
	if err != nil {
		return nil, err
	}

	memPercent, err := p.MemoryPercent()
	if err != nil {
		return nil, err
	}

	memInfo, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}

	createTime, err := p.CreateTime()
	if err != nil {
		return nil, err
	}

	return &ProcessInfo{
		PID:        p.Pid,
		CPUPercent: cpuPercent,
		MemPercent: memPercent,
		RSS:        memInfo.RSS,
		VMS:        memInfo.VMS,
		StartTime:  time.Unix(createTime/1000, 0),
	}, nil
}

// Restart 重启服务
func (s *service) Restart() error {
	// TODO: 实现服务重启逻辑
	return nil
}

// Start 启动服务
func (s *service) Start() error {
	if s.running {
		return nil
	}

	// 启动系统监控
	go s.monitorLoop()

	s.running = true
	Info("系统监控已启动")
	return nil
}

// Stop 停止服务
func (s *service) Stop() error {
	if !s.running {
		return nil
	}

	// 停止监控
	s.cancel()
	s.running = false
	Info("系统监控已停止")
	return nil
}

// monitorLoop 系统监控循环
func (s *service) monitorLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// 获取系统状态
			status, err := s.GetStatus()
			if err != nil {
				Error("获取系统状态失败", zap.Error(err))
				continue
			}

			// 获取进程信息
			procInfo, err := s.GetProcessInfo()
			if err != nil {
				Error("获取进程信息失败", zap.Error(err))
				continue
			}

			// 记录监控信息
			Debug("系统状态",
				zap.Float64("cpu", status.CPU),
				zap.Float64("memory", status.Memory),
				zap.Int64("uptime", status.Uptime),
				zap.Int("goroutines", status.NumGo),
			)

			Debug("进程信息",
				zap.Float64("cpu", procInfo.CPUPercent),
				zap.Float32("memory", procInfo.MemPercent),
				zap.Uint64("rss", procInfo.RSS),
				zap.Uint64("vms", procInfo.VMS),
			)
		}
	}
}
