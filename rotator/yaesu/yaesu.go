package yaesu

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	serial "github.com/tarm/serial"

	"github.com/dh1tw/remoteRotator/rotator"
)

// Yaesu is the implementation of the Yaesu GS232A/B rotator protocol
type Yaesu struct {
	sync.RWMutex
	name                 string
	azimuthMin           int
	azimuthMax           int
	azimuthStop          int
	azimuthOverlap       bool
	elevationMin         int
	elevationMax         int
	azimuth              int
	azPreset             int
	elevation            int
	elPreset             int
	hasAzimuth           bool
	hasElevation         bool
	azInitialized        bool
	elInitialized        bool
	pollingInterval      time.Duration
	pollingTicker        *time.Ticker
	eventHandler         func(rotator.Rotator, rotator.Heading)
	sp                   io.ReadWriteCloser
	spRead               sync.Mutex
	spWrite              sync.Mutex
	spPortName           string
	spBaudrate           int
	closeCh              chan struct{}
	errorCh              chan struct{}
	closer               sync.Once
	headingPatternGS232A *regexp.Regexp
	headingPatternGS232B *regexp.Regexp
	watchdogTs           time.Time
}

// New creates a new Yaesu object which satisfies implicitly the
// rotator.Rotator interface. Configuration settings can be set through
// functional options.
// Default settings are:
// hasAzimuth: true,
// portname: /dev/ttyACM0 (or 127.0.0.1:6001),
// pollingInterval: 5sec,
// baudrate: 9600.
func New(opts ...func(*Yaesu)) (*Yaesu, error) {

	// regex Pattern for the az&el position as per GS232A
	headingPattern232A, err := regexp.Compile(`\+[\d]{4}`)
	if err != nil {
		return nil, err
	}

	// regex Pattern for the az&el position as per GS232B
	headingPattern232B, err := regexp.Compile(`((AZ)|(EL))=[\d]{3}`)
	if err != nil {
		return nil, err
	}

	r := &Yaesu{
		hasAzimuth:           true,
		pollingInterval:      time.Second * 5,
		spPortName:           "/dev/ttyACM0",
		spBaudrate:           9600,
		headingPatternGS232A: headingPattern232A,
		headingPatternGS232B: headingPattern232B,
		azimuthMax:           450,
		elevationMax:         180,
		closeCh:              make(chan struct{}),
	}

	for _, opt := range opts {
		opt(r)
	}

	if strings.Contains(r.spPortName, ":") {
		tcpConn, err := net.Dial("tcp", r.spPortName)
		if err != nil {
			return nil, err
		}
		r.sp = tcpConn
	} else {
		spConfig := &serial.Config{
			Name:        r.spPortName,
			Baud:        r.spBaudrate,
			ReadTimeout: time.Second,
			Parity:      serial.ParityNone,
			Size:        8,
			StopBits:    1,
		}
		sp, err := serial.OpenPort(spConfig)
		if err != nil {
			return nil, err
		}
		r.sp = sp
	}

	go r.start()

	return r, nil
}

// Close shuts down the object
func (r *Yaesu) Close() {
	r.Lock()
	r.spRead.Lock()
	r.spWrite.Lock()
	defer r.Unlock()
	defer r.spWrite.Unlock()
	defer r.spRead.Unlock()

	if r.pollingTicker != nil {
		r.pollingTicker.Stop()
	}
	// makes sure that the serial port and the event loop just gets closed once
	r.closer.Do(func() {
		close(r.closeCh)
		r.sp.Close()
	})
}

// resetWatchdog resets the watchdog. This means that a packet has been
// received from the Yaesu rotator
func (r *Yaesu) resetWatchdog() {
	r.Lock()
	defer r.Unlock()
	r.watchdogTs = time.Now()
}

// checkWatchdog compares the watchdog timestamp with the current time
// and returns true if this value is greater than 5x updateInterval.
func (r *Yaesu) checkWatchdog() bool {
	r.Lock()
	defer r.Unlock()
	return time.Since(r.watchdogTs) > 5*r.pollingInterval
}

