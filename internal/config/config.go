package config

import (
	"database-manager/internal/model"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Config struct {
	mu          sync.RWMutex
	DataDir     string              `json:"-"`
	Port        int                 `json:"port"`
	JWTSecret   string              `json:"jwtSecret"`
	User        model.User          `json:"user"`
	Connections []model.Connection  `json:"connections"`
	History     []model.QueryHistory `json:"history"`
}

var (
	globalConfig *Config
	once         sync.Once
)

func Init(dataDir string) (*Config, error) {
	var initErr error
	once.Do(func() {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			initErr = err
			return
		}
		cfg := &Config{
			DataDir:   dataDir,
			Port:      9090,
			JWTSecret: "db-manager-secret-key-change-me",
		}
		cfgFile := filepath.Join(dataDir, "config.json")
		if data, err := os.ReadFile(cfgFile); err == nil {
			json.Unmarshal(data, cfg)
		}
		// default admin user
		if cfg.User.Username == "" {
			cfg.User = model.User{
				Username:  "admin",
				Password:  "admin",
				CreatedAt: timeNow(),
			}
		}
		if cfg.Connections == nil {
			cfg.Connections = []model.Connection{}
		}
		if cfg.History == nil {
			cfg.History = []model.QueryHistory{}
		}
		cfg.save()
		globalConfig = cfg
	})
	return globalConfig, initErr
}

func Get() *Config {
	return globalConfig
}

func (c *Config) save() error {
	c.mu.RLock()
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	cfgFile := filepath.Join(c.DataDir, "config.json")
	return os.WriteFile(cfgFile, data, 0644)
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.save()
}

func (c *Config) GetConnections() []model.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]model.Connection, len(c.Connections))
	copy(result, c.Connections)
	return result
}

func (c *Config) GetConnection(id string) *model.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for i := range c.Connections {
		if c.Connections[i].ID == id {
			conn := c.Connections[i]
			return &conn
		}
	}
	return nil
}

func (c *Config) AddConnection(conn model.Connection) error {
	c.mu.Lock()
	c.Connections = append(c.Connections, conn)
	c.mu.Unlock()
	return c.save()
}

func (c *Config) UpdateConnection(conn model.Connection) error {
	c.mu.Lock()
	for i := range c.Connections {
		if c.Connections[i].ID == conn.ID {
			c.Connections[i] = conn
			break
		}
	}
	c.mu.Unlock()
	return c.save()
}

func (c *Config) DeleteConnection(id string) error {
	c.mu.Lock()
	for i := range c.Connections {
		if c.Connections[i].ID == id {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			break
		}
	}
	c.mu.Unlock()
	return c.save()
}

func (c *Config) AddHistory(h model.QueryHistory) {
	c.mu.Lock()
	c.History = append([]model.QueryHistory{h}, c.History...)
	if len(c.History) > 1000 {
		c.History = c.History[:1000]
	}
	c.mu.Unlock()
	c.save()
}

func (c *Config) GetHistory(connID string, limit int) []model.QueryHistory {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []model.QueryHistory
	for _, h := range c.History {
		if connID == "" || h.ConnID == connID {
			result = append(result, h)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

func timeNow() time.Time {
	return time.Now()
}
