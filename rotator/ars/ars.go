package ars

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"
	"time"

	serial "github.com/tarm/serial"

	"github.com/dh1tw/remoteRotator/rotator"
)

// Ars is the implementation of the Rotator interface for EA4TX's
// Antenna Rotator System (ARS)
type Ars struct {
	sync.Mutex
	name            string
	description     string
	azimuthMin      int
	azimuthMax      int
	azimuthStop     int
	elevationMin    int
	elevationMax    int
	azimuth         int
	azPreset        int
	elevation       int
	elPreset        int
	hasAzimuth      bool
	hasElevation    bool
	pollingInterval time.Duration
	pollingTicker   *time.Ticker
	eventHandler    func(rotator.Rotator, rotator.Event, ...interface{})
	sp              io.ReadWriteCloser
	spPortName      string
	spBaudrate      int
	spCloser        sync.Once
	headingPattern  *regexp.Regexp
	watchdogTs      time.Time
}

// Name is a functional option to set the name of the rotator
func Name(name string) func(*Ars) {
	return func(r *Ars) {
		r.name = name
	}
}

// HasAzimuth is a functional option to enable Azimuth
func HasAzimuth(set bool) func(*Ars) {
	return func(r *Ars) {
		r.hasAzimuth = set
	}
}

// HasElevation is a functional option to enable Elevation
func HasElevation(set bool) func(*Ars) {
	return func(r *Ars) {
		r.hasElevation = set
	}
}

// UpdateInterval is a functional option the set the frequency
// by which the rotator will be queried
func UpdateInterval(d time.Duration) func(*Ars) {
	return func(r *Ars) {
		r.pollingInterval = d
	}
}

// EventHandler sets a callback function through which the rotator
// will report Event
func EventHandler(h func(rotator.Rotator, rotator.Event, ...interface{})) func(*Ars) {
	return func(r *Ars) {
		r.eventHandler = h
	}
}

// Baudrate is a functional option to set the baurate of the serial port.
func Baudrate(baudrate int) func(*Ars) {
	return func(r *Ars) {
		r.spBaudrate = baudrate
	}
}

// Portname is a functional option to set the portname of the serial port.
// On Windows this will be "COMx", on Linux & MacOS "/dev/tty/xxx"
func Portname(pn string) func(*Ars) {
	return func(r *Ars) {
		r.spPortName = pn
	}
}

// AzimuthMin is a functional option to set the minimum azimuth angle.
func AzimuthMin(min int) func(*Ars) {
	return func(r *Ars) {
		r.azimuthMin = min
	}
}

// AzimuthMax is a functional option to set the maximum azimuth angle.
func AzimuthMax(max int) func(*Ars) {
	return func(r *Ars) {
		r.azimuthMax = max
	}
}

// AzimuthStop is a functional option to set the mechanical stop of the rotator.
func AzimuthStop(stop int) func(*Ars) {
	return func(r *Ars) {
		r.azimuthStop = stop
	}
}

// ElevationMin is a functional option to set the minimum elevation angle.
func ElevationMin(min int) func(*Ars) {
	return func(r *Ars) {
		r.elevationMin = min
	}
}

// ElevationMax is a functional option to set the maximum elevation angle.
func ElevationMax(max int) func(*Ars) {
	return func(r *Ars) {
		r.elevationMax = max
	}
}

// NewArs creates a new Ars object which satisfies the rotator.Rotator interface.
// Configuration settings are set through functional options. The the Ars
// can not be initialized nil and the corresponding error will be returned.
// Default settings are:
// hasAzimuth: true,
// portname: /dev/ttyACM0,
// pollingInterval: 5sec,
// baudrate: 9600.
func NewArs(opts ...func(*Ars)) (*Ars, error) {

	// regex Pattern [0-9]{4} -> 0310..etc
	headingPattern, err := regexp.Compile("[\\d]{4}")
	if err != nil {
		return nil, err
	}

	r := &Ars{
		hasAzimuth:      true,
		pollingInterval: time.Second * 5,
		spPortName:      "/dev/ttyACM0",
		spBaudrate:      9600,
		headingPattern:  headingPattern,
		azimuthMax:      450,
		elevationMax:    180,
	}

	for _, opt := range opts {
		opt(r)
	}

	config := &serial.Config{
		Name:        r.spPortName,
		Baud:        r.spBaudrate,
		ReadTimeout: time.Millisecond * 100,
		Parity:      serial.ParityNone,
		Size:        8,
		StopBits:    1,
	}

	sp, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}

	r.sp = sp

	return r, nil
}