// Start the main event loop for the serial port.
// It will query the Yaesu rotator for the current heading (azimuth + elevation)
// with the pollingrate defined during initialization.
// A watchdog detects if the Yaesu rotator does not respond anymore.
// If an error occures, the errorCh will be closed.
// Consequently the communication will be shut down and the object
// prepared for garbage collection.
func (r *Yaesu) start() {
	defer r.Close()

	r.Lock()
	r.pollingTicker = time.NewTicker(r.pollingInterval)
	r.watchdogTs = time.Now()
	r.Unlock()

	// start async polling
	go r.poll()

	for {
		select {
		// when closing has been signaled, stop reading
		// from the serial port by exiting this function
		case <-r.closeCh:
			return
		default:
		}

		// this is a blocking function which will run eventually
		// into a timeout if no data is received
		msg, err := r.read()
		if err != nil {
			// serialport read is expected to timeout after 100ms
			// to unblock this routine
			if err == io.EOF {
				continue
			}
			fmt.Printf("serial port read error (%s on %s): %s\n",
				r.name, r.spPortName, err)
			close(r.errorCh)
			return // exit
		}
		r.resetWatchdog()
		r.parseMsg(msg)
	}
}

// poll the Yaesu rotator for the current heading (azimuth + elevation)
func (r *Yaesu) poll() {
	defer r.Close()

	for {
		select {
		case <-r.pollingTicker.C:
			if err := r.query(); err != nil {
				fmt.Println("serial port write error:", err)
				close(r.errorCh)
				return
			}
			if r.checkWatchdog() {
				fmt.Println("communication lost with Yaesu rotator")
				close(r.errorCh)
				return
			}
		// when closing has been signaled, stop polling and return
		case <-r.closeCh:
			return
		}
	}
}

// read from the Yaesu rotator through this wrapper function
func (r *Yaesu) read() (string, error) {
	r.spRead.Lock()
	defer r.spRead.Unlock()
	a, b := bufio.NewReader(r.sp).ReadString('\n')
	return a, b
}

// request Azimuth + Elevation from Yaesu rotator
func (r *Yaesu) query() error {
	//query azimuth + elevation
	_, err := r.write([]byte("C2\r\n"))
	return err
}

// all functions write to the Yaesu rotator / serial port through this wrapper function
func (r *Yaesu) write(data []byte) (int, error) {
	r.spWrite.Lock()
	defer r.spWrite.Unlock()
	return r.sp.Write(data)
}

// parseMsg checks the content of the received message from the Yaesu rotator
// and then further stores them and executes the event callback
func (r *Yaesu) parseMsg(msg string) {

	gotNewValue := false

	res := r.parseGS232A(msg)
	if len(res) == 0 {
		res = r.parseGS232B(msg)
	}

	if len(res) == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	az, ok := res["azimuth"]
	if ok {
		// on startup we initialize azPreset with the current azimuth position
		if !r.azInitialized {
			r.azPreset = az
			r.azInitialized = true
			gotNewValue = true
		}

		if r.azimuth != az {
			r.azimuth = az
			gotNewValue = true
		}
	}

	el, ok := res["elevation"]
	if ok {
		// on startup we initialize elPreset with the current elevation position
		if !r.elInitialized {
			r.elPreset = el
			r.elInitialized = true
		}

		if r.elevation != el {
			r.elevation = el
		}
	}

	if r.eventHandler != nil && gotNewValue {
		// cb launched async to avoid deadlock on yaesu.*()
		heading := r.serialize().Heading
		go r.eventHandler(r, heading)
	}
}

func (r *Yaesu) parseGS232A(msg string) map[string]int {

	headings := []string{}

	result := make(map[string]int)

	if r.headingPatternGS232A != nil {
		headings = r.headingPatternGS232A.FindAllString(msg, -1)
	}

	if len(headings) == 0 {
		return result
	}

	//contains at least azimuth (e.g. +0030)
	az, _ := strconv.Atoi(headings[0][2:]) //discard the first two characters

	result["azimuth"] = az

	// azimuth + elevation (e.g. +0030+0050)
	if len(headings) == 2 {

		el, _ := strconv.Atoi(headings[1][2:]) // discard the the first characters
		result["elevation"] = el
	}

	return result
}

