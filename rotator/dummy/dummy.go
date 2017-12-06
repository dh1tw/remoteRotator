package dummy

import (
	"math"
	"sync"
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
)

// Dummy is the implementation of a Dummy rotator which can be used
// for testing purposes
type Dummy struct {
	sync.RWMutex
	eventHandler   func(rotator.Rotator, rotator.Event, ...interface{})
	name           string
	azimuthMin     int
	azimuthMax     int
	azimuthStop    int
	azimuthOverlap bool
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
// azimuthMax: 360,
// elevationMax: 180,
// azSpeed: 8, (deg/sec)
// elSpeed: 5, (deg/sec)
func NewDummyRotator(options ...func(*Dummy)) (*Dummy, error) {

	r := &Dummy{
		hasAzimuth:     true,
		azimuthMax:     360,
		elevationMax:   180,
		azSpeed:        8,
		elSpeed:        5,
		tickerInterval: 100,
	}

	for _, opt := range options {
		opt(r)
	}

	if r.azimuthMin > 0 {
		r.azPreset = float32(r.azimuthMin)
		r.azimuth = float32(r.azimuthMin)
	}

	if r.elevationMin > 0 {
		r.elPreset = float32(r.elevationMin)
		r.elevation = float32(r.elevationMin)
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
	r.RLock()
	defer r.RUnlock()
	return r.name
}

// HasAzimuth returns a boolean value indicating if this rotator supports
// horizontal rotation
func (r *Dummy) HasAzimuth() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasAzimuth
}

// HasElevation returns a boolean value indicating if this rotator supports
// vertical rotation
func (r *Dummy) HasElevation() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasElevation
}

// Azimuth returns the current horizontal heading of the rotator in degrees
func (r *Dummy) Azimuth() int {
	r.RLock()
	defer r.RUnlock()
	return int(r.azimuth)
}

// AzPreset returns the horizontal heading (preset) to which the rotator
// shall turn to
func (r *Dummy) AzPreset() int {
	r.RLock()
	defer r.RUnlock()
	return int(r.azPreset)
}

// SetAzimuth sets to value of the horizontal heading to which the
// rotator shall turn to. Allowed values are 0 ... 450. Values outside
// of this range will be clipped.
func (r *Dummy) SetAzimuth(az int) error {
	r.Lock()
	defer r.Unlock()

	if !r.hasAzimuth {
		return nil
	}

	if az > r.azimuthMax {
		az = r.azimuthMax
	}

	if az < r.azimuthMin {
		az = r.azimuthMin
	}

	abs := math.Abs(float64(r.azimuthMax - r.azimuthMin))

	// if rotation is only allowed on less than 360°
	if abs < 360 {

		// special case: overlapping 0°
		if r.azimuthMin > r.azimuthMax {
			if az >= r.azimuthMin || az <= r.azimuthMax {
				r.azPreset = float32(az)
				return nil
			}

			if math.Abs(float64(az-r.azimuthMin)) < math.Abs(float64(az-r.azimuthMax)) {
				r.azPreset = float32(r.azimuthMin)
			} else {
				r.azPreset = float32(r.azimuthMax)
			}
			return nil
		}

		if az >= r.azimuthMin && az <= r.azimuthMax {
			r.azPreset = float32(az)
			return nil
		}

		if math.Abs(float64(az-r.azimuthMin)) < math.Abs(float64(az-r.azimuthMax)) {
			r.azPreset = float32(r.azimuthMin)
		} else {
			r.azPreset = float32(r.azimuthMax)
		}
		return nil
	}

	r.azPreset = float32(az)
	return nil
}

// Elevation returns the current vertical elevation of the rotator in degrees
func (r *Dummy) Elevation() int {
	r.RLock()
	defer r.RUnlock()
	return int(r.elevation)
}

// ElPreset returns the vertical elevation (preset) to which the rotator
// shall turn to
func (r *Dummy) ElPreset() int {
	r.RLock()
	defer r.RUnlock()
	return int(r.elPreset)
}

