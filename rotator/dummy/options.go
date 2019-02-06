package dummy

import "github.com/dh1tw/remoteRotator/rotator"

// Name is a functional option to set the name of the rotator
func Name(name string) func(*Dummy) {
	return func(r *Dummy) {
		r.name = name
	}
}

// HasAzimuth is a functional option to enable Azimuth
func HasAzimuth(set bool) func(*Dummy) {
	return func(r *Dummy) {
		r.hasAzimuth = set
	}
}

// HasElevation is a functional option to enable Elevation
func HasElevation(set bool) func(*Dummy) {
	return func(r *Dummy) {
		r.hasElevation = set
	}
}

// AzimuthMin is a functional option to set the minimum azimuth angle.
func AzimuthMin(min int) func(*Dummy) {
	return func(r *Dummy) {
		r.azimuthMin = min
	}
}

// AzimuthMax is a functional option to set the maximum azimuth angle.
func AzimuthMax(max int) func(*Dummy) {
	return func(r *Dummy) {
		r.azimuthMax = max
	}
}

// AzimuthStop is a functional option to set the mechanical stop of the rotator.
func AzimuthStop(stop int) func(*Dummy) {
	return func(r *Dummy) {
		r.azimuthStop = stop
	}
}

// AzimuthSpeed sets the simulated speed of the rotator in degrees / second
func AzimuthSpeed(speed int) func(*Dummy) {
	return func(r *Dummy) {
		r.azSpeed = float32(speed)
	}
}

// ElevationMin is a functional option to set the minimum elevation angle.
func ElevationMin(min int) func(*Dummy) {
	return func(r *Dummy) {
		r.elevationMin = min
	}
}

// ElevationMax is a functional option to set the maximum elevation angle.
func ElevationMax(max int) func(*Dummy) {
	return func(r *Dummy) {
		r.elevationMax = max
	}
}

// ElevationSpeed sets the simulated speed of the rotator in degrees / second
func ElevationSpeed(speed int) func(*Dummy) {
	return func(r *Dummy) {
		r.elSpeed = float32(speed)
	}
}

// EventHandler sets a callback function through which the rotator
// will report Event
func EventHandler(h func(rotator.Rotator, rotator.Heading)) func(*Dummy) {
	return func(r *Dummy) {
		r.eventHandler = h
	}
}
