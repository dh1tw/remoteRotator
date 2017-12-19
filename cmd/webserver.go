package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dh1tw/remoteRotator/discovery"
	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/proxy"
	"github.com/dh1tw/remoteRotator/rotator/sb_proxy"
	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
	"github.com/micro/go-micro/selector/cache"
	"github.com/micro/go-micro/transport"
	natsBroker "github.com/micro/go-plugins/broker/nats"
	natsReg "github.com/micro/go-plugins/registry/nats"
	natsTr "github.com/micro/go-plugins/transport/nats"
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
	webServerCmd.Flags().StringP("station", "X", "mystation", "Your station callsign")
	webServerCmd.Flags().StringP("transport", "t", "nats", "shackbus transport protocol (nats/lan)")
	webServerCmd.Flags().StringP("broker-url", "u", "localhost", "Broker URL")
	webServerCmd.Flags().IntP("broker-port", "p", 4222, "Broker Port")
	webServerCmd.Flags().StringP("password", "P", "", "NATS Password")
	webServerCmd.Flags().StringP("username", "U", "", "NATS Username")
}

func webServer(cmd *cobra.Command, args []string) {

	viper.BindPFlag("web.host", cmd.Flags().Lookup("host"))
	viper.BindPFlag("web.port", cmd.Flags().Lookup("port"))
	viper.BindPFlag("shackbus.station", cmd.Flags().Lookup("station"))
	viper.BindPFlag("shackbus.transport", cmd.Flags().Lookup("transport"))
	viper.BindPFlag("nats.broker-url", cmd.Flags().Lookup("broker-url"))
	viper.BindPFlag("nats.broker-port", cmd.Flags().Lookup("broker-port"))
	viper.BindPFlag("nats.password", cmd.Flags().Lookup("password"))
	viper.BindPFlag("nats.username", cmd.Flags().Lookup("username"))

	h, err := hub.NewHub()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var reg registry.Registry
	var tr transport.Transport
	var br broker.Broker
	var cl client.Client

	sbTransport := strings.ToLower(viper.GetString("shackbus.transport"))

	if sbTransport == "nats" {
		url := viper.GetString("nats.broker-url")
		port := viper.GetInt("nats.broker-port")
		username := viper.GetString("nats.username")
		password := viper.GetString("nats.password")
		credentials := ""
		if len(username) > 0 && len(password) > 0 {
			credentials = fmt.Sprintf("%s:%s@", username, password)
		}
		addr := fmt.Sprintf("nats://%s%s:%v", credentials, url, port)

		regTimeout := registry.Timeout(time.Second * 2)

		reg = natsReg.NewRegistry(registry.Addrs(addr), regTimeout)
		tr = natsTr.NewTransport(transport.Addrs(addr))
		br = natsBroker.NewBroker(broker.Addrs(addr))
		cl = client.NewClient(
			client.Broker(br),
			client.Transport(tr),
			client.Registry(reg),
			client.PoolSize(2),
			client.Selector(cache.NewSelector(selector.Registry(reg))),
		)

		// connect to broker
		if err := br.Init(); err != nil {
			fmt.Println(err)
			return
		}

	}

	w := webserver{h, cl, strings.ToLower(viper.GetString("shackbus.station"))}

	if sbTransport == "nats" {
		go w.updateMicro()
	}

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
			switch sbTransport {
			case "lan":
				go w.update()
				// case "nats":
				// 	go w.updateMicro()
			}
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
	cli     client.Client
	station string
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

// func (w *webserver) updateMicro() {
// 	services, err := w.cli.Options().Registry.ListServices()
// 	if err != nil {
// 		log.Println(err)
// 		os.Exit(1)
// 		return
// 	}
// 	for _, service := range services {

// 		if !strings.Contains(service.Name, w.station) {
// 			continue
// 		}

// 		if !strings.Contains(service.Name, "shackbus.rotator.") {
// 			continue
// 		}

// 		splitted := strings.Split(service.Name, ".")
// 		rotatorName := splitted[len(splitted)-1]

// 		if !w.HasRotator(rotatorName) {
// 			doneCh := make(chan struct{})
// 			done := sbProxy.DoneCh(doneCh)
// 			cli := sbProxy.Client(w.cli)
// 			eh := sbProxy.EventHandler(ev)
// 			name := sbProxy.Name(rotatorName)
// 			serviceName := sbProxy.ServiceName(service.Name)
// 			r, err := sbProxy.New(done, cli, eh, name, serviceName)
// 			if err != nil {
// 				log.Println("unable to create shackbus proxy object:", err)
// 				r = nil
// 				continue
// 			}
// 			if err := w.AddRotator(r); err != nil {
// 				log.Println(err)
// 				continue
// 			}
// 			go func() {
// 				<-doneCh
// 				w.RemoveRotator(r)
// 			}()
// 		}
// 	}
// }

func (w *webserver) updateMicro() {
	watcher, err := w.cli.Options().Registry.Watch()
	if err != nil {
		log.Println(err)
		os.Exit(1)
		return
	}

	for {
		res, err := watcher.Next()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			continue
		}

		if res.Action != "create" {
			continue
		}

		if !strings.Contains(res.Service.Name, w.station) {
			continue
		}

		if !strings.Contains(res.Service.Name, "shackbus.rotator.") {
			continue
		}

		splitted := strings.Split(res.Service.Name, ".")
		rotatorName := splitted[len(splitted)-1]

		if !w.HasRotator(rotatorName) {
			doneCh := make(chan struct{})
			done := sbProxy.DoneCh(doneCh)
			cli := sbProxy.Client(w.cli)
			eh := sbProxy.EventHandler(ev)
			name := sbProxy.Name(rotatorName)
			serviceName := sbProxy.ServiceName(res.Service.Name)
			r, err := sbProxy.New(done, cli, eh, name, serviceName)
			if err != nil {
				log.Println("unable to create shackbus proxy object:", err)
				r = nil
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
				log.Println("unable to create proxy object:", err)
				r = nil
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
