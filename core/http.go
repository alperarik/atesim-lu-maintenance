package core

import (
	"fmt"
	"log"
	"net/http"
)

func (p *PlutoServer) StartHTTPReloadServer(port int) {
	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed - use POST", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Manual device reload triggered via HTTP API")

		rows, err := p.Db.Query("SELECT ip, current_count, total_count, last_seen, registered_at FROM devices")
		if err != nil {
			log.Printf("Error reloading devices from database: %v", err)
			http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		updatedCount := 0
		errorCount := 0

		for rows.Next() {
			var device Device
			var lastSeen, registeredAt string

			err := rows.Scan(&device.IP, &device.CurrentCount, &device.TotalCount, &lastSeen, &registeredAt)
			if err != nil {
				log.Printf("Error scanning device row during reload: %v", err)
				errorCount++
				continue
			}

			device.LastSeen = parseTime(lastSeen)
			device.RegisteredAt = parseTime(registeredAt)

			if existingDevice, exists := p.Devices[device.IP]; exists {
				if existingDevice.CurrentCount != device.CurrentCount || existingDevice.TotalCount != device.TotalCount {
					log.Printf("Updating device %s: current %d->%d, total %d->%d",
						device.IP, existingDevice.CurrentCount, device.CurrentCount,
						existingDevice.TotalCount, device.TotalCount)
				}
			} else {
				log.Printf("Loading device %s: current=%d, total=%d", device.IP, device.CurrentCount, device.TotalCount)
			}

			p.Devices[device.IP] = &device
			updatedCount++
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error during device reload iteration: %v", err)
			http.Error(w, fmt.Sprintf("Database iteration failed: %v", err), http.StatusInternalServerError)
			return
		}

		responseMsg := fmt.Sprintf("Device reload completed successfully. Processed: %d devices", updatedCount)
		if errorCount > 0 {
			responseMsg += fmt.Sprintf(" (with %d errors - check logs)", errorCount)
		}

		log.Printf("Device reload completed: %d devices processed, %d errors", updatedCount, errorCount)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseMsg))
	})

	log.Printf("HTTP reload API server starting on port %d (endpoint: POST /reload)", port)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Printf("HTTP reload server error: %v", err)
		}
	}()
}
