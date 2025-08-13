package core_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	. "svrn.com/pluto/core"
)

func TestDatabaseOperations(t *testing.T) {
	// Setup
	server := &PlutoServer{
		Devices: make(map[string]*Device),
	}

	// Use a test database
	dbPath := "test_pluto.db"
	os.Remove(dbPath) // Clean up any previous test db
	server.Db, _ = sql.Open("sqlite3", dbPath+"?_crypto_key="+PlutoDBPassword)
	defer os.Remove(dbPath)
	defer server.Db.Close()

	// Initialize database
	if err := server.InitDB(dbPath); err != nil {
		t.Fatalf("initDB failed: %v", err)
	}

	// Test device operations
	device := &Device{
		IP:           "192.168.1.1",
		CurrentCount: 5,
		TotalCount:   10,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	// Test saveDevice
	if err := server.SaveDevice(device); err != nil {
		t.Errorf("saveDevice failed: %v", err)
	}

	// Test loadDevices
	server.Devices = make(map[string]*Device)
	if err := server.LoadDevices(); err != nil {
		t.Errorf("loadDevices failed: %v", err)
	}
	if len(server.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(server.Devices))
	}

	// Test saveLog
	if err := server.SaveLog("192.168.1.1", "test", 5, 1); err != nil {
		t.Errorf("saveLog failed: %v", err)
	}
}
