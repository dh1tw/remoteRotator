package rotator

//go:generate stringer -type=Event

// Event represents events which are emitted from the rotator
type Event int

const (
	Azimuth   Event = iota // int
	Elevation              // int
)

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
	Serialize() Status
}

// Status contains the current information from a rotator. The struct
// can be converted into a JSON object.
type Status struct {
	Name         string `json:"name,omitempty"`
	HasAzimuth   bool   `json:"has_azimuth,omitempty"`
	HasElevation bool   `json:"has_elevation,omitempty"`
	Azimuth      int    `json:"azimuth,omitempty"`
	AzPreset     int    `json:"az_preset,omitempty"`
	Elevation    int    `json:"elevation,omitempty"`
	ElPreset     int    `json:"el_preset,omitempty"`
}
