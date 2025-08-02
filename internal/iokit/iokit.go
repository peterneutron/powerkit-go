//go:build darwin
// +build darwin

// Package iokit provides internal access to IOKit for both polling
// and streaming power source data.
package iokit

import "sync"

// RawData holds the unprocessed data returned from the IOKit C functions.
// This struct is intended for internal use within the library.
type RawData struct {
	CurrentCharge   int
	IsCharging      bool
	IsConnected     bool
	IsFullyCharged  bool
	CycleCount      int
	DesignCapacity  int
	MaxCapacity     int
	NominalCapacity int
	CurrentCapacity int
	TimeToEmpty     int
	TimeToFull      int
	Temperature     int
	Voltage         int
	Amperage        int
	SerialNumber    string
	DeviceName      string
	AdapterWatts    int
	AdapterVoltage  int
	AdapterAmperage int
	AdapterDesc     string
	SourceVoltage   int
	SourceAmperage  int
	CellVoltages    []int
}

// --- Streaming Globals ---

// Updates is a channel that receives a signal when IOKit properties change.
var Updates = make(chan struct{}, 1)

var startOnce sync.Once
