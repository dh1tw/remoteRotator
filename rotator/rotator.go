package rotator

//go:generate stringer -type=Event

// Event represents events which are emitted from the rotator
type Event int

// Events send from a rotator
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
	ExecuteRequest(Request) error
	Info() Info
}

// Status contains the current information from a rotator. The struct
// can be converted into a JSON object.
type Status struct {
	Name      string `json:"name"`
	Azimuth   int    `json:"azimuth"`
	AzPreset  int    `json:"az_preset"`
	Elevation int    `json:"elevation"`
	ElPreset  int    `json:"el_preset"`
}

// Request contains the fields to control a rotator.
type Request struct {
	HasAzimuth    bool `json:"has_azimuth,omitempty"`
	HasElevation  bool `json:"has_elevation,omitempty"`
	Azimuth       int  `json:"azimuth,omitempty"`
	Elevation     int  `json:"elevation,omitempty"`
	StopAzimuth   bool `json:"stop_azimuth,omitempty"`
	StopElevation bool `json:"stop_elevation,omitempty"`
	Stop          bool `json:"stop,omitempty"`
}

// Info contains the meta data of a rotator
type Info struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	HasAzimuth   bool   `json:"has_azimuth,omitempty"`
	HasElevation bool   `json:"has_elevation,omitempty"`
	AzimuthMin   int    `json:"azimuth_min,omitempty"`
	AzimuthMax   int    `json:"azimuth_max,omitempty"`
	AzimuthStop  int    `json:"azimuth_stop,omitempty"`
	ElevationMin int    `json:"elevation_min,omitempty"`
	ElevationMax int    `json:"elevation_max,omitempty"`
}
