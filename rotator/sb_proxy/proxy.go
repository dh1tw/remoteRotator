package sbProxy

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/dh1tw/remoteRotator/rotator"
	sbRotator "github.com/dh1tw/remoteRotator/sb_rotator"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
)

type SbProxy struct {
	sync.RWMutex
	cli            client.Client
	rcli           sbRotator.RotatorClient
	eventHandler   func(rotator.Rotator, rotator.Event, ...interface{})
	name           string
	azimuthMin     int
	azimuthMax     int
	azimuthStop    int
	azimuthOverlap bool
	elevationMin   int
	elevationMax   int
	hasAzimuth     bool
	hasElevation   bool
	azimuth        int
	azPreset       int
	elevation      int
	elPreset       int
	doneCh         chan struct{}
	doneOnce       sync.Once
	subscriber     broker.Subscriber
	serviceName    string
}

func Client(cli client.Client) func(*SbProxy) {
	return func(r *SbProxy) {
		r.cli = cli
	}
}

// DoneCh is a functional option allows you to pass a channel to the proxy object.
// The channel will be closed and thus notifies you when the object has been deleted.
func DoneCh(ch chan struct{}) func(*SbProxy) {
	return func(r *SbProxy) {
		r.doneCh = ch
	}
}

func Name(name string) func(*SbProxy) {
	return func(r *SbProxy) {
		r.name = name
	}
}

func ServiceName(name string) func(*SbProxy) {
	return func(r *SbProxy) {
		r.serviceName = name
	}
}

// EventHandler sets a callback function through which the proxy rotator
// will report Events
func EventHandler(h func(rotator.Rotator, rotator.Event, ...interface{})) func(*SbProxy) {
	return func(r *SbProxy) {
		r.eventHandler = h
	}
}

// New returns the pointer to an initalized Rotator proxy object.
func New(opts ...func(*SbProxy)) (*SbProxy, error) {

	r := &SbProxy{
		name:        "rotatorProxy",
		serviceName: "mystation.shackbus.rotator.myRotator",
	}

	for _, opt := range opts {
		opt(r)
	}

	r.rcli = sbRotator.NewRotatorClient(r.serviceName, r.cli)

	if err := r.getInfo(); err != nil {
		return nil, err
	}

	br := r.cli.Options().Broker
	if err := br.Connect(); err != nil {
		return nil, err
	}

	sub, err := br.Subscribe(r.serviceName+".state", r.updateHandler)
	if err != nil {
		return nil, err
	}
	r.subscriber = sub

	return r, nil
}

// the doneCh must be closed through this function to avoid
// multiple times closing this channel
func (r *SbProxy) closeDone() {
	r.doneOnce.Do(func() { close(r.doneCh) })
}

func (r *SbProxy) updateHandler(p broker.Publication) error {

	state := sbRotator.State{}
	err := json.Unmarshal(p.Message().Body, &state)
	if err != nil {
		return err
	}

	r.Lock()
	defer r.Unlock()

	r.azimuth = int(state.Azimuth)
	r.azPreset = int(state.AzimuthPreset)
	r.elevation = int(state.Elevation)
	r.elPreset = int(state.ElevationPreset)

	status := rotator.Status{
		Name:           r.name,
		Azimuth:        r.azimuth,
		AzPreset:       r.azPreset,
		AzimuthOverlap: r.azimuthOverlap,
		Elevation:      r.elevation,
		ElPreset:       r.elPreset,
	}

	if r.eventHandler != nil {
		go r.eventHandler(r, rotator.Azimuth, status)
	}

	return nil
}