func (r *Yaesu) parseGS232B(msg string) map[string]int {

	headings := []string{}

	result := make(map[string]int)

	if r.headingPatternGS232B != nil {
		headings = r.headingPatternGS232B.FindAllString(msg, -1)
	}

	if len(headings) == 0 {
		return result
	}

	for _, heading := range headings {

		if heading[0:3] == "AZ=" { //contains azimuth (e.g. AZ=030)
			az, _ := strconv.Atoi(heading[3:6])
			result["azimuth"] = az
		} else if heading[0:3] == "EL=" { //contains elevation (e.g. EL=090)
			el, _ := strconv.Atoi(heading[3:6])
			result["elevation"] = el
		}
	}

	return result
}

// Name returns the name of the rotator
func (r *Yaesu) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.name
}

// Azimuth returns the current horizontal heading of the rotator in degrees
func (r *Yaesu) Azimuth() int {
	r.RLock()
	defer r.RUnlock()
	return r.azimuth
}

// AzPreset returns the horizontal heading (preset) to which the rotator
// shall turn to
func (r *Yaesu) AzPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.azPreset
}

// HasAzimuth returns a boolean value indicating if this rotator supports
// horizontal rotation
func (r *Yaesu) HasAzimuth() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasAzimuth
}

// SetAzimuth sets to value of the horizontal heading to which the
// rotator shall turn to. Allowed values are 0 ... 450. Values outside
// of this range will be clipped.
func (r *Yaesu) SetAzimuth(az int) error {
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
func (r *Yaesu) Elevation() int {
	r.RLock()
	defer r.RUnlock()
	return r.elevation
}

// ElPreset returns the vertical elevation (preset) to which the rotator
// shall turn to
func (r *Yaesu) ElPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.elPreset
}

// HasElevation returns a boolean value indicating if this rotator supports
// vertical rotation
func (r *Yaesu) HasElevation() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasElevation
}

// SetElevation sets to value of the vertical elevation to which the
// rotator shall turn to. Allowed values are 0 ... 180. Values outside
// of this range will be clipped.
func (r *Yaesu) SetElevation(el int) error {
	r.RLock()
	defer r.RUnlock()

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

	if _, err := r.write([]byte(fmt.Sprintf("W%.3d %.3d\r\n", r.azPreset, r.elPreset))); err != nil {
		return err
	}

	return nil
}

// Stop stops all rotator movement
func (r *Yaesu) Stop() error {
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
func (r *Yaesu) StopAzimuth() error {
	r.Lock()
	defer r.Unlock()

	r.azPreset = r.azimuth

	if _, err := r.write([]byte("A\r\n")); err != nil {
		return err
	}

	return nil
}

// StopElevation stops vertical rotator movement
func (r *Yaesu) StopElevation() error {
	r.Lock()
	defer r.Unlock()

	r.elPreset = r.elevation

	if _, err := r.write([]byte("E\r\n")); err != nil {
		return err
	}

	return nil
}

// Serialize the data of the rotator
func (r *Yaesu) Serialize() rotator.Object {
	r.RLock()
	defer r.RUnlock()

	return r.serialize()
}

func (r *Yaesu) serialize() rotator.Object {

	obj := rotator.Object{
		Name: r.name,
		Heading: rotator.Heading{
			Azimuth:   r.azimuth,
			AzPreset:  r.azPreset,
			Elevation: r.elevation,
			ElPreset:  r.elPreset,
		},
		Config: rotator.Config{
			HasAzimuth:   r.hasAzimuth,
			AzimuthMax:   r.azimuthMax,
			AzimuthMin:   r.azimuthMin,
			AzimuthStop:  r.azimuthStop,
			HasElevation: r.hasElevation,
			ElevationMax: r.elevationMax,
			ElevationMin: r.elevationMin,
		},
	}

	return obj
}
