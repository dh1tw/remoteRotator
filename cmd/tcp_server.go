package cmd

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/ars"
	// _ "net/http/pprof"
)

var tcpServerCmd = &cobra.Command{
	Use:   "tcp",
	Short: "expose a rotator to the network",
	Long:  `expose a rotator to the network`,
	Run:   tcpServer,
}

func init() {
	serverCmd.AddCommand(tcpServerCmd)

	tcpServerCmd.Flags().BoolP("tcp-enabled", "", true, "enable TCP Server")
	tcpServerCmd.Flags().StringP("tcp-host", "u", "127.0.0.1", "Host (use '0.0.0.0' to listen on all network adapters)")
	tcpServerCmd.Flags().IntP("tcp-port", "p", 7373, "TCP Port")
	tcpServerCmd.Flags().BoolP("http-enabled", "", true, "enable HTTP Server")
	tcpServerCmd.Flags().StringP("http-host", "", "127.0.0.1", "Host (use '0.0.0.0' to listen on all network adapters)")
	tcpServerCmd.Flags().IntP("http-port", "", 7070, "Port for the HTTP access to the rotator")
	tcpServerCmd.Flags().BoolP("discovery-enabled", "", true, "make rotator discoverable on the network")
	tcpServerCmd.Flags().StringP("portname", "P", "/dev/ttyACM0", "portname / path to the rotator (e.g. COM1)")
	tcpServerCmd.Flags().IntP("baudrate", "b", 9600, "baudrate")
	tcpServerCmd.Flags().StringP("type", "t", "ARS", "Rotator type (supported: ARS")
	tcpServerCmd.Flags().StringP("name", "n", "myRotator", "Name tag for the rotator")
	tcpServerCmd.Flags().StringP("description", "d", "Yaesu G1000 with 4el 20m Yagi@18m ASL", "Description")
	tcpServerCmd.Flags().BoolP("has-azimuth", "", true, "Indicate if the rotator supports Azimuth")
	tcpServerCmd.Flags().BoolP("has-elevation", "", false, "Indicate if the rotator supports Elevation")
	tcpServerCmd.Flags().DurationP("pollingrate", "", time.Second*1, "rotator polling rate")
	tcpServerCmd.Flags().IntP("azimuth-min", "", 0, "metadata: minimum azimuth (in deg)")
	tcpServerCmd.Flags().IntP("azimuth-max", "", 450, "metadata: maximum azimuth (in deg)")
	tcpServerCmd.Flags().IntP("azimuth-stop", "", 0, "metadata: mechanical azimuth stop (in deg)")
	tcpServerCmd.Flags().IntP("elevation-min", "", 0, "metadata: minimum elevation (in deg)")
	tcpServerCmd.Flags().IntP("elevation-max", "", 180, "metadata: maximum elevation (in deg)")
}