// Close terminates the serial port connection. Afterwards the
// object should be destroyed.
func (r *Ars) Close() {
	r.Lock()
	defer r.Unlock()
	if r.pollingTicker != nil {
		r.pollingTicker.Stop()
	}
	// makes sure that the serial port just gets closed once
	r.spCloser.Do(func() { r.sp.Close() })
}

// resetWatchdog resets the watchdog. This means that a packet has been
// received from the ARS
func (r *Ars) resetWatchdog() {
	r.Lock()
	defer r.Unlock()
	r.watchdogTs = time.Now()
}

// checkWatchdog compares the watchdog timestamp with the current time
// and returns true if this value is greater than 5x updateInterval.
func (r *Ars) checkWatchdog() bool {
	r.Lock()
	defer r.Unlock()
	if time.Since(r.watchdogTs) > 5*r.pollingInterval {
		return true
	}
	return false
}

// Start starts the main event loop for the serial port.
// It will query the ARS for the current heading (azimuth + elevation)
// with the pollingrate defined during initialization.
// A watchdog detects if the ARS does not respond anymore. Since this
// function runs in an endless loop, it is typically executed in a
// go routine.
// If an error occures, the communication will be shut down and the
// arsEerror channel will be closed.
func (r *Ars) Start(arsError chan<- bool, shutdown <-chan bool) {
	defer close(arsError)
	defer r.Close()

	r.Lock()
	r.pollingTicker = time.NewTicker(r.pollingInterval)
	r.watchdogTs = time.Now()
	r.Unlock()

	for {
		select {
		case <-r.pollingTicker.C:
			// fmt.Println("tick")
			if err := r.query(); err != nil {
				fmt.Println("serial port write error:", err)
				return
			}
			if r.checkWatchdog() {
				fmt.Println("communication lost with ARS")
				return
			}
		case <-shutdown:
			return
		default:
			// pass
		}

		msg, err := r.read()
		if err != nil {
			// Timeout... continue
			if err == io.EOF {
				continue
			}
			fmt.Printf("serial port read error (%s on %s): %s\n",
				r.name, r.spPortName, err)
			return // exit
		}
		r.resetWatchdog()
		r.parseMsg(msg)
	}
}

// read from the ARS through this wrapper function
func (r *Ars) read() (string, error) {
	r.Lock()
	defer r.Unlock()
	return bufio.NewReader(r.sp).ReadString('\n')
}

// request Azimuth + Elevation from ARS
func (r *Ars) query() error {
	//query azimuth + elevation
	r.Lock()
	defer r.Unlock()
	_, err := r.write([]byte("C2\r\n"))
	return err
}

// all functions write to the ARS / serial port through this wrapper function
func (r *Ars) write(data []byte) (int, error) {
	return r.sp.Write(data)
}

// parseMsg checks the content of the received message from the ARS
// and then further stores them and executes the event callback
func (r *Ars) parseMsg(msg string) {

	headings := []string{}

	if r.headingPattern != nil {
		headings = r.headingPattern.FindAllString(msg, -1)
	}

	if len(headings) > 0 {
		//contains always 4 digits
		az, _ := strconv.Atoi(headings[0][1:]) //discard the first digit, since it's always 0
		r.setValueAndCallEvent(rotator.Azimuth, az)
	}

	if len(headings) == 2 {
		// contains always 4 digits
		el, _ := strconv.Atoi(headings[1][1:])
		r.setValueAndCallEvent(rotator.Elevation, el)
	}
}

