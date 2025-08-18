package core

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func (p *PlutoServer) StartUDPServer(port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	p.Conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", port, err)
	}

	log.Printf("Pluto UDP server listening on port %d (threshold: %d)", port, p.Threshold)

	go p.handleUDPMessages()

	return nil
}

func (p *PlutoServer) handleUDPMessages() {
	buffer := make([]byte, 64)

	for {
		n, addr, err := p.Conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP message: %v", err)
			continue
		}

		deviceIP := addr.IP.String()
		message := strings.TrimSpace(string(buffer[:n]))

		var response StartupResponse

		switch message {
		case "0":
			response = p.HandleStartup(deviceIP)
		default:
			increment, err := strconv.Atoi(message)
			if err != nil {
				log.Printf("Invalid message from %s: '%s'", deviceIP, message)
				continue
			}

			if increment == 0 {
				increment = 1
			}

			response = p.HandleCountIncrement(deviceIP, increment)
		}

		if response > 0 {
			responseMsg := strconv.Itoa(int(response))
			_, err = p.Conn.WriteToUDP([]byte(responseMsg), addr)
			if err != nil {
				log.Printf("Error sending response to %s: %v", deviceIP, err)
			}
		}
	}
}