func tcpServer(cmd *cobra.Command, args []string) {

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

	// check if values from config file / pflags are valid
	// if !checkAudioParameterValues() {
	// 	os.Exit(-1)
	// }

	// bind the pflags to viper settings
	viper.BindPFlag("tcp.enabled", cmd.Flags().Lookup("tcp-enabled"))
	viper.BindPFlag("tcp.host", cmd.Flags().Lookup("tcp-host"))
	viper.BindPFlag("tcp.port", cmd.Flags().Lookup("tcp-port"))
	viper.BindPFlag("http.enabled", cmd.Flags().Lookup("http-enabled"))
	viper.BindPFlag("http.host", cmd.Flags().Lookup("http-host"))
	viper.BindPFlag("http.port", cmd.Flags().Lookup("http-port"))
	viper.BindPFlag("discovery.enabled", cmd.Flags().Lookup("discovery-enabled"))
	viper.BindPFlag("rotator.portname", cmd.Flags().Lookup("portname"))
	viper.BindPFlag("rotator.baudrate", cmd.Flags().Lookup("baudrate"))
	viper.BindPFlag("rotator.type", cmd.Flags().Lookup("type"))
	viper.BindPFlag("rotator.name", cmd.Flags().Lookup("name"))
	viper.BindPFlag("rotator.description", cmd.Flags().Lookup("description"))
	viper.BindPFlag("rotator.has-azimuth", cmd.Flags().Lookup("has-azimuth"))
	viper.BindPFlag("rotator.has-elevation", cmd.Flags().Lookup("has-elevation"))
	viper.BindPFlag("rotator.pollingrate", cmd.Flags().Lookup("pollingrate"))
	viper.BindPFlag("rotator.azimuth-min", cmd.Flags().Lookup("azimuth-min"))
	viper.BindPFlag("rotator.azimuth-max", cmd.Flags().Lookup("azimuth-max"))
	viper.BindPFlag("rotator.azimuth-stop", cmd.Flags().Lookup("azimuth-stop"))
	viper.BindPFlag("rotator.elevation-min", cmd.Flags().Lookup("elevation-min"))
	viper.BindPFlag("rotator.elevation-max", cmd.Flags().Lookup("elevation-max"))

	// go func() {
	// 	log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
	// }()

	bcast := make(chan rotator.Status, 10)

	var arsEventHandler = func(r rotator.Rotator, ev rotator.Event, value ...interface{}) {
		fmt.Println(ev, value)
		switch ev {
		case rotator.Azimuth, rotator.Elevation:
			bcast <- r.Serialize()
		default:
			log.Printf("unknown event: %v with value(s): %v\n", ev, value)
		}
	}

	evHandler := ars.EventHandler(arsEventHandler)
	name := ars.Name(viper.GetString("rotator.name"))
	interval := ars.UpdateInterval(viper.GetDuration("rotator.pollingrate"))
	spPortName := ars.Portname(viper.GetString("rotator.portname"))
	baudrate := ars.Baudrate(viper.GetInt("rotator.baudrate"))
	hasAzimuth := ars.HasAzimuth(viper.GetBool("rotator.has-azimuth"))
	hasElevation := ars.HasElevation(viper.GetBool("rotator.has-elevation"))

	ars, err := ars.NewArs(name, interval, evHandler,
		spPortName, baudrate, hasAzimuth, hasElevation)
	if err != nil {
		fmt.Println("unable to initialize ARS:", err)
		os.Exit(1)
	}

	defer ars.Close()

	h := hub.NewHub(ars)

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	arsError := make(chan bool)
	arsShutdown := make(chan bool)
	go ars.Start(arsError, arsShutdown)

	tcpError := make(chan bool)

	// start TCP server
	if viper.GetBool("tcp.enabled") {
		go h.ListenTCP(viper.GetString("tcp.host"), viper.GetInt("tcp.port"), tcpError)
	}

	wsError := make(chan bool)

	// start HTTP server
	if viper.GetBool("http.enabled") {
		go h.ListenHTTP(viper.GetString("http.host"), viper.GetInt("http.port"), wsError)
	}

	// start mDNS server
	if viper.GetBool("discovery.enabled") {

		i := rotator.Info{
			Name:         viper.GetString("rotator.name"),
			Description:  viper.GetString("rotator.description"),
			HasAzimuth:   viper.GetBool("rotator.has-azimuth"),
			HasElevation: viper.GetBool("rotator.has-elevation"),
			AzimuthMin:   viper.GetInt("rotator.azimuth-min"),
			AzimuthMax:   viper.GetInt("rotator.azimuth-max"),
			AzimuthStop:  viper.GetInt("rotator.azimuth-stop"),
			ElevationMin: viper.GetInt("rotator.elevation-min"),
			ElevationMax: viper.GetInt("rotator.elevation-max"),
		}

		info, err := encodeInfo(i)
		if err != nil {
			fmt.Printf("unable to marshal rotator description: %s\n", err)
			return
		}

		mDNSService, err := mdns.NewMDNSService(viper.GetString("rotator.name"),
			"rotators.shackbus", "", "", 7375, nil, []string{info})

		if err != nil {
			fmt.Printf("unable to start mDNS discovery service: %v", err)
			return
		}
		mDNSServer, _ := mdns.NewServer(&mdns.Config{Zone: mDNSService})
		defer mDNSServer.Shutdown()
	}

	for {
		select {
		case sig := <-osSignals:
			if sig == os.Interrupt {
				close(arsShutdown)
				return
			}
		case msg := <-bcast:
			h.Broadcast(msg)
		case <-arsError:
			return
		case <-tcpError:
			return
		case <-wsError:
			return
		}
	}

}

func encodeInfo(i rotator.Info) (string, error) {
	res, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	uEnc := b64.URLEncoding.EncodeToString(res)
	return uEnc, nil
}