// verify if the data has changed. If so, notify the application through
// the callback
func (r *Ars) setValueAndCallEvent(ev rotator.Event, value int) {
	r.Lock()
	defer r.Unlock()

	switch ev {
	case rotator.Azimuth:
		if r.azimuth != value {
			r.azimuth = value
			if r.eventHandler != nil {
				// cb launched async to avoid deadlock
				// on ars.*()
				go r.eventHandler(r, rotator.Azimuth, value)
			}
		}
	case rotator.Elevation:
		if r.elevation != value {
			r.elevation = value
			if r.eventHandler != nil {
				// cb launched async
				go r.eventHandler(r, rotator.Elevation, value)
			}
		}
	}
}

// Name returns the name of the rotator
func (r *Ars) Name() string {
	return r.name
}

// Azimuth returns the current horizontal heading of the rotator in degrees
func (r *Ars) Azimuth() int {
	r.Lock()
	defer r.Unlock()
	return r.azimuth
}

// AzPreset returns the horizontal heading (preset) to which the rotator
// shall turn to
func (r *Ars) AzPreset() int {
	r.Lock()
	defer r.Unlock()
	return r.azPreset
}

// HasAzimuth returns a boolean value indicating if this rotator supports
// horizontal rotation
func (r *Ars) HasAzimuth() bool {
	r.Lock()
	defer r.Unlock()
	return r.hasAzimuth
}

// SetAzimuth sets to value of the horizontal heading to which the
// rotator shall turn to. Allowed values are 0 ... 450. Values outside
// of this range will be clipped.
func (r *Ars) SetAzimuth(az int) error {
	r.Lock()
	defer r.Unlock()

	if !r.hasAzimuth {
		return nil
	}

	if az > 450 {
		az = 450
	}

	if az < 0 {
		az = 0
	}

	r.azPreset = az

	if _, err := r.write([]byte(fmt.Sprintf("M%.3d\r\n", az))); err != nil {
		return err
	}

	return nil
}

// Elevation returns the current vertical elevation of the rotator in degrees
func (r *Ars) Elevation() int {
	r.Lock()
	defer r.Unlock()
	return r.elevation
}

// ElPreset returns the vertical elevation (preset) to which the rotator
// shall turn to
func (r *Ars) ElPreset() int {
	r.Lock()
	defer r.Unlock()
	return r.elPreset
}

// HasElevation returns a boolean value indicating if this rotator supports
// vertical rotation
func (r *Ars) HasElevation() bool {
	r.Lock()
	defer r.Unlock()
	return r.hasElevation
}

// SetElevation sets to value of the vertical elevation to which the
// rotator shall turn to. Allowed values are 0 ... 180. Values outside
// of this range will be clipped.
func (r *Ars) SetElevation(el int) error {
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

	r.elPreset = el

	if _, err := r.write([]byte(fmt.Sprintf("N%.3d\r\n",
		r.elPreset))); err != nil {
		return err
	}

	return nil
}

// Stop stops all rotator movement
func (r *Ars) Stop() error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = r.azimuth
	r.elPreset = r.elevation

	if _, err := r.write([]byte("S\r\n")); err != nil {
		return err
	}

	return nil
}

// StopAzimuth stops horizontal rotator movement
func (r *Ars) StopAzimuth() error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = r.azimuth

	if _, err := r.write([]byte("A\r\n")); err != nil {
		return err
	}

	return nil
}

// StopElevation stops vertical rotator movement
func (r *Ars) StopElevation() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation

	if _, err := r.write([]byte("E\r\n")); err != nil {
		return err
	}

	return nil
}

// Serialize returns a a rotator.Status struct with the information
// of this rotator.
func (r *Ars) Serialize() rotator.Status {
	r.Lock()
	defer r.Unlock()
	return rotator.Status{
		Name:      r.name,
		Azimuth:   r.azimuth,
		AzPreset:  r.azPreset,
		Elevation: r.elevation,
		ElPreset:  r.elPreset,
	}
}

// ExecuteRequest takes a request struct and sets the new values
func (r *Ars) ExecuteRequest(req rotator.Request) error {
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

	return nil
}

func (r *Ars) Info() rotator.Info {
	r.Lock()
	defer r.Unlock()

	info := rotator.Info{
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
	return info
}
