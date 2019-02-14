package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dh1tw/remoteRotator/rotator"
	sbRotator "github.com/dh1tw/remoteRotator/sb_rotator"
	"github.com/gogo/protobuf/proto"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/server"
	natsBroker "github.com/micro/go-plugins/broker/nats"
	natsReg "github.com/micro/go-plugins/registry/nats"
	natsTr "github.com/micro/go-plugins/transport/nats"
	nats "github.com/nats-io/go-nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// _ "net/http/pprof"
)

var natsServerCmd = &cobra.Command{
	Use:   "nats",
	Short: "expose your rotator via a nats broker",
	Long: `
The nats server allows you to expose a rotator on a nats.io broker. The broker
can be located within your local lan or somewhere on the internet.

You can select the following rotator types:
1. Yaesu (GS232 compatible)
2. Dummy (great for testing)

remoteRotator allows to assign a series of meta data to a rotator:
1. Name
2. Azimuth/Elevation minimum value
3. Azimuth/Elevation maximum value
4. Azimuth Mechanical stop

These metadata enhance the rotators view (e.g. showing overlap) in the web
interface and can also help to limit for example the rotators range if it does
not support full 360°.

`,
	Run: natsServer,
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
			os.Exit(1)
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
	viper.BindPFlag("nats.broker-url", cmd.Flags().Lookup("broker-url"))
	viper.BindPFlag("nats.broker-port", cmd.Flags().Lookup("broker-port"))
	viper.BindPFlag("nats.password", cmd.Flags().Lookup("password"))
	viper.BindPFlag("nats.username", cmd.Flags().Lookup("username"))

	if err := sanityCheckRotatorInputs(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Profiling (uncomment if needed)
	// go func() {
	// 	log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
	// }()

	// struct which holds the rotator.Rotator instance, implements the
	// RPC Service methods and publishes changes via the Broker
	rpcRot := &rpcRotator{}

	rotatorError := make(chan struct{})

	// initialize our Rotator
	r, err := initRotator(viper.GetString("rotator.type"), rpcRot.PublishState, rotatorError)
	if err != nil {
		fmt.Println("unable to initialize rotator:", err)
		os.Exit(1)
	}

	// better call this Addrs(?)
	serviceName := fmt.Sprintf("shackbus.rotator.%s", viper.GetString("rotator.name"))

	username := viper.GetString("nats.username")
	password := viper.GetString("nats.password")
	url := viper.GetString("nats.broker-url")
	port := viper.GetInt("nats.broker-port")
	addr := fmt.Sprintf("nats://%s:%v", url, port)

	// start from default nats config and add the common options
	nopts := nats.GetDefaultOptions()
	nopts.Servers = []string{addr}
	nopts.User = username
	nopts.Password = password
	nopts.Timeout = time.Second * 10

	connClosed := make(chan struct{})

	disconnectedHdlr := func(conn *nats.Conn) {
		log.Println("connection to nats broker closed")
		connClosed <- struct{}{}
	}

	errorHdlr := func(conn *nats.Conn, sub *nats.Subscription, err error) {
		log.Printf("Error Handler called (%s): %s", sub.Subject, err)
	}
	nopts.AsyncErrorCB = errorHdlr

	regNatsOpts := nopts
	brNatsOpts := nopts
	trNatsOpts := nopts
	regNatsOpts.DisconnectedCB = disconnectedHdlr
	// we want to set the nats.Options.Name so that we can distinguish
	// them when monitoring the nats server with nats-top
	regNatsOpts.Name = serviceName + ":registry"
	brNatsOpts.Name = serviceName + ":broker"
	trNatsOpts.Name = serviceName + ":transport"

	// create instances of our nats Registry, Broker and Transport
	reg := natsReg.NewRegistry(natsReg.Options(regNatsOpts))
	br := natsBroker.NewBroker(natsBroker.Options(brNatsOpts))
	tr := natsTr.NewTransport(natsTr.Options(trNatsOpts))

	// this is a workaround since we must set server.Address with the
	// sanitized version of our service name. The server.Address will be
	// used in nats as the topic on which the server (transport) will be
	// listening on.
	svr := server.NewServer(
		server.Name(serviceName),
		server.Address(validateSubject(serviceName)),
		server.RegisterInterval(time.Second*10),
		server.Transport(tr),
		server.Registry(reg),
		server.Broker(br),
	)

	// version is typically defined through a git tag and injected during
	// compilation; if not, just set it to "dev"
	if version == "" {
		version = "dev"
	}

	// let's create the new rotator service
	rs := micro.NewService(
		micro.Name(serviceName),
		micro.Broker(br),
		micro.Transport(tr),
		micro.Registry(reg),
		micro.Version(version),
		micro.Server(svr),
	)

	// initalize our service
	rs.Init()

	// before we annouce this service, we have to ensure that no other
	// service with the same name exists. Therefore we query the
	// registry for all other existing services.
	services, err := reg.ListServices()
	if err != nil {
		log.Fatal(err)
	}

	// if a service with this name already exists, then exit
	for _, service := range services {
		if service.Name == serviceName {
			log.Fatalf("service '%s' already exists", service.Name)
		}
	}

	rpcRot.rotator = r
	rpcRot.service = rs
	rpcRot.pubSubTopic = fmt.Sprintf("%s.state", strings.Replace(serviceName, " ", "_", -1))

	// register our Rotator RPC handler
	sbRotator.RegisterRotatorHandler(rs.Server(), rpcRot)

	rpcRot.initialized = true

	go func() {
		for {
			select {
			case <-rotatorError:
				rs.Server().Stop()
				os.Exit(1)
			case <-connClosed:
				rs.Server().Stop()
				os.Exit(1)
			}
		}
	}()

	if err := rs.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

type rpcRotator struct {
	initialized bool
	service     micro.Service
	rotator     rotator.Rotator
	pubSubTopic string
}

func (r *rpcRotator) PublishState(rot rotator.Rotator, heading rotator.Heading) {

	if !r.initialized {
		return
	}

	state := sbRotator.State{
		Azimuth:         int32(heading.Azimuth),
		AzimuthPreset:   int32(heading.AzPreset),
		Elevation:       int32(heading.Elevation),
		ElevationPreset: int32(heading.ElPreset),
	}

	data, err := proto.Marshal(&state)
	if err != nil {
		log.Println(err)
	}

	msg := broker.Message{
		Body: data,
	}

	if err := r.service.Options().Broker.Publish(r.pubSubTopic, &msg); err != nil {
		log.Println(err)
	}
}

//implementation of the RPC shackbus.Rotator.Rotator Service
func (r *rpcRotator) SetAzimuth(ctx context.Context, req *sbRotator.HeadingReq, resp *sbRotator.None) error {
	if r.rotator.HasAzimuth() {
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
	config := r.rotator.Serialize().Config
	resp.AzimuthMax = int32(config.AzimuthMax)
	resp.AzimuthMin = int32(config.AzimuthMin)
	resp.AzimuthStop = int32(config.AzimuthStop)
	resp.ElevationMax = int32(config.ElevationMax)
	resp.ElevationMin = int32(config.ElevationMin)
	resp.HasAzimuth = config.HasAzimuth
	resp.HasElevation = config.HasElevation
	return nil
}

func (r *rpcRotator) GetState(ctx context.Context, req *sbRotator.None, resp *sbRotator.State) error {
	heading := r.rotator.Serialize().Heading
	resp.Azimuth = int32(heading.Azimuth)
	resp.AzimuthPreset = int32(heading.AzPreset)
	resp.Elevation = int32(heading.Elevation)
	resp.ElevationPreset = int32(heading.ElPreset)
	return nil
}
