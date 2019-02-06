package yaesu

import (
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
)

// Name is a functional option to set the name of the rotator
func Name(name string) func(*Yaesu) {
	return func(r *Yaesu) {
		r.name = name
	}
}

// HasAzimuth is a functional option to enable Azimuth
func HasAzimuth(set bool) func(*Yaesu) {
	return func(r *Yaesu) {
		r.hasAzimuth = set
	}
}

// HasElevation is a functional option to enable Elevation
func HasElevation(set bool) func(*Yaesu) {
	return func(r *Yaesu) {
		r.hasElevation = set
	}
}

// UpdateInterval is a functional option the set the frequency
// by which the rotator will be queried
func UpdateInterval(d time.Duration) func(*Yaesu) {
	return func(r *Yaesu) {
		r.pollingInterval = d
	}
}

// EventHandler sets a callback function through which the rotator
// will report Event
func EventHandler(h func(rotator.Rotator, rotator.Heading)) func(*Yaesu) {
	return func(r *Yaesu) {
		r.eventHandler = h
	}
}

// Baudrate is a functional option to set the baurate of the serial port.
func Baudrate(baudrate int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.spBaudrate = baudrate
	}
}

// Portname is a functional option to set the portname of the serial port.
// On Windows this will be "COMx", on Linux & MacOS "/dev/tty/xxx"
func Portname(pn string) func(*Yaesu) {
	return func(r *Yaesu) {
		r.spPortName = pn
	}
}

// AzimuthMin is a functional option to set the minimum azimuth angle.
func AzimuthMin(min int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.azimuthMin = min
	}
}

// AzimuthMax is a functional option to set the maximum azimuth angle.
func AzimuthMax(max int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.azimuthMax = max
	}
}

// AzimuthStop is a functional option to set the mechanical stop of the rotator.
func AzimuthStop(stop int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.azimuthStop = stop
	}
}

// ElevationMin is a functional option to set the minimum elevation angle.
func ElevationMin(min int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.elevationMin = min
	}
}

// ElevationMax is a functional option to set the maximum elevation angle.
func ElevationMax(max int) func(*Yaesu) {
	return func(r *Yaesu) {
		r.elevationMax = max
	}
}

// ErrorCh is a functional option allows you to pass a channel to the rotator.
// The channel will be closed when an internal error occures.
func ErrorCh(ch chan struct{}) func(*Yaesu) {
	return func(r *Yaesu) {
		r.errorCh = ch
	}
}
