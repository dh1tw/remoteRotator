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

// func neverRetry(ctx context.Context, req client.Request, retryCount int, err error) (bool, error) {
// 	return false, nil
// }

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
		trTimeout := transport.Timeout(time.Second * 2)

		reg = natsReg.NewRegistry(registry.Addrs(addr), regTimeout)
		tr = natsTr.NewTransport(transport.Addrs(addr), trTimeout)
		br = natsBroker.NewBroker(broker.Addrs(addr))
		cl = client.NewClient(
			client.Broker(br),
			client.Transport(tr),
			client.Registry(reg),
			client.PoolSize(2),
			// client.Retry(neverRetry),
			client.Selector(cache.NewSelector(selector.Registry(reg))),
		)

		// connect to broker
		if err := br.Init(); err != nil {
			fmt.Println(err)
			return
		}

	}

	w := webserver{h, cl, strings.ToLower(viper.GetString("shackbus.station"))}

	// will be closed when an error occures in the webserver goroutine
	webserverErrorCh := make(chan struct{})

	go w.ListenHTTP(viper.GetString("web.host"), viper.GetInt("web.port"), webserverErrorCh)

	// at startup query the registry and add all existing rotators
	if err := w.listAndAddRotators(); err != nil {
		log.Println(err)
	}

	// watch the registry in a seperate thread for changes
	if sbTransport == "nats" {
		go w.watchRegistry()
	}

	ticker := time.NewTicker(time.Second * 5)

	// Channel to handle OS signals
	osSignals := make(chan os.Signal, 1)

	//subscribe to os.Interrupt (CTRL-C signal)
	signal.Notify(osSignals, os.Interrupt)

	for {
		select {
		case sig := <-osSignals:
			if sig == os.Interrupt {
				return
			}
		case <-webserverErrorCh:
			fmt.Println("web server crashed")
			return
		case <-ticker.C:
			switch sbTransport {
			case "lan":
				go w.update()
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

//extract the service's name from its fully qualified service name (FQSN)
func nameFromFQSN(serviceName string) string {
	splitted := strings.Split(serviceName, ".")
	return splitted[len(splitted)-1]
}

func (w *webserver) addRotator(rotatorServiceName string) error {

	rotatorName := nameFromFQSN(rotatorServiceName)

	// only continue if this rotator(name) does not exist yet
	_, exists := w.Rotator(rotatorName)
	if exists {
		return nil
	}

	doneCh := make(chan struct{})

	done := sbProxy.DoneCh(doneCh)
	cli := sbProxy.Client(w.cli)
	eh := sbProxy.EventHandler(ev)
	name := sbProxy.Name(rotatorName)
	serviceName := sbProxy.ServiceName(rotatorServiceName)

	// create new rotator proxy object
	r, err := sbProxy.New(done, cli, eh, name, serviceName)
	if err != nil {
		return fmt.Errorf("unable to create proxy object: %v", err)
	}

	if err := w.AddRotator(r); err != nil {
		return fmt.Errorf("unable to add proxy objects: %v", err)
	}

	go func() {
		<-doneCh
		w.RemoveRotator(r)
	}()

	return nil
}

// listAndAddRotators is a convenience function which queries the
// registry for all rotator services and then add proxy objects for
// each of them.
func (w *webserver) listAndAddRotators() error {

	services, err := w.cli.Options().Registry.ListServices()
	if err != nil {
		return err
	}

	for _, service := range services {
		fmt.Println("found:", service.Name)
		if !isRotator(service.Name, w.station) {
			continue
		}
		if err := w.addRotator(service.Name); err != nil {
			log.Println(err)
		}
	}

	return nil
}

// isRotator checks a serviceName string if it is a shackbus
// rotator for the selected station
func isRotator(serviceName, station string) bool {

	if !strings.Contains(serviceName, station) {
		return false
	}

	if !strings.Contains(serviceName, "shackbus.rotator.") {
		return false
	}
	return true
}

// watchRegistry is a blocking function which continously
// checks the registry for changes (new rotators being added/updated/removed).
func (w *webserver) watchRegistry() {
	watcher, err := w.cli.Options().Registry.Watch()
	if err != nil {
		log.Println(err)
		os.Exit(1)
		return
	}

	for {
		res, err := watcher.Next()
		if err != nil {
			// in case of a timeout (which most likely is a disconnect)
			// close the application
			fmt.Println(err)
			os.Exit(1)
		}

		if !isRotator(res.Service.Name, w.station) {
			continue
		}

		if res.Action == "create" {
			if err := w.addRotator(res.Service.Name); err != nil {
				log.Println(err)
			}
		}

		if res.Action == "delete" {
			rotatorName := nameFromFQSN(res.Service.Name)
			r, exists := w.Rotator(rotatorName)
			if !exists {
				continue
			}
			r.Close()
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

		// only add when the rotator is not registed yet
		_, exists := w.Rotator(dr.Name)
		if exists {
			continue
		}

		doneCh := make(chan struct{})
		done := proxy.DoneCh(doneCh)
		host := proxy.Host(dr.AddrV4.String())
		port := proxy.Port(dr.Port)
		eh := proxy.EventHandler(ev)
		r, err := proxy.New(done, host, port, eh)
		if err != nil {
			log.Println("unable to create proxy object:", err)
			r.Close()
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
