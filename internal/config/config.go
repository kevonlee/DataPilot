package config

import (
	"database-manager/internal/logger"
	"database-manager/internal/model"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	mu          sync.RWMutex
	DataDir     string              `json:"-"`
	Port        int                 `json:"port"`
	LogLevel    string              `json:"logLevel"`
	JWTSecret   string              `json:"jwtSecret"`
	User        model.User          `json:"user"`
	Connections []model.Connection  `json:"connections"`
	History     []model.QueryHistory `json:"history"`
}

var (
	globalConfig *Config
	once         sync.Once
)

func generateSecret(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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
			LogLevel:  "info",
			JWTSecret: generateSecret(32),
		}
		cfgFile := filepath.Join(dataDir, "config.json")
		if data, err := os.ReadFile(cfgFile); err == nil {
			json.Unmarshal(data, cfg)
		}

		// apply log level
		if cfg.LogLevel != "" {
			logger.SetLevel(cfg.LogLevel)
		}

		// auto-generate JWT secret if empty or default
		if cfg.JWTSecret == "" || cfg.JWTSecret == "db-manager-secret-key-change-me" {
			cfg.JWTSecret = generateSecret(32)
		}

		// default admin user
		if cfg.User.Username == "" {
			hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
			cfg.User = model.User{
				Username:  "admin",
				Password:  string(hash),
				CreatedAt: timeNow(),
			}
		}

		// migrate plaintext password to bcrypt
		if cfg.User.Password != "" && !strings.HasPrefix(cfg.User.Password, "$2") {
			logger.Info("Migrating plaintext password to bcrypt")
			hash, err := bcrypt.GenerateFromPassword([]byte(cfg.User.Password), bcrypt.DefaultCost)
			if err == nil {
				cfg.User.Password = string(hash)
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
		logger.Info("Config loaded from %s, port=%d, logLevel=%s", dataDir, cfg.Port, cfg.LogLevel)
	})
	return globalConfig, initErr
}

func Get() *Config {
	return globalConfig
}

// save persists config to disk. Caller must hold at least RLock.
func (c *Config) save() error {
	data, err := json.MarshalIndent(c, "", "  ")
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
	defer c.mu.Unlock()
	c.Connections = append(c.Connections, conn)
	return c.save()
}

func (c *Config) UpdateConnection(conn model.Connection) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Connections {
		if c.Connections[i].ID == conn.ID {
			c.Connections[i] = conn
			break
		}
	}
	return c.save()
}

func (c *Config) DeleteConnection(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.Connections {
		if c.Connections[i].ID == id {
			c.Connections = append(c.Connections[:i], c.Connections[i+1:]...)
			break
		}
	}
	return c.save()
}

func (c *Config) AddHistory(h model.QueryHistory) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.History = append([]model.QueryHistory{h}, c.History...)
	if len(c.History) > 1000 {
		c.History = c.History[:1000]
	}
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

// CheckPassword verifies a plaintext password against the stored bcrypt hash.
func (c *Config) CheckPassword(plaintext string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.User.Password), []byte(plaintext))
	return err == nil
}

// SetPassword hashes and stores a new password.
func (c *Config) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.User.Password = string(hash)
	c.mu.Unlock()
	return c.Save()
}
