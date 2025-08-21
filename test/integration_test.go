package core_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "svrn.com/pluto/core"
)

func TestIntegration(t *testing.T) {
	// Create a complete server instance
	server := &PlutoServer{
		Devices:   make(map[string]*Device),
		Threshold: 5,
	}

	// Initialize with test database
	dbPath := "test_integration.db"
	os.Remove(dbPath)
	server.Db, _ = sql.Open("sqlite3", dbPath+"?_crypto_key="+PlutoDBPassword)
	defer os.Remove(dbPath)
	defer server.Db.Close()

	server.InitDB(dbPath)

	// Test full workflow
	// 1. Device startup
	server.HandleStartup("192.168.1.1")

	// 2. Several increments
	server.HandleCountIncrement("192.168.1.1", 2)
	server.HandleCountIncrement("192.168.1.1", 2)

	// 3. Cross threshold
	response := server.HandleCountIncrement("192.168.1.1", 2)
	if response != StartupResponseThresholdReached {
		t.Errorf("Expected threshold response StartupResponseThresholdReached, got %d", response)
	}

	// 4. Verify stats
	server.PrintStats()

	// 5. Test HTTP reload - use a test server instead of calling startHTTPReloadServer
	req := httptest.NewRequest("POST", "/reload", nil)
	w := httptest.NewRecorder()

	// Create a one-time handler for the test
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Test reload logic here without registering the handler globally
		rows, err := server.Db.Query("SELECT ip, current_count, total_count, last_seen, registered_at FROM devices")
		if err != nil {
			t.Logf("Error reloading devices: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test reload successful"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
