package core

import (
	"database/sql"
	"net"
	"time"
)

const PlutoDBPassword = "a_very_secret_pluto_password_!@#"

type Device struct {
	IP           string    // IP address of a device
	CurrentCount int       // Total trigger count after a maintenance operation
	TotalCount   int       // Total trigger count after service deployment (doesn't reset after maintenance)
	LastSeen     time.Time // The last timestamp for a device be seen as online
	RegisteredAt time.Time // First registration timestamp of a device to this service
}

type PlutoServer struct {
	Db        *sql.DB
	Devices   map[string]*Device
	Conn      *net.UDPConn
	Threshold int // After the trigger count of a device exceeds a certain Threshold value, it must go to maintenance
}
