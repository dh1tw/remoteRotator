package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport"
	natsBroker "github.com/micro/go-plugins/broker/nats"
	natsReg "github.com/micro/go-plugins/registry/nats"
	natsTr "github.com/micro/go-plugins/transport/nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sbRotator "github.com/dh1tw/remoteRotator/sb_rotator"

	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/dummy"
	"github.com/dh1tw/remoteRotator/rotator/yaesu"
	// _ "net/http/pprof"
)

var natsServerCmd = &cobra.Command{
	Use:   "nats",
	Short: "nats",
	Long:  ``,
	Run:   natsServer,
}

func init() {
	serverCmd.AddCommand(natsServerCmd)

	natsServerCmd.Flags().StringP("portname", "d", "/dev/ttyACM0", "portname / path to the rotator (e.g. COM1)")
	natsServerCmd.Flags().IntP("baudrate", "b", 9600, "baudrate")
	natsServerCmd.Flags().StringP("type", "t", "yaesu", "Rotator type (supported: yaesu, dummy")
	natsServerCmd.Flags().StringP("name", "n", "myRotator", "Name tag for the rotator")
	natsServerCmd.Flags().BoolP("has-azimuth", "", true, "rotator supports Azimuth")
	natsServerCmd.Flags().BoolP("has-elevation", "", false, "rotator supports Elevation")
	natsServerCmd.Flags().DurationP("pollingrate", "", time.Second*1, "rotator polling rate")
	natsServerCmd.Flags().IntP("azimuth-min", "", 0, "metadata: minimum azimuth (in deg)")
	natsServerCmd.Flags().IntP("azimuth-max", "", 360, "metadata: maximum azimuth (in deg)")
	natsServerCmd.Flags().IntP("azimuth-stop", "", 0, "metadata: mechanical azimuth stop (in deg)")
	natsServerCmd.Flags().IntP("elevation-min", "", 0, "metadata: minimum elevation (in deg)")
	natsServerCmd.Flags().IntP("elevation-max", "", 180, "metadata: maximum elevation (in deg)")
	natsServerCmd.Flags().StringP("station", "X", "mystation", "Your station callsign")
	natsServerCmd.Flags().StringP("broker-url", "u", "localhost", "Broker URL")
	natsServerCmd.Flags().IntP("broker-port", "p", 4222, "Broker Port")
	natsServerCmd.Flags().StringP("password", "P", "", "NATS Password")
	natsServerCmd.Flags().StringP("username", "U", "", "NATS Username")
}

