package rotator

//go:generate stringer -type=Event

// Event represents events which are emitted from the rotator
type Event int

// Events send from a rotator
const (
	Azimuth   Event = iota // sending as value Status{}
	Elevation              // sending as value Status{}
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
	Status() Status
	ExecuteRequest(Request) error
	Info() Info
}

// Status contains the current information from a rotator. The struct
// can be converted into a JSON object.
type Status struct {
	Name           string `json:"name"`
	Azimuth        int    `json:"azimuth"`
	AzPreset       int    `json:"az_preset"`
	AzimuthOverlap bool   `json:"az_overlap"`
	Elevation      int    `json:"elevation"`
	ElPreset       int    `json:"el_preset"`
}

// Request contains the fields to control a rotator. This message is
// typically sent from a client (e.g. via a websocket) to the rotator.
type Request struct {
	Name          string `json:"name,omitempty"`
	HasAzimuth    bool   `json:"has_azimuth,omitempty"`
	HasElevation  bool   `json:"has_elevation,omitempty"`
	Azimuth       int    `json:"azimuth,omitempty"`
	Elevation     int    `json:"elevation,omitempty"`
	StopAzimuth   bool   `json:"stop_azimuth,omitempty"`
	StopElevation bool   `json:"stop_elevation,omitempty"`
	Stop          bool   `json:"stop,omitempty"`
}

// Info exports all the attributes of a rotator.
type Info struct {
	Name           string `json:"name"`
	HasAzimuth     bool   `json:"has_azimuth"`
	HasElevation   bool   `json:"has_elevation"`
	AzimuthMin     int    `json:"az_min"`
	AzimuthMax     int    `json:"az_max"`
	AzimuthStop    int    `json:"az_stop"`
	AzimuthOverlap bool   `json:"az_overlap"`
	ElevationMin   int    `json:"el_min,omitempty"`
	ElevationMax   int    `json:"el_max,omitempty"`
	Azimuth        int    `json:"azimuth"`
	AzPreset       int    `json:"az_preset"`
	Elevation      int    `json:"elevation"`
	ElPreset       int    `json:"el_preset"`
}
