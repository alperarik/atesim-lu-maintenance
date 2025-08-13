package main

import (
	"flag"
	"log"
	. "svrn.com/pluto/core"
)

func main() {
	port := flag.Int("udp-port", 8080, "UDP port to listen on")
	httpPort := flag.Int("http-port", 8081, "HTTP port for reload API")
	threshold := flag.Int("maintenance-threshold", 5000, "Count threshold value for current count")
	flag.Parse()
	server := &PlutoServer{
		Devices:   make(map[string]*Device),
		Threshold: *threshold,
	}

	if err := server.InitDB("pluto.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer server.Db.Close()

	if err := server.LoadDevices(); err != nil {
		log.Printf("Warning: Failed to load devices (this is normal on first run or after password change): %v", err)
	}

	if err := server.StartUDPServer(*port); err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}
	defer server.Conn.Close()

	server.StartHTTPReloadServer(*httpPort)

	server.StartPeriodicTasks()

	log.Printf("Pluto server ready - UDP port: %d, maintenance threshold: %d, HTTP reload port: %d", *port, *threshold, *httpPort)
	server.PrintStats()

	select {}
}
