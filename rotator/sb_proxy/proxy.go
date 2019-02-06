package sbProxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/dh1tw/remoteRotator/rotator"
	sbRotator "github.com/dh1tw/remoteRotator/sb_rotator"
	"github.com/gogo/protobuf/proto"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
)

type SbProxy struct {
	sync.RWMutex
	cli            client.Client
	rcli           sbRotator.RotatorClient
	eventHandler   func(rotator.Rotator, rotator.Heading)
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
	serviceName    string //better call it address (?)
}

// New returns the pointer to an initialized Rotator proxy object.
func New(opts ...func(*SbProxy)) (*SbProxy, error) {

	r := &SbProxy{
		name:        "rotatorProxy",
		serviceName: "shackbus.rotator.myRotator",
	}

	for _, opt := range opts {
		opt(r)
	}

	fmt.Println("serviceName is:", r.serviceName)
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
// multiple times closing this channel. Closing the doneCh signals the
// application that this object can be disposed
func (r *SbProxy) closeDone() {
	r.doneOnce.Do(func() { close(r.doneCh) })
}

func (r *SbProxy) updateHandler(p broker.Publication) error {

	state := sbRotator.State{}
	err := proto.Unmarshal(p.Message().Body, &state)
	if err != nil {
		return err
	}

	r.Lock()
	defer r.Unlock()

	r.azimuth = int(state.Azimuth)
	r.azPreset = int(state.AzimuthPreset)
	r.elevation = int(state.Elevation)
	r.elPreset = int(state.ElevationPreset)

	if r.eventHandler != nil {
		go r.eventHandler(r, r.serialize().Heading)
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
	_, err := r.rcli.SetAzimuth(context.Background(), &sbRotator.HeadingReq{Heading: int32(az)})
	return err
}

func (r *SbProxy) SetElevation(el int) error {
	_, err := r.rcli.SetElevation(context.Background(), &sbRotator.HeadingReq{Heading: int32(el)})
	return err
}

func (r *SbProxy) StopAzimuth() error {
	_, err := r.rcli.StopAzimuth(context.Background(), &sbRotator.None{})
	return err
}

func (r *SbProxy) StopElevation() error {
	_, err := r.rcli.StopElevation(context.Background(), &sbRotator.None{})
	return err
}

func (r *SbProxy) Stop() error {
	_, err := r.rcli.StopAzimuth(context.Background(), &sbRotator.None{})
	_, err = r.rcli.StopElevation(context.Background(), &sbRotator.None{})
	return err
}

// Serialize the data of the rotator
func (r *SbProxy) Serialize() rotator.Object {
	r.RLock()
	defer r.RUnlock()
	return r.serialize()
}

func (r *SbProxy) serialize() rotator.Object {

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
func (r *SbProxy) Close() {
	if r.subscriber != nil {
		r.subscriber.Unsubscribe()
	}
	r.closeDone()
}
