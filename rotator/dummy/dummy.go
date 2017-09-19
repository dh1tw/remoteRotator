package dummy

import (
	"sync"
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
)

// Dummy is the implementation of a Dummy rotator which can be used
// for testing purposes
type Dummy struct {
	sync.Mutex
	eventHandler   func(rotator.Rotator, rotator.Event, ...interface{})
	name           string
	description    string
	azimuthMin     int
	azimuthMax     int
	azimuthStop    int
	elevationMin   int
	elevationMax   int
	azimuth        float32
	azPreset       float32
	elevation      float32
	elPreset       float32
	hasAzimuth     bool
	hasElevation   bool
	azSpeed        float32
	elSpeed        float32
	ticker         *time.Ticker
	tickerInterval float32 //ms
	closeCh        chan struct{}
	starter        sync.Once
}

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

// AzimuthSpeed sets the simulated speed of the rotator in degrees / second
func AzimuthSpeed(speed int) func(*Dummy) {
	return func(r *Dummy) {
		r.azSpeed = float32(speed)
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
func EventHandler(h func(rotator.Rotator, rotator.Event, ...interface{})) func(*Dummy) {
	return func(r *Dummy) {
		r.eventHandler = h
	}
}

// NewDummyRotator creates a new dummy rotator which satisfies the
// rotator.Rotator interface. Options can be injected through functional
// options. If the Dummy can not be initialized, nil and the corresponding error
// will be returned.
// Default settings are:
// hasAzimuth: true,
// azimuthMax: 450,
// elevationMax: 180,
// azSpeed: 8, (deg/sec)
// elSpeed: 5, (deg/sec)
func NewDummyRotator(options ...func(*Dummy)) (*Dummy, error) {

	r := &Dummy{
		hasAzimuth:     true,
		azimuthMax:     450,
		elevationMax:   180,
		azSpeed:        8,
		elSpeed:        5,
		tickerInterval: 100,
	}

	for _, opt := range options {
		opt(r)
	}

	return r, nil
}

// Start starts the main event loop for the dummy rotator. The event loop can be
// shutdown by closing the shutdown channel.
func (r *Dummy) Start(shutdown <-chan struct{}) {
	// ensure that the event loop is only started once
	r.starter.Do(func() {
		go r.start(shutdown)
	})
}

// start the event loop
func (r *Dummy) start(shutdown <-chan struct{}) {

	r.ticker = time.NewTicker(time.Millisecond * time.Duration(r.tickerInterval))
	defer r.ticker.Stop()

	for {
		select {
		case <-r.ticker.C:
			r.updateHeadings()
		case <-shutdown:
			return
		}
	}
}

// Name returns the name of the rotator
func (r *Dummy) Name() string {
	r.Lock()
	defer r.Unlock()
	return r.name
}

// HasAzimuth returns a boolean value indicating if this rotator supports
// horizontal rotation
func (r *Dummy) HasAzimuth() bool {
	r.Lock()
	defer r.Unlock()
	return r.hasAzimuth
}

// HasElevation returns a boolean value indicating if this rotator supports
// vertical rotation
func (r *Dummy) HasElevation() bool {
	r.Lock()
	defer r.Unlock()
	return r.hasElevation
}

// Azimuth returns the current horizontal heading of the rotator in degrees
func (r *Dummy) Azimuth() int {
	r.Lock()
	defer r.Unlock()
	return int(r.azimuth)
}

// AzPreset returns the horizontal heading (preset) to which the rotator
// shall turn to
func (r *Dummy) AzPreset() int {
	r.Lock()
	defer r.Unlock()
	return int(r.azPreset)
}

// SetAzimuth sets to value of the horizontal heading to which the
// rotator shall turn to. Allowed values are 0 ... 450. Values outside
// of this range will be clipped.
func (r *Dummy) SetAzimuth(az int) error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = float32(az)
	return nil
}

// Elevation returns the current vertical elevation of the rotator in degrees
func (r *Dummy) Elevation() int {
	r.Lock()
	defer r.Unlock()
	return int(r.elevation)
}

// ElPreset returns the vertical elevation (preset) to which the rotator
// shall turn to
func (r *Dummy) ElPreset() int {
	r.Lock()
	defer r.Unlock()
	return int(r.elPreset)
}

// SetElevation sets to value of the vertical elevation to which the
// rotator shall turn to. Allowed values are 0 ... 180. Values outside
// of this range will be clipped.
func (r *Dummy) SetElevation(el int) error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = float32(el)
	return nil
}

// StopAzimuth stops horizontal rotator movement
func (r *Dummy) StopAzimuth() error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = r.azimuth

	return nil
}

// StopElevation stops vertical rotator movement
func (r *Dummy) StopElevation() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation
	return nil
}

// Stop stops all rotator movement
func (r *Dummy) Stop() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation
	r.azPreset = r.azimuth

	return nil
}

func (r *Dummy) status() rotator.Status {
	return rotator.Status{
		Name:      r.name,
		Azimuth:   int(r.azimuth),
		AzPreset:  int(r.azPreset),
		Elevation: int(r.elevation),
		ElPreset:  int(r.elPreset),
	}
}

// Status returns a a rotator.Status struct with the information
// of this rotator.
func (r *Dummy) Status() rotator.Status {
	r.Lock()
	defer r.Unlock()
	return r.status()
}

// ExecuteRequest takes a request struct and sets the new values
func (r *Dummy) ExecuteRequest(req rotator.Request) error {
	if req.HasAzimuth {
		if err := r.SetAzimuth(req.Azimuth); err != nil {
			return err
		}
	}

	if req.HasElevation {
		if err := r.SetElevation(req.Elevation); err != nil {
			return err
		}
	}

	if req.StopAzimuth {
		if err := r.StopAzimuth(); err != nil {
			return err
		}
	}

	if req.StopElevation {
		if err := r.StopElevation(); err != nil {
			return err
		}
	}

	if req.Stop {
		if err := r.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// Info returns a rotator.Info struct with the current values of the rotator
func (r *Dummy) Info() rotator.Info {
	r.Lock()
	defer r.Unlock()

	return rotator.Info{
		Name:         r.name,
		Description:  r.description,
		HasAzimuth:   r.hasAzimuth,
		HasElevation: r.hasElevation,
		AzimuthMin:   r.azimuthMin,
		AzimuthMax:   r.azimuthMax,
		AzimuthStop:  r.azimuthStop,
		ElevationMin: r.elevationMin,
		ElevationMax: r.elevationMax,
	}
}

func (r *Dummy) updateHeadings() {
	r.Lock()
	defer r.Unlock()
	r.updateAzimuth()
	r.updateElevation()
}

func (r *Dummy) updateAzimuth() {

	if r.hasAzimuth {
		v, changed := calcNewHeading(r.azimuth, r.azPreset, r.azSpeed, r.tickerInterval)
		if changed {
			r.azimuth = v
			r.eventHandler(r, rotator.Azimuth, r.status())
		}
	}
}

func (r *Dummy) updateElevation() {

	if r.hasElevation {
		v, changed := calcNewHeading(r.elevation, r.elPreset, r.elSpeed, r.tickerInterval)
		if changed {
			r.azimuth = v
			r.eventHandler(r, rotator.Azimuth, r.status())
		}
	}
}

func calcNewHeading(position, preset, speed, interval float32) (float32, bool) {

	if int(position) == int(preset) {
		return position, false
	}

	if preset > position {
		return position + speed/(interval/10), true
	}
	return position - speed/(interval/10), true
}
