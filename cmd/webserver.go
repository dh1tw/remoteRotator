package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/dh1tw/remoteRotator/discovery"
	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var webServerCmd = &cobra.Command{
	Use:   "web",
	Short: "webserver providing access to all rotators on the network",
	Long:  `webserver providing access to all rotators on the network`,
	Run:   webServer,
}

func init() {
	RootCmd.AddCommand(webServerCmd)
	webServerCmd.Flags().StringP("host", "w", "127.0.0.1", "Host (use '0.0.0.0' to listen on all network adapters)")
	webServerCmd.Flags().IntP("port", "k", 7000, "webserver http port")
}

func webServer(cmd *cobra.Command, args []string) {

	viper.BindPFlag("web.host", cmd.Flags().Lookup("host"))
	viper.BindPFlag("web.port", cmd.Flags().Lookup("port"))

	h, err := hub.NewHub()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	w := webserver{h}

	ticker := time.NewTicker(time.Second * 5)

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	done := make(chan bool)

	go w.ListenHTTP(viper.GetString("web.host"), viper.GetInt("web.port"), done)

	for {
		select {
		case sig := <-osSignals:
			if sig == os.Interrupt {
				return
			}
		case <-done:
			fmt.Println("web server crashed")
			return
		case <-ticker.C:
			go w.update()
		case s := <-bcast:
			ev := hub.Event{
				Name:   hub.UpdateHeading,
				Status: s,
			}
			w.BroadcastToWsClients(ev)
		}
	}
}

type webserver struct {
	*hub.Hub
}

var bcast = make(chan rotator.Status, 10)

var ev = func(r rotator.Rotator, ev rotator.Event, value ...interface{}) {
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

func (w *webserver) update() {

	dsvrdRotators, err := discovery.LookupRotators()
	if err != nil {
		log.Println(err)
		return
	}

	// check if rotator(s) are not registered yet
	for _, dr := range dsvrdRotators {

		// if the rotator is new, then add it
		if !w.HasRotator(dr.Name) {

			doneCh := make(chan struct{})
			done := proxy.DoneCh(doneCh)
			host := proxy.Host(dr.AddrV4.String())
			port := proxy.Port(dr.Port)
			eh := proxy.EventHandler(ev)
			r, err := proxy.New(done, host, port, eh)
			if err != nil {
				log.Println(err)
				continue
			}
			if err := w.AddRotator(r); err != nil {
				log.Println(err)
				continue
			}
			go func() {
				<-doneCh
				w.RemoveRotator(r)
			}()
		}
	}
}