func (r *SbProxy) getInfo() error {

	md, err := r.rcli.GetMetadata(context.Background(), &sbRotator.None{})
	if err != nil {
		return err
	}
	r.azimuthMax = int(md.AzimuthMax)
	r.azimuthMin = int(md.AzimuthMin)
	r.azimuthStop = int(md.AzimuthStop)
	r.elevationMin = int(md.ElevationMin)
	r.elevationMax = int(md.ElevationMax)
	r.hasAzimuth = md.HasAzimuth
	r.hasElevation = md.HasElevation

	state, err := r.rcli.GetState(context.Background(), &sbRotator.None{})
	if err != nil {
		return err
	}
	r.azimuth = int(state.Azimuth)
	r.azPreset = int(state.AzimuthPreset)
	r.elevation = int(state.Elevation)
	r.elPreset = int(state.ElevationPreset)

	return nil
}

func (r *SbProxy) Name() string {
	r.RLock()
	defer r.RUnlock()
	return r.name
}

func (r *SbProxy) HasAzimuth() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasAzimuth
}

func (r *SbProxy) HasElevation() bool {
	r.RLock()
	defer r.RUnlock()
	return r.hasElevation
}

func (r *SbProxy) Azimuth() int {
	r.RLock()
	defer r.RUnlock()
	return r.azimuth
}

func (r *SbProxy) AzPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.azPreset
}

func (r *SbProxy) Elevation() int {
	r.RLock()
	defer r.RUnlock()
	return r.elevation
}

func (r *SbProxy) ElPreset() int {
	r.RLock()
	defer r.RUnlock()
	return r.elPreset
}

func (r *SbProxy) SetAzimuth(az int) error {
	log.Println("setting azimuth")

	_, err := r.rcli.SetAzimuth(context.Background(), &sbRotator.HeadingReq{Heading: int32(az)})
	return err
}

func (r *SbProxy) SetElevation(el int) error {
	log.Println("setting elevation")
	_, err := r.rcli.SetElevation(context.Background(), &sbRotator.HeadingReq{Heading: int32(el)})
	return err
}

func (r *SbProxy) StopAzimuth() error {
	log.Println("stopping azimuth")
	_, err := r.rcli.StopAzimuth(context.Background(), &sbRotator.None{})
	return err
}

func (r *SbProxy) StopElevation() error {
	log.Println("stopping elevation")
	_, err := r.rcli.StopElevation(context.Background(), &sbRotator.None{})
	return err
}

func (r *SbProxy) Stop() error {
	log.Println("stopping all")
	_, err := r.rcli.StopAzimuth(context.Background(), &sbRotator.None{})
	_, err = r.rcli.StopElevation(context.Background(), &sbRotator.None{})
	return err
}

func (r *SbProxy) Status() rotator.Status {
	r.RLock()
	defer r.RUnlock()
	s := rotator.Status{
		Azimuth:        r.azimuth,
		AzPreset:       r.azPreset,
		AzimuthOverlap: false,
		Elevation:      r.elevation,
		ElPreset:       r.elPreset,
	}
	return s
}

func (r *SbProxy) ExecuteRequest(req rotator.Request) error {

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

	if req.Stop {
		if err := r.Stop(); err != nil {
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

func (r *SbProxy) Info() rotator.Info {
	r.RLock()
	defer r.RUnlock()
	i := rotator.Info{
		Name:           r.name,
		HasAzimuth:     r.hasAzimuth,
		HasElevation:   r.hasElevation,
		AzimuthMin:     r.azimuthMin,
		AzimuthMax:     r.azimuthMax,
		AzimuthStop:    r.azimuthStop,
		AzimuthOverlap: r.azimuthOverlap,
		ElevationMin:   r.elevationMin,
		ElevationMax:   r.elevationMax,
		Azimuth:        r.azimuth,
		AzPreset:       r.azPreset,
		Elevation:      r.elevation,
		ElPreset:       r.elPreset,
	}
	return i
}

func (r *SbProxy) Close() {
	if r.subscriber != nil {
		err := r.subscriber.Unsubscribe()
		if err != nil {
			log.Println("unsubscribe problem:", err)
		}
	}
	r.closeDone()
}
