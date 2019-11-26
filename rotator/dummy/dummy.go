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
	eventHandler   func(rotator.Rotator, rotator.Heading)
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
	closer         sync.Once
}

// New creates a new dummy rotator which satisfies the
// rotator.Rotator interface. Options can be injected through functional
// options. If the Dummy can not be initialized, nil and the corresponding error
// will be returned.
// Default settings are:
// hasAzimuth: true,
// azimuthMax: 360,
// elevationMax: 180,
// azSpeed: 8, (deg/sec)
// elSpeed: 5, (deg/sec)
func New(options ...func(*Dummy)) (*Dummy, error) {

	r := &Dummy{
		hasAzimuth:     true,
		azimuthMax:     360,
		elevationMax:   180,
		azSpeed:        8,
		elSpeed:        5,
		tickerInterval: 100,
		closeCh:        make(chan struct{}),
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

	r.ticker = time.NewTicker(time.Millisecond * time.Duration(r.tickerInterval))

	go r.start()

	return r, nil
}

// // start the event loop
func (r *Dummy) start() {

	r.ticker = time.NewTicker(time.Millisecond * time.Duration(r.tickerInterval))
	defer r.ticker.Stop()

	for {
		select {
		case <-r.ticker.C:
			r.updateHeadings()
		case <-r.closeCh:
			return
		}
	}
}

// Close shuts down the rotator and prepares it for garbage collection
func (r *Dummy) Close() {
	r.closer.Do(func() {
		close(r.closeCh)
	})
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
		heading := r.serialize().Heading
		go r.eventHandler(r, heading)
	}

	return nil
}

// StopElevation stops vertical rotator movement
func (r *Dummy) StopElevation() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation
	if r.eventHandler != nil {
		heading := r.serialize().Heading
		go r.eventHandler(r, heading)
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
		heading := r.serialize().Heading
		go r.eventHandler(r, heading)
	}

	return nil
}

// Serialize the data of the rotator
func (r *Dummy) Serialize() rotator.Object {
	r.RLock()
	defer r.RUnlock()
	return r.serialize()
}

func (r *Dummy) serialize() rotator.Object {

	obj := rotator.Object{
		Name: r.name,
		Heading: rotator.Heading{
			Azimuth:   int(r.azimuth),
			AzPreset:  int(r.azPreset),
			Elevation: int(r.elevation),
			ElPreset:  int(r.elPreset),
		},
		Config: rotator.Config{
			HasAzimuth:   r.hasAzimuth,
			HasElevation: r.hasElevation,
			AzimuthMax:   r.azimuthMax,
			AzimuthMin:   r.azimuthMin,
			AzimuthStop:  r.azimuthStop,
			ElevationMax: r.elevationMax,
			ElevationMin: r.elevationMin,
		},
	}

	return obj
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
			heading := r.serialize().Heading
			go r.eventHandler(r, heading)
		}
	}
}

func (r *Dummy) updateElevation() {

	if r.hasElevation {
		changed := r.calcNewElHeading()
		if changed {
			heading := r.serialize().Heading
			go r.eventHandler(r, heading)
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
