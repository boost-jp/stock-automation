package config

import (
	"os"
	"strconv"
	"time"

	"github.com/boost-jp/stock-automation/app/infrastructure/database"
)

// Config holds application configuration.
type Config struct {
	Database DatabaseConfig `json:"database"`
	Yahoo    YahooConfig    `json:"yahoo"`
	Server   ServerConfig   `json:"server"`
	Log      LogConfig      `json:"log"`
	Slack    SlackConfig    `json:"slack"`
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	User         string        `json:"user"`
	Password     string        `json:"password"`
	DatabaseName string        `json:"database_name"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
}

// YahooConfig holds Yahoo Finance API configuration.
type YahooConfig struct {
	BaseURL       string        `json:"base_url"`
	Timeout       time.Duration `json:"timeout"`
	RetryCount    int           `json:"retry_count"`
	RetryWaitTime time.Duration `json:"retry_wait_time"`
	RetryMaxWait  time.Duration `json:"retry_max_wait"`
	RateLimitRPS  int           `json:"rate_limit_rps"`
	UserAgent     string        `json:"user_agent"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputPath string `json:"output_path"`
}

// SlackConfig holds Slack notification configuration.
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 3306),
			User:         getEnv("DB_USER", "root"),
			Password:     getEnv("DB_PASSWORD", ""),
			DatabaseName: getEnv("DB_NAME", "stock_automation"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute),
		},
		Yahoo: YahooConfig{
			BaseURL:       getEnv("YAHOO_BASE_URL", "https://query1.finance.yahoo.com"),
			Timeout:       getEnvAsDuration("YAHOO_TIMEOUT", 30*time.Second),
			RetryCount:    getEnvAsInt("YAHOO_RETRY_COUNT", 3),
			RetryWaitTime: getEnvAsDuration("YAHOO_RETRY_WAIT", 1*time.Second),
			RetryMaxWait:  getEnvAsDuration("YAHOO_RETRY_MAX_WAIT", 10*time.Second),
			RateLimitRPS:  getEnvAsInt("YAHOO_RATE_LIMIT_RPS", 10),
			UserAgent:     getEnv("YAHOO_USER_AGENT", "Mozilla/5.0 (compatible; StockAutomation/1.0)"),
		},
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT", "stdout"),
		},
		Slack: SlackConfig{
			WebhookURL: getEnv("SLACK_WEBHOOK_URL", ""),
			Channel:    getEnv("SLACK_CHANNEL", "#general"),
			Username:   getEnv("SLACK_USERNAME", "Stock Bot"),
		},
	}
}

// Load loads configuration from file (for compatibility)
func Load(path string) (*Config, error) {
	// For now, just return LoadConfig() which loads from environment
	// In the future, this could be extended to load from YAML/JSON files
	return LoadConfig(), nil
}

// ToDatabaseConfig converts config to database.DatabaseConfig.
func (c *Config) ToDatabaseConfig() database.DatabaseConfig {
	return database.DatabaseConfig{
		Host:         c.Database.Host,
		Port:         c.Database.Port,
		User:         c.Database.User,
		Password:     c.Database.Password,
		DatabaseName: c.Database.DatabaseName,
		MaxOpenConns: c.Database.MaxOpenConns,
		MaxIdleConns: c.Database.MaxIdleConns,
		MaxLifetime:  c.Database.MaxLifetime,
	}
}

// Helper functions for environment variable parsing.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
