package config

import (
	"time"
)

// Duration is deserializable wrapper around time.Duration
type Duration struct {
	time.Duration
}

func NewDuration(d time.Duration) Duration {
	return Duration{Duration: d}
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(src []byte) (err error) {
	d.Duration, err = time.ParseDuration(string(src))
	return err
}