// SetElevation sets to value of the vertical elevation to which the
// rotator shall turn to. Allowed values are 0 ... 180. Values outside
// of this range will be clipped.
func (r *Dummy) SetElevation(el int) error {
	r.Lock()
	defer r.Unlock()

	if !r.hasElevation {
		return nil
	}

	if el > 180 {
		el = 180
	}

	if el < 0 {
		el = 0
	}

	if el < r.elevationMin {
		r.elPreset = float32(r.elevationMin)
	} else if el > r.elevationMax {
		r.elPreset = float32(r.elevationMax)
	} else {
		r.elPreset = float32(el)
	}

	return nil
}

// StopAzimuth stops horizontal rotator movement
func (r *Dummy) StopAzimuth() error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = r.azimuth
	if r.eventHandler != nil {
		r.eventHandler(r, rotator.Azimuth, r.status())
	}

	return nil
}

// StopElevation stops vertical rotator movement
func (r *Dummy) StopElevation() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation
	if r.eventHandler != nil {
		r.eventHandler(r, rotator.Elevation, r.status())
	}
	return nil
}

// Stop stops all rotator movement
func (r *Dummy) Stop() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation
	r.azPreset = r.azimuth
	if r.eventHandler != nil {
		status := r.status()
		r.eventHandler(r, rotator.Azimuth, status)
		r.eventHandler(r, rotator.Elevation, status)
	}

	return nil
}

func (r *Dummy) status() rotator.Status {
	return rotator.Status{
		Name:           r.name,
		Azimuth:        int(r.azimuth),
		AzPreset:       int(r.azPreset),
		AzimuthOverlap: r.azimuthOverlap,
		Elevation:      int(r.elevation),
		ElPreset:       int(r.elPreset),
	}
}

// Status returns a a rotator.Status struct with the information
// of this rotator.
func (r *Dummy) Status() rotator.Status {
	r.RLock()
	defer r.RUnlock()
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
	r.RLock()
	defer r.RUnlock()

	return rotator.Info{
		Name:           r.name,
		HasAzimuth:     r.hasAzimuth,
		HasElevation:   r.hasElevation,
		AzimuthMin:     r.azimuthMin,
		AzimuthMax:     r.azimuthMax,
		AzimuthStop:    r.azimuthStop,
		AzimuthOverlap: r.azimuthOverlap,
		ElevationMin:   r.elevationMin,
		ElevationMax:   r.elevationMax,
		Azimuth:        int(r.azimuth),
		AzPreset:       int(r.azPreset),
		Elevation:      int(r.elevation),
		ElPreset:       int(r.elPreset),
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
		changed := r.calcNewAzHeading()
		if changed {
			r.eventHandler(r, rotator.Azimuth, r.status())
		}
	}
}

func (r *Dummy) updateElevation() {

	if r.hasElevation {
		changed := r.calcNewElHeading()
		if changed {
			r.eventHandler(r, rotator.Elevation, r.status())
		}
	}
}

func (r *Dummy) calcNewElHeading() bool {

	if int(r.elevation) == int(r.elPreset) {
		return false
	}

	moveCCW := false
	moveCW := false

	delta := r.elSpeed / (r.tickerInterval / 10)

	min := float32(r.elevationMin)
	max := float32(r.elevationMax)

	if r.elPreset < r.elevation && r.elevation > min {
		moveCCW = true
	} else if r.elPreset > r.elevation && r.elevation < max {
		moveCW = true
	}

	if moveCW {
		r.elevation += delta
	}

	if moveCCW {
		r.elevation -= delta
	}

	return true
}

