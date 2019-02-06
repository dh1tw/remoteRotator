package rotator

// Rotator is the interface which has to be implemented by each Rotator
type Rotator interface {
	Name() string
	HasAzimuth() bool
	HasElevation() bool
	Azimuth() int
	AzPreset() int
	SetAzimuth(az int) error
	Elevation() int
	ElPreset() int
	SetElevation(el int) error
	StopAzimuth() error
	StopElevation() error
	Stop() error
	Serialize() Object
	Close()
}

// EventHandler is called whenever a variable of a rotator changes
type EventHandler func(Rotator, Heading)
