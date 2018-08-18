package rotator

type AzimuthGet struct {
	HasAzimuth bool `json:"has_azimuth"`
	Azimuth    int  `json:"azimuth"`
	Preset     int  `json:"preset"`
}

type AzimuthPut struct {
	Azimuth *int `json:"azimuth"`
}

type ElevationGet struct {
	HasElevation bool `json:"has_elevation"`
	Elevation    int  `json:"elevation"`
	Preset       int  `json:"preset"`
}

type ElevationPut struct {
	Elevation *int `json:"elevation"`
}

type Object struct {
	Name    string  `json:"name"`
	Heading Heading `json:"heading"`
	Config  Config  `json:"config"`
}

type Heading struct {
	Azimuth   int `json:"azimuth"`
	AzPreset  int `json:"az_preset"`
	Elevation int `json:"elevation"`
	ElPreset  int `json:"el_preset"`
}

type Objects map[string]Object

type Config struct {
	HasAzimuth   bool `json:"has_azimuth"`
	AzimuthMin   int  `json:"azimuth_min"`
	AzimuthMax   int  `json:"azimuth_max"`
	AzimuthStop  int  `json:"azimuth_stop"`
	HasElevation bool `json:"has_elevation"`
	ElevationMin int  `json:"elevation_min"`
	ElevationMax int  `json:"elevation_max"`
}
