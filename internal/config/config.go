package config

import "time"

var (
	jwtSecret []byte
)

const (
	TokenValidityPeriod = time.Hour * 24
	IpcSockPath         = "/tmp/metrics.sock"
)