func natsServer(cmd *cobra.Command, args []string) {

	// Try to read config file
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if strings.Contains(err.Error(), "Not Found in") {
			fmt.Println("no config file found")
		} else {
			fmt.Println("Error parsing config file", viper.ConfigFileUsed())
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	// bind the pflags to viper settings
	viper.BindPFlag("rotator.portname", cmd.Flags().Lookup("portname"))
	viper.BindPFlag("rotator.baudrate", cmd.Flags().Lookup("baudrate"))
	viper.BindPFlag("rotator.type", cmd.Flags().Lookup("type"))
	viper.BindPFlag("rotator.name", cmd.Flags().Lookup("name"))
	viper.BindPFlag("rotator.has-azimuth", cmd.Flags().Lookup("has-azimuth"))
	viper.BindPFlag("rotator.has-elevation", cmd.Flags().Lookup("has-elevation"))
	viper.BindPFlag("rotator.pollingrate", cmd.Flags().Lookup("pollingrate"))
	viper.BindPFlag("rotator.azimuth-min", cmd.Flags().Lookup("azimuth-min"))
	viper.BindPFlag("rotator.azimuth-max", cmd.Flags().Lookup("azimuth-max"))
	viper.BindPFlag("rotator.azimuth-stop", cmd.Flags().Lookup("azimuth-stop"))
	viper.BindPFlag("rotator.elevation-min", cmd.Flags().Lookup("elevation-min"))
	viper.BindPFlag("rotator.elevation-max", cmd.Flags().Lookup("elevation-max"))
	viper.BindPFlag("shackbus.station", cmd.Flags().Lookup("station"))
	viper.BindPFlag("nats.broker-url", cmd.Flags().Lookup("broker-url"))
	viper.BindPFlag("nats.broker-port", cmd.Flags().Lookup("broker-port"))
	viper.BindPFlag("nats.password", cmd.Flags().Lookup("password"))
	viper.BindPFlag("nats.username", cmd.Flags().Lookup("username"))

	if len(viper.GetString("rotator.name")) == 0 {
		log.Println("rotator name must not be empty")
		os.Exit(1)
	}

	if viper.GetBool("rotator.has-azimuth") {

		if viper.GetInt("rotator.azimuth-min") >= viper.GetInt("rotator.azimuth-max") {
			log.Println("azimuth-min must be smaller than azimuth-max")
			os.Exit(1)
		}

		if viper.GetInt("rotator.azimuth-max") > 360 && viper.GetInt("rotator.azimuth-min") > 360 {
			log.Println("if azimuth-max is >360, azimuth-min must be < 360")
			os.Exit(1)
		}

		if viper.GetInt("rotator.azimuth-min") < 0 {
			log.Println("azimuth-min must be >= 0")
			os.Exit(1)
		}

		if viper.GetInt("rotator.azimuth-max") > 500 {
			log.Println("azimuth-min must be <= 500")
			os.Exit(1)
		}
	}

	if viper.GetBool("rotator.has-elevation") {

		if viper.GetInt("rotator.elevation-min") < 0 {
			log.Println("elevation-min must be >= 0")
			os.Exit(1)
		}

		if viper.GetInt("rotator.elevation-max") > 180 {
			log.Println("elevation-min must be <= 180")
			os.Exit(1)
		}
	}

	// go func() {
	// 	log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
	// }()

	bcast := make(chan rotator.Status, 10)

	var yaesuEventHandler = func(r rotator.Rotator, ev rotator.Event, value ...interface{}) {
		// fmt.Println(ev, value)
		switch ev {
		case rotator.Azimuth, rotator.Elevation:
			if len(value) == 0 {
				return
			}
			switch value[0].(type) {
			case rotator.Status:
				bcast <- value[0].(rotator.Status)
			}
		default:
			log.Printf("unknown event: %v with value(s): %v\n", ev, value)
		}
	}

	rotatorError := make(chan struct{})

	var r rotator.Rotator

	switch strings.ToUpper(viper.GetString("rotator.type")) {

	case "YAESU":
		evHandler := yaesu.EventHandler(yaesuEventHandler)
		name := yaesu.Name(viper.GetString("rotator.name"))
		interval := yaesu.UpdateInterval(viper.GetDuration("rotator.pollingrate"))
		spPortName := yaesu.Portname(viper.GetString("rotator.portname"))
		baudrate := yaesu.Baudrate(viper.GetInt("rotator.baudrate"))
		hasAzimuth := yaesu.HasAzimuth(viper.GetBool("rotator.has-azimuth"))
		hasElevation := yaesu.HasElevation(viper.GetBool("rotator.has-elevation"))
		azMin := yaesu.AzimuthMin(viper.GetInt("rotator.azimuth-min"))
		azMax := yaesu.AzimuthMax(viper.GetInt("rotator.azimuth-max"))
		elMin := yaesu.ElevationMin(viper.GetInt("rotator.elevation-min"))
		elMax := yaesu.ElevationMax(viper.GetInt("rotator.elevation-max"))
		azStop := yaesu.AzimuthStop(viper.GetInt("rotator.azimuth-stop"))
		errorCh := yaesu.ErrorCh(rotatorError)

		yaesu, err := yaesu.New(name, interval, evHandler,
			spPortName, baudrate, hasAzimuth, hasElevation, azMin, azMax, elMin,
			elMax, azStop, errorCh)
		if err != nil {
			fmt.Println("unable to initialize YAESU rotator:", err)
			os.Exit(1)
		}
		r = yaesu

	case "DUMMY":
		evHandler := dummy.EventHandler(yaesuEventHandler)
		name := dummy.Name(viper.GetString("rotator.name"))
		hasAzimuth := dummy.HasAzimuth(viper.GetBool("rotator.has-azimuth"))
		hasElevation := dummy.HasElevation(viper.GetBool("rotator.has-elevation"))
		azMin := dummy.AzimuthMin(viper.GetInt("rotator.azimuth-min"))
		azMax := dummy.AzimuthMax(viper.GetInt("rotator.azimuth-max"))
		elMin := dummy.ElevationMin(viper.GetInt("rotator.elevation-min"))
		elMax := dummy.ElevationMax(viper.GetInt("rotator.elevation-max"))
		azStop := dummy.AzimuthStop(viper.GetInt("rotator.azimuth-stop"))

		dummyRotator, err := dummy.New(name, evHandler, hasAzimuth, hasElevation, azMin, azMax, azStop, elMin, elMax)
		if err != nil {
			fmt.Println("unable to initialize Dummy rotator:", err)
			os.Exit(1)
		}

		r = dummyRotator

	default:
		log.Printf("unknown rotator type (%v)\n", viper.GetString("rotator.type"))
		os.Exit(1)
	}

	serviceName := fmt.Sprintf("%s.shackbus.rotator.%s",
		viper.GetString("shackbus.station"),
		viper.GetString("rotator.name"))

	username := viper.GetString("nats.username")
	password := viper.GetString("nats.password")
	credentials := ""
	if len(username) > 0 && len(password) > 0 {
		credentials = fmt.Sprintf("%s:%s@", username, password)
	}
	url := viper.GetString("nats.broker-url")
	port := viper.GetInt("nats.broker-port")
	addr := fmt.Sprintf("nats://%s%s:%v", credentials, url, port)

	regTimeout := registry.Timeout(time.Millisecond * 200)

	reg := natsReg.NewRegistry(registry.Addrs(addr), regTimeout)
	tr := natsTr.NewTransport(transport.Addrs(addr))
	br := natsBroker.NewBroker(broker.Addrs(addr))

	if version == "" {
		version = "dev"
	}

	rs := micro.NewService(
		micro.Name(serviceName),
		micro.RegisterInterval(time.Second*10),
		micro.Broker(br),
		micro.Transport(tr),
		micro.Registry(reg),
		micro.Version(version),
	)

	rs.Init()

	if err := br.Init(); err != nil {
		log.Println(err)
	}

	if err := br.Connect(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	rpcr := rpcRotator{
		rotator:     r,
		service:     rs,
		broker:      br,
		pubSubTopic: fmt.Sprintf("%s.state", serviceName),
	}

	sbRotator.RegisterRotatorHandler(rs.Server(), &rpcr)

	go func() {
		for {
			select {
			case newState := <-bcast:
				rpcr.PublishState(newState)
			case <-rotatorError:
				rs.Server().Stop()
				return
			}
		}
	}()

	if err := rs.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

type rpcRotator struct {
	service     micro.Service
	rotator     rotator.Rotator
	broker      broker.Broker
	pubSubTopic string
}

func (r *rpcRotator) PublishState(status rotator.Status) {
	state := sbRotator.State{
		Azimuth:         int32(status.Azimuth),
		AzimuthPreset:   int32(status.AzPreset),
		Elevation:       int32(status.Elevation),
		ElevationPreset: int32(status.ElPreset),
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Println(err)
	}

	msg := broker.Message{
		Body: data,
	}

	if err := r.broker.Publish(r.pubSubTopic, &msg); err != nil {
		log.Println(err)
		r.shutdown()
	}
}

func (r *rpcRotator) shutdown() {
	r.service.Server().Stop()
	os.Exit(1)
}

func (r *rpcRotator) SetAzimuth(ctx context.Context, req *sbRotator.HeadingReq, resp *sbRotator.None) error {
	if r.rotator.HasAzimuth() {
		fmt.Printf("setting azimuth to %v\n", req.Heading)
		err := r.rotator.SetAzimuth(int(req.Heading))
		return err
	}
	return fmt.Errorf("rotator does not support azimuth")
}

func (r *rpcRotator) SetElevation(ctx context.Context, req *sbRotator.HeadingReq, resp *sbRotator.None) error {
	if r.rotator.HasElevation() {
		err := r.rotator.SetElevation(int(req.Heading))
		return err
	}
	return fmt.Errorf("rotator does not support elevation")
}

func (r *rpcRotator) StopAzimuth(ctx context.Context, req *sbRotator.None, resp *sbRotator.None) error {
	if r.rotator.HasAzimuth() {
		return r.rotator.StopAzimuth()
	}
	return fmt.Errorf("rotator does not support azimuth")
}

func (r *rpcRotator) StopElevation(ctx context.Context, req *sbRotator.None, resp *sbRotator.None) error {
	if r.rotator.HasElevation() {
		return r.rotator.StopElevation()
	}
	return fmt.Errorf("rotator does not support elevation")
}

func (r *rpcRotator) GetMetadata(ctx context.Context, req *sbRotator.None, resp *sbRotator.Metadata) error {
	info := r.rotator.Info()
	resp.AzimuthMax = int32(info.AzimuthMax)
	resp.AzimuthMin = int32(info.AzimuthMin)
	resp.AzimuthStop = int32(info.AzimuthStop)
	resp.ElevationMax = int32(info.ElevationMax)
	resp.ElevationMin = int32(info.ElevationMin)
	resp.HasAzimuth = info.HasAzimuth
	resp.HasElevation = info.HasElevation
	return nil
}

func (r *rpcRotator) GetState(ctx context.Context, req *sbRotator.None, resp *sbRotator.State) error {
	info := r.rotator.Info()
	resp.Azimuth = int32(info.Azimuth)
	resp.AzimuthPreset = int32(info.AzPreset)
	resp.Elevation = int32(info.Elevation)
	resp.ElevationPreset = int32(info.ElPreset)
	return nil
}
