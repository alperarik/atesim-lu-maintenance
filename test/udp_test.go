package core_test

import (
	"database/sql"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	. "svrn.com/pluto/core"
)

func TestUDPServer(t *testing.T) {
	// Setup test database
	dbPath := "test_udp.db"
	os.Remove(dbPath)

	server := &PlutoServer{
		Devices:   make(map[string]*Device),
		Threshold: 5,
	}

	// Initialize database
	var err error
	server.Db, err = sql.Open("sqlite3", dbPath+"?_crypto_key="+PlutoDBPassword)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer os.Remove(dbPath)
	defer server.Db.Close()

	if err := server.InitDB(dbPath); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Start test server
	serverConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	// Response channel for synchronization
	respChan := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	stopChan := make(chan struct{})

	server.Conn = serverConn

	// Start UDP server
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				buffer := make([]byte, 64)
				serverConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, addr, err := serverConn.ReadFromUDP(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					if !strings.Contains(err.Error(), "closed") {
						t.Logf("UDP read error: %v", err)
					}
					continue
				}

				message := strings.TrimSpace(string(buffer[:n]))
				deviceIP := addr.IP.String()

				var response StartupResponse
				switch message {
				case "0":
					response = server.HandleStartup(deviceIP)
				default:
					increment, err := strconv.Atoi(message)
					if err != nil {
						continue
					}
					response = server.HandleCountIncrement(deviceIP, increment)
				}

				respChan <- strconv.Itoa(int(response))
			}
		}
	}()

	// Test client
	clientConn, err := net.DialUDP("udp", nil, serverConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		close(stopChan)
		serverConn.Close()
		t.Fatalf("Failed to create client: %v", err)
	}
	defer clientConn.Close()

	tests := []struct {
		name     string
		message  string
		wantResp string
	}{
		{"Startup", "0", strconv.Itoa(int(StartupResponseNormal))},
		{"Increment", "3", strconv.Itoa(int(StartupResponseNormal))},
		{"Threshold", "10", strconv.Itoa(int(StartupResponseThresholdReached))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = clientConn.Write([]byte(tt.message))
			if err != nil {
				t.Fatalf("Failed to send message: %v", err)
			}

			if tt.wantResp == "" {
				return
			}

			select {
			case resp := <-respChan:
				if resp != tt.wantResp {
					t.Errorf("Expected response '%s', got '%s'", tt.wantResp, resp)
				}
			case <-time.After(1 * time.Second):
				t.Errorf("Timeout waiting for response to '%s'", tt.message)
			}
		})
	}

	// Cleanup
	close(stopChan)
	serverConn.Close()
	wg.Wait()
	close(respChan)
}