func (r *Dummy) calcNewAzHeading() bool {

	if int(r.azimuth) == int(r.azPreset) {
		return false
	}

	moveCCW := false
	moveCW := false

	delta := r.azSpeed / (r.tickerInterval / 10)

	abs := math.Abs(float64(r.azimuthMax - r.azimuthMin))

	// if rotation is allowed on < 360°
	if abs < 360 {
		min := float32(r.azimuthMin)
		max := float32(r.azimuthMax)

		// max overlapping 0°
		if min >= max {
			if (r.azimuth >= min && r.azPreset < 360 && r.azPreset > r.azimuth) ||
				(r.azimuth >= min && r.azPreset <= max) ||
				(r.azimuth < max && r.azPreset > r.azimuth && r.azPreset <= max) {
				moveCW = true
			} else {
				moveCCW = true
			}
		} else {
			if r.azPreset > r.azimuth {
				moveCW = true
			} else {
				moveCCW = true
			}
		}
	} else { // rotation allowed on >= 360°
		// moving CW or CCW?
		if r.azPreset-r.azimuth > 0 {
			if (float32(r.azimuthStop) < r.azPreset) && (float32(r.azimuthStop) > r.azimuth) {
				moveCCW = true
			} else {
				moveCW = true
			}
		} else {
			//crossing mechanical stop
			if (float32(r.azimuthStop) > r.azPreset) && (float32(r.azimuthStop) < r.azimuth) {
				moveCW = true
			} else {
				moveCCW = true
			}
		}
	}
	// } else { // rotation allowed on >= 360°
	// 	azMax := r.azimuthMax - 360 + r.azimuthStop

	// 	// overlap not crossing 0°
	// 	if r.azimuthStop <= azMax {

	// 		if r.azPreset > r.azimuth {
	// 			fmt.Println("1")
	// 			if r.azPreset < float32(azMax) && r.azimuth < float32(azMax) {
	// 				fmt.Println("2")
	// 				moveCW = true
	// 				if int(r.azimuth) == r.azimuthStop {
	// 					r.azimuthOverlap = true
	// 					fmt.Println("3")
	// 				}
	// 			} else if r.azPreset > float32(azMax) && r.azimuth < float32(azMax) {
	// 				fmt.Println("4")
	// 				moveCCW = true
	// 				if int(r.azimuth) > r.azimuthStop && !r.azimuthOverlap {
	// 					moveCW = true
	// 					// r.azimuthOverlap = true
	// 					fmt.Println("5")
	// 				} else if int(r.azimuth) > r.azimuthStop && r.azimuthOverlap {
	// 					fmt.Println(55)
	// 				} else {
	// 					r.azimuthOverlap = false
	// 					fmt.Println("6")
	// 				}
	// 			} else if r.azPreset > float32(azMax) && r.azimuth > float32(azMax) {
	// 				fmt.Println("7")
	// 				moveCCW = true
	// 			} else {
	// 				fmt.Println(100)
	// 			}
	// 		} else {
	// 			if r.azPreset < float32(azMax) && r.azimuth < float32(azMax) {
	// 				fmt.Println(10)
	// 				// if int(r.azimuth) > r.azimuthStop && !r.azimuthOverlap {
	// 				// 	fmt.Println(11)
	// 				// 	moveCW = true
	// 				if int(r.azimuth) == r.azimuthStop {
	// 					r.azimuthOverlap = false
	// 				}
	// 				if int(r.azimuth) > r.azimuthStop {
	// 					moveCCW = true
	// 					fmt.Println(12)
	// 					// } else if int(r.azimuth) < r.azimuthStop {
	// 					// 	moveCCW = true
	// 					// 	fmt.Println(122)
	// 				} else {
	// 					// r.azimuthOverlap = false
	// 					fmt.Println(13)
	// 					moveCCW = true
	// 				}
	// 			} else if r.azPreset > float32(r.azimuthStop) && r.azimuth > float32(r.azimuthStop) {
	// 				fmt.Println(14)
	// 				moveCCW = true
	// 			} else if r.azPreset < float32(r.azimuthStop) && r.azimuth < float32(azMax) {
	// 				moveCW = true
	// 			} else {
	// 				fmt.Println(200)
	// 			}
	// 			// moveCCW = true
	// 			// fmt.Println("10")
	// 		}
	// 		// overlap crossing 0°
	// 	} else if r.azimuthStop < azMax {

	// 	}
	// }

	if moveCW {
		// crossing from 359 -> 0
		if r.azimuth <= 359 && r.azimuth+delta > 359 {
			r.azimuth = delta
			// fmt.Printf("%.2f\n", r.azimuth)
			return true
		}
		r.azimuth += delta
		// fmt.Printf("%.2f\n", r.azimuth)

	} else if moveCCW {
		// crossing from 0 <- 360
		if r.azimuth >= 0 && r.azimuth-delta < 0 {
			r.azimuth = 359 - delta
			// fmt.Printf("%.2f\n", r.azimuth)
			return true
		}
		r.azimuth -= delta
		// fmt.Printf("%.2f\n", r.azimuth)
	}

	return true
}
