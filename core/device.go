package core

import (
	"fmt"
	"log"
	"time"
)

// StartupResponse represents the possible responses for device startup
type StartupResponse int

const (
	StartupResponseNormal           StartupResponse = iota // 0
	StartupResponseThresholdReached                        // 1
)

func (p *PlutoServer) HandleStartup(deviceIP string) StartupResponse {
	now := time.Now()

	device, exists := p.Devices[deviceIP]
	if !exists {
		device = &Device{
			IP:           deviceIP,
			CurrentCount: 0,
			TotalCount:   0,
			LastSeen:     now,
			RegisteredAt: now,
		}
		p.Devices[deviceIP] = device
		log.Printf("New device registered: %s", deviceIP)
	} else {
		device.LastSeen = now
		log.Printf("Device startup: %s (current count: %d)", deviceIP, device.CurrentCount)
	}

	if err := p.SaveDevice(device); err != nil {
		log.Printf("Error saving device: %v", err)
	}

	response := StartupResponseNormal
	if device.CurrentCount >= p.Threshold {
		response = StartupResponseThresholdReached
	}

	if err := p.SaveLog(deviceIP, "startup", device.CurrentCount, int(response)); err != nil {
		log.Printf("Error saving log: %v", err)
	}

	return response
}

func (p *PlutoServer) HandleCountIncrement(deviceIP string, increment int) StartupResponse {
	now := time.Now()

	device, exists := p.Devices[deviceIP]
	if !exists {
		device = &Device{
			IP:           deviceIP,
			CurrentCount: 0,
			TotalCount:   0,
			LastSeen:     now,
			RegisteredAt: now,
		}
		p.Devices[deviceIP] = device
		log.Printf("Auto-registered device: %s", deviceIP)
	}

	oldCount := device.CurrentCount
	device.CurrentCount += increment
	device.TotalCount += increment
	device.LastSeen = now

	if err := p.SaveDevice(device); err != nil {
		log.Printf("Error saving device: %v", err)
	}

	response := StartupResponseNormal
	wasAbove := oldCount >= p.Threshold
	isAbove := device.CurrentCount >= p.Threshold

	if !wasAbove && isAbove {
		response = StartupResponseThresholdReached
		log.Printf("Device %s crossed threshold: %d -> %d", deviceIP, oldCount, device.CurrentCount)
	}

	action := fmt.Sprintf("increment+%d", increment)
	if err := p.SaveLog(deviceIP, action, device.CurrentCount, int(response)); err != nil {
		log.Printf("Error saving log: %v", err)
	}

	log.Printf("Count update %s: %d -> %d (Total: %d)", deviceIP, oldCount, device.CurrentCount, device.TotalCount)
	return response
}

func (p *PlutoServer) PrintStats() {
	totalDevices := len(p.Devices)
	activeDevices := 0
	belowThreshold := 0
	aboveThreshold := 0
	totalCurrentCount := 0
	totalAggregateCount := 0

	now := time.Now()
	for _, device := range p.Devices {
		totalCurrentCount += device.CurrentCount
		totalAggregateCount += device.TotalCount

		if now.Sub(device.LastSeen) < 10*time.Minute {
			activeDevices++
		}

		if device.CurrentCount < p.Threshold {
			belowThreshold++
		} else {
			aboveThreshold++
		}
	}

	log.Printf("Stats - Total devices: %d, Active: %d, Below threshold: %d, Above: %d, Total current count: %d, Grand total count: %d",
		totalDevices, activeDevices, belowThreshold, aboveThreshold, totalCurrentCount, totalAggregateCount)
}

func (p *PlutoServer) StartPeriodicTasks() {
	statsTicker := time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-statsTicker.C:
				p.PrintStats()
			}
		}
	}()
}
