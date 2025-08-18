package core

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (p *PlutoServer) InitDB(dbName string) error {
	var err error
	dbPath := fmt.Sprintf("%s?_crypto_key=%s", dbName, PlutoDBPassword)
	p.Db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Create devices table with two count columns
	createDevicesTable := `
	CREATE TABLE IF NOT EXISTS devices (
		ip TEXT PRIMARY KEY,
		current_count INTEGER NOT NULL DEFAULT 0,
		total_count INTEGER NOT NULL DEFAULT 0,
		last_seen DATETIME NOT NULL,
		registered_at DATETIME NOT NULL
	);`

	// Create logs table
	createLogsTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_ip TEXT NOT NULL,
		action TEXT NOT NULL,
		count_value INTEGER,
		timestamp DATETIME NOT NULL,
		response INTEGER,
		FOREIGN KEY (device_ip) REFERENCES devices (ip)
	);`

	if _, err = p.Db.Exec(createDevicesTable); err != nil {
		return fmt.Errorf("failed to create devices table: %v", err)
	}

	if _, err = p.Db.Exec(createLogsTable); err != nil {
		return fmt.Errorf("failed to create logs table: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

func (p *PlutoServer) LoadDevices() error {
	rows, err := p.Db.Query("SELECT ip, current_count, total_count, last_seen, registered_at FROM devices")
	if err != nil {
		return fmt.Errorf("failed to load devices: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var device Device
		var lastSeen, registeredAt string

		err := rows.Scan(&device.IP, &device.CurrentCount, &device.TotalCount, &lastSeen, &registeredAt)
		if err != nil {
			log.Printf("Error scanning device row: %v", err)
			continue
		}

		device.LastSeen = parseTime(lastSeen)
		device.RegisteredAt = parseTime(registeredAt)

		p.Devices[device.IP] = &device
	}

	log.Printf("Loaded %d devices from database", len(p.Devices))
	return nil
}

func parseTime(timeStr string) time.Time {

	utc3Location := time.FixedZone("UTC+3", 3*3600)
	t, err := time.ParseInLocation("15:04:05 02/01/2006", timeStr, utc3Location)
	if err == nil {
		return t.UTC()
	}

	log.Printf("Error parsing time string: %s", timeStr)
	return time.Now()
}

func (p *PlutoServer) SaveDevice(device *Device) error {
	query := `
	INSERT OR REPLACE INTO devices (ip, current_count, total_count, last_seen, registered_at)
	VALUES (?, ?, ?, ?, ?)`

	_, err := p.Db.Exec(query, device.IP, device.CurrentCount, device.TotalCount,
		device.LastSeen.Format("2006-01-02 15:04:05"),
		device.RegisteredAt.Format("2006-01-02 15:04:05"))

	if err != nil {
		return fmt.Errorf("failed to save device %s: %v", device.IP, err)
	}

	return nil
}

func (p *PlutoServer) SaveLog(deviceIP, action string, countValue, response int) error {
	query := `
	INSERT INTO logs (device_ip, action, count_value, timestamp, response)
	VALUES (?, ?, ?, ?, ?)`

	utc3Location := time.FixedZone("UTC+3", 3*3600)
	timestamp := time.Now().In(utc3Location).Format("15:04:05 02/01/2006")

	_, err := p.Db.Exec(query, deviceIP, action, countValue, timestamp, response)
	if err != nil {
		return fmt.Errorf("failed to save log for device %s: %v", deviceIP, err)
	}

	return nil
}
