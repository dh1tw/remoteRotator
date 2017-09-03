package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

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

	tcpServerCmd.Flags().StringP("host", "u", "127.0.0.1", "Host (use '0.0.0.0' for public access)")
	tcpServerCmd.Flags().IntP("port", "p", 7373, "TCP Port")
	tcpServerCmd.Flags().StringP("portname", "P", "/dev/ttyACM0", "portname / path to the rotator (e.g. COM1)")
	tcpServerCmd.Flags().IntP("baudrate", "b", 9600, "baudrate")
	tcpServerCmd.Flags().StringP("type", "t", "ARS", "Rotator type (supported: ARS")
	tcpServerCmd.Flags().StringP("name", "n", "myRotator", "Name tag for the rotator")
	tcpServerCmd.Flags().BoolP("has-azimuth", "", true, "Indicate if the rotator supports Azimuth")
	tcpServerCmd.Flags().BoolP("has-elevation", "", false, "Indicate if the rotator supports Elevation")
	tcpServerCmd.Flags().DurationP("pollingrate", "", time.Second*1, "rotator polling rate")
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
	viper.BindPFlag("tcp.host", cmd.Flags().Lookup("host"))
	viper.BindPFlag("tcp.port", cmd.Flags().Lookup("port"))
	viper.BindPFlag("rotator.portname", cmd.Flags().Lookup("portname"))
	viper.BindPFlag("rotator.baudrate", cmd.Flags().Lookup("baudrate"))
	viper.BindPFlag("rotator.type", cmd.Flags().Lookup("type"))
	viper.BindPFlag("rotator.name", cmd.Flags().Lookup("name"))
	viper.BindPFlag("rotator.has-azimuth", cmd.Flags().Lookup("has-azimuth"))
	viper.BindPFlag("rotator.has-elevation", cmd.Flags().Lookup("has-elevation"))
	viper.BindPFlag("rotator.pollingrate", cmd.Flags().Lookup("pollingrate"))

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
	go h.ListenTCP(viper.GetString("tcp.host"), viper.GetInt("tcp.port"), tcpError)

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
		}
	}

}
