package gowebsocket

import "time"

type Config struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

func newConfig() *Config {
	return &Config{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     (60 * time.Second * 9) / 10,
		MaxMessageSize: 512,
	}
}
