// Package proxy provides proxy service implementation
package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
)

type service struct {
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
}

// NewService creates a new proxy service
func NewService() types.ProxyService {
	return &service{}
}

// Start starts the proxy service
func (s *service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// Get sing-box path
	singboxPath := "bin/sing-box"
	if _, err := os.Stat(singboxPath); err != nil {
		return fmt.Errorf("sing-box not found: %v", err)
	}

	// Check config file
	configPath := "config/sing-box.json"
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found: %v", err)
	}

	// Start sing-box
	cmd := exec.Command(singboxPath, "run", "-c", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sing-box: %v", err)
	}

	s.cmd = cmd
	s.running = true

	system.Info("Proxy service started")
	return nil
}

// Stop stops the proxy service
func (s *service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop sing-box: %v", err)
		}
		s.cmd.Wait()
	}

	s.cmd = nil
	s.running = false

	system.Info("Proxy service stopped")
	return nil
}

// Restart restarts the proxy service
func (s *service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}
