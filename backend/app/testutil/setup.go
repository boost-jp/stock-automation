package testutil

import (
	"database/sql"
	"testing"
)

// SetupTestDB creates a test database and returns the connection and cleanup function
// This is a convenience function that wraps NewTestDB for backward compatibility
func SetupTestDB() (*sql.DB, func(), error) {
	// Create a temporary testing.T for this setup
	t := &testing.T{}

	testDB := NewTestDB(t)

	cleanup := func() {
		if err := testDB.Cleanup(); err != nil {
			// Log error but don't panic in cleanup
			t.Logf("Failed to cleanup test database: %v", err)
		}
	}

	return testDB.GetDB(), cleanup, nil
}

// SetupTestDBWithT creates a test database using the provided testing.T
func SetupTestDBWithT(t *testing.T) (*sql.DB, func(), error) {
	t.Helper()

	testDB := NewTestDB(t)

	cleanup := func() {
		if err := testDB.Cleanup(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}

	return testDB.GetDB(), cleanup, nil
}
