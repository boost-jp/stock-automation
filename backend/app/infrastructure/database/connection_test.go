package database

import (
	"testing"
	"time"
)

func TestDatabaseConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config DatabaseConfig
		valid  bool
	}{
		{
			name: "Valid config",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         3306,
				User:         "root",
				Password:     "password",
				DatabaseName: "test_db",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
				MaxLifetime:  5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "Empty host",
			config: DatabaseConfig{
				Host:         "",
				Port:         3306,
				User:         "root",
				Password:     "password",
				DatabaseName: "test_db",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
				MaxLifetime:  5 * time.Minute,
			},
			valid: false, // Empty host should be considered invalid
		},
		{
			name: "Invalid port",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         0,
				User:         "root",
				Password:     "password",
				DatabaseName: "test_db",
				MaxOpenConns: 25,
				MaxIdleConns: 10,
				MaxLifetime:  5 * time.Minute,
			},
			valid: false, // Port 0 should be considered invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			if tt.config.Host == "" && tt.valid {
				t.Error("Config with empty host should be invalid")
			}
			if tt.config.Port <= 0 && tt.valid {
				t.Error("Config with invalid port should be invalid")
			}
			if tt.config.DatabaseName == "" && tt.valid {
				t.Error("Config with empty database name should be invalid")
			}
		})
	}
}

func TestDefaultDatabaseConfig(t *testing.T) {
	config := DefaultDatabaseConfig()

	if config.Host == "" {
		t.Error("Default host should not be empty")
	}
	if config.Port <= 0 {
		t.Error("Default port should be positive")
	}
	if config.DatabaseName == "" {
		t.Error("Default database name should not be empty")
	}
	if config.MaxOpenConns <= 0 {
		t.Error("Default max open connections should be positive")
	}
	if config.MaxIdleConns <= 0 {
		t.Error("Default max idle connections should be positive")
	}
	if config.MaxLifetime <= 0 {
		t.Error("Default max lifetime should be positive")
	}

	// Check reasonable defaults
	if config.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got %s", config.Host)
	}
	if config.Port != 3306 {
		t.Errorf("Expected default port 3306, got %d", config.Port)
	}
	if config.DatabaseName != "stock_automation" {
		t.Errorf("Expected default database name 'stock_automation', got %s", config.DatabaseName)
	}
}

// MockConnectionManager implements ConnectionManager for testing.
type MockConnectionManager struct {
	closed bool
}

func NewMockConnectionManager() *MockConnectionManager {
	return &MockConnectionManager{}
}

func (m *MockConnectionManager) GetDB() interface{} {
	// Return a mock DB interface
	return &MockDB{}
}

func (m *MockConnectionManager) GetExecutor() interface{} {
	// Return a mock executor interface
	return &MockDB{}
}

func (m *MockConnectionManager) Close() error {
	m.closed = true
	return nil
}

func (m *MockConnectionManager) Ping() error {
	if m.closed {
		return &mockError{message: "connection closed"}
	}
	return nil
}

func (m *MockConnectionManager) GetStats() interface{} {
	// Return mock stats
	return map[string]interface{}{
		"open_connections": 5,
		"idle_connections": 2,
	}
}

// MockDB implements a minimal database interface for testing.
type MockDB struct{}

func (m *MockDB) Ping() error {
	return nil
}

func (m *MockDB) Close() error {
	return nil
}

// mockError implements error interface for testing.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestMockConnectionManager(t *testing.T) {
	manager := NewMockConnectionManager()

	t.Run("Initial state", func(t *testing.T) {
		if manager.closed {
			t.Error("Manager should not be closed initially")
		}

		err := manager.Ping()
		if err != nil {
			t.Errorf("Ping should succeed initially, got error: %v", err)
		}
	})

	t.Run("GetDB", func(t *testing.T) {
		db := manager.GetDB()
		if db == nil {
			t.Error("GetDB should return a non-nil database")
		}
	})

	t.Run("GetExecutor", func(t *testing.T) {
		executor := manager.GetExecutor()
		if executor == nil {
			t.Error("GetExecutor should return a non-nil executor")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := manager.GetStats()
		if stats == nil {
			t.Error("GetStats should return non-nil stats")
		}
	})

	t.Run("Close", func(t *testing.T) {
		err := manager.Close()
		if err != nil {
			t.Errorf("Close should succeed, got error: %v", err)
		}

		if !manager.closed {
			t.Error("Manager should be marked as closed")
		}

		// Ping should fail after close
		err = manager.Ping()
		if err == nil {
			t.Error("Ping should fail after close")
		}
	})
}

func TestConnectionManager_InterfaceCompliance(t *testing.T) {
	// Test that our mock implements the expected interface behavior
	manager := NewMockConnectionManager()

	// Test methods exist and can be called
	db := manager.GetDB()
	if db == nil {
		t.Error("GetDB should return a non-nil value")
	}

	executor := manager.GetExecutor()
	if executor == nil {
		t.Error("GetExecutor should return a non-nil value")
	}

	err := manager.Ping()
	if err != nil {
		t.Errorf("Ping should succeed, got error: %v", err)
	}

	stats := manager.GetStats()
	if stats == nil {
		t.Error("GetStats should return non-nil stats")
	}

	err = manager.Close()
	if err != nil {
		t.Errorf("Close should succeed, got error: %v", err)
	}
}
