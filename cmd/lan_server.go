package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	// _ "net/http/pprof"

	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/micro/mdns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lanServerCmd = &cobra.Command{
	Use:   "lan",
	Short: "expose a rotator on your local network",
	Long: `
The local lan server allows you to expose a rotator to a local area network.
By default, the rotator will only be listening on the loopback adapter. In
order to make it available and discoverable on the local network, a network
connected adapter has to be selected.

remoteRotator supports access via TCP, emulating the Yaesu GS232 protocol
(disabled by default) and through a web interface (HTTP + Websocket).

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
	Run: lanServer,
}

func init() {
	serverCmd.AddCommand(lanServerCmd)

	lanServerCmd.Flags().BoolP("tcp-enabled", "", false, "enable TCP Server")
	lanServerCmd.Flags().StringP("tcp-host", "u", "127.0.0.1", "Host (use '0.0.0.0' to listen on all network adapters)")
	lanServerCmd.Flags().IntP("tcp-port", "p", 7373, "TCP Port")
	lanServerCmd.Flags().BoolP("http-enabled", "", true, "enable HTTP Server")
	lanServerCmd.Flags().StringP("http-host", "w", "127.0.0.1", "Host (use '0.0.0.0' to listen on all network adapters)")
	lanServerCmd.Flags().IntP("http-port", "k", 7070, "Port for the HTTP access to the rotator")
	lanServerCmd.Flags().BoolP("discovery-enabled", "", true, "make rotator discoverable on the network")
	lanServerCmd.Flags().StringP("portname", "P", "/dev/ttyACM0", "portname / path to the rotator (e.g. COM1)")
	lanServerCmd.Flags().IntP("baudrate", "b", 9600, "baudrate")
	lanServerCmd.Flags().StringP("type", "t", "yaesu", "Rotator type (supported: yaesu, dummy")
	lanServerCmd.Flags().StringP("name", "n", "myRotator", "Name tag for the rotator")
	lanServerCmd.Flags().BoolP("has-azimuth", "", true, "rotator supports Azimuth")
	lanServerCmd.Flags().BoolP("has-elevation", "", false, "rotator supports Elevation")
	lanServerCmd.Flags().DurationP("pollingrate", "", time.Second*1, "rotator polling rate")
	lanServerCmd.Flags().IntP("azimuth-min", "", 0, "metadata: minimum azimuth (in deg)")
	lanServerCmd.Flags().IntP("azimuth-max", "", 360, "metadata: maximum azimuth (in deg)")
	lanServerCmd.Flags().IntP("azimuth-stop", "", 0, "metadata: mechanical azimuth stop (in deg)")
	lanServerCmd.Flags().IntP("elevation-min", "", 0, "metadata: minimum elevation (in deg)")
	lanServerCmd.Flags().IntP("elevation-max", "", 180, "metadata: maximum elevation (in deg)")
}

func lanServer(cmd *cobra.Command, args []string) {

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
	viper.BindPFlag("rotator.has-azimuth", cmd.Flags().Lookup("has-azimuth"))
	viper.BindPFlag("rotator.has-elevation", cmd.Flags().Lookup("has-elevation"))
	viper.BindPFlag("rotator.pollingrate", cmd.Flags().Lookup("pollingrate"))
	viper.BindPFlag("rotator.azimuth-min", cmd.Flags().Lookup("azimuth-min"))
	viper.BindPFlag("rotator.azimuth-max", cmd.Flags().Lookup("azimuth-max"))
	viper.BindPFlag("rotator.azimuth-stop", cmd.Flags().Lookup("azimuth-stop"))
	viper.BindPFlag("rotator.elevation-min", cmd.Flags().Lookup("elevation-min"))
	viper.BindPFlag("rotator.elevation-max", cmd.Flags().Lookup("elevation-max"))

	if err := sanityCheckRotatorInputs(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := sanityCheckDiscovery(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Profiling (uncomment if needed)
	// go func() {
	// 	log.Println(http.ListenAndServe("0.0.0.0:6060", http.DefaultServeMux))
	// }()

	bcast := make(chan hub.Event, 10)

	var rEventHandler = func(r rotator.Rotator, heading rotator.Heading) {
		e := hub.Event{
			Name:        hub.UpdateHeading,
			RotatorName: r.Name(),
			Heading:     heading,
		}
		bcast <- e
	}

	rotatorError := make(chan struct{})

	// initialize our Rotator
	r, err := initRotator(viper.GetString("rotator.type"), rEventHandler, rotatorError)
	if err != nil {
		fmt.Println("unable to initialize rotator:", err)
		os.Exit(1)
	}

	h, err := hub.NewHub(r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tcpError := make(chan bool)

	// start TCP server
	if viper.GetBool("tcp.enabled") {
		go h.ListenTCP(viper.GetString("tcp.host"), viper.GetInt("tcp.port"), tcpError)
	}

	webServerError := make(chan struct{})

	// start HTTP server
	if viper.GetBool("http.enabled") {
		go h.ListenHTTP(viper.GetString("http.host"), viper.GetInt("http.port"), webServerError)
	}

	// start mDNS server
	mDNSShutdown := make(chan struct{})

	if viper.GetBool("discovery.enabled") {
		if err := startMdnsServer(mDNSShutdown); err != nil {
			log.Println(err)
		}
	}

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	for {
		select {
		case sig := <-osSignals:
			if sig == os.Interrupt {
				r.Close()
				close(mDNSShutdown)
				return
			}
		case msg := <-bcast:
			h.Broadcast(msg)
		case <-rotatorError:
			return
		case <-tcpError:
			return
		case <-webServerError:
			return
		}
	}
}

func startMdnsServer(shutdown <-chan struct{}) error {

	if !viper.GetBool("http.enabled") {
		return fmt.Errorf("discovery disabled; the HTTP server must be enabled and accessible over a network interface (e.g. 0.0.0.0)")
	}

	netif := net.ParseIP(viper.GetString("http.host"))

	if bytes.Compare(netif, net.IPv4zero) != 0 &&
		bytes.Compare(netif, net.IPv6zero) != 0 &&
		bytes.Compare(netif, getOutboundIP()) != 0 {
		return fmt.Errorf("discovery disabled; the HTTP server must listen on an accessible network interface (e.g. 0.0.0.0)")
	}

	go func() {
		mDNSService, err := mdns.NewMDNSService(viper.GetString("rotator.name"),
			"_rotator._tcp", "", "", viper.GetInt("http.port"),
			[]net.IP{getOutboundIP()}, nil)

		if err != nil {
			log.Printf("discovery disabled; unable to start mDNS service: %s\n", err)
			return
		}

		mDNSServer, err := mdns.NewServer(&mdns.Config{Zone: mDNSService})
		if err != nil {
			log.Printf("discovery disabled; unable to start mDNS service: %s\n", err)
			return
		}
		defer mDNSServer.Shutdown()
		<-shutdown
	}()

	return nil
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println("No network adapter detected; Using Loopback only")
		return net.IPv4(127, 0, 0, 1)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
