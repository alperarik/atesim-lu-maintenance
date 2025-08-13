package core_test

import (
	"database/sql"
	"os"
	"testing"

	. "svrn.com/pluto/core"
)

func TestDeviceOperations(t *testing.T) {
	// Setup test database
	dbPath := "test_pluto.db"
	os.Remove(dbPath) // Clean up any previous test db

	server := &PlutoServer{
		Devices:   make(map[string]*Device),
		Threshold: 10,
	}

	// Initialize test database
	var err error
	server.Db, err = sql.Open("sqlite3", dbPath+"?_crypto_key="+PlutoDBPassword)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer os.Remove(dbPath)
	defer server.Db.Close()

	// Initialize database tables
	if err := server.InitDB(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Test handleStartup with new device
	response := server.HandleStartup("192.168.1.1")
	if response != 1 {
		t.Errorf("Expected response 1 for new device, got %d", response)
	}
	if len(server.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(server.Devices))
	}

	// Test handleStartup with existing device
	response = server.HandleStartup("192.168.1.1")
	if response != 1 {
		t.Errorf("Expected response 1 for existing device, got %d", response)
	}

	// Test handleCountIncrement below threshold
	response = server.HandleCountIncrement("192.168.1.1", 5)
	if response != 0 {
		t.Errorf("Expected no response for increment below threshold, got %d", response)
	}

	// Test handleCountIncrement crossing threshold
	response = server.HandleCountIncrement("192.168.1.1", 6)
	if response != 2 {
		t.Errorf("Expected response 2 for crossing threshold, got %d", response)
	}

	// Test printStats (just ensure it doesn't panic)
	server.PrintStats()
}
