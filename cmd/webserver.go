package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/client/selector/static"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport"
	natsBroker "github.com/micro/go-plugins/broker/nats"
	natsReg "github.com/micro/go-plugins/registry/nats"
	natsTr "github.com/micro/go-plugins/transport/nats"
	nats "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dh1tw/remoteRotator/discovery"
	"github.com/dh1tw/remoteRotator/hub"
	"github.com/dh1tw/remoteRotator/rotator"
	"github.com/dh1tw/remoteRotator/rotator/proxy"
	sbProxy "github.com/dh1tw/remoteRotator/rotator/sb_proxy"
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

	connClosed := make(chan struct{})

	sbTransport := strings.ToLower(viper.GetString("shackbus.transport"))

	if sbTransport == "nats" {
		url := viper.GetString("nats.broker-url")
		port := viper.GetInt("nats.broker-port")
		username := viper.GetString("nats.username")
		password := viper.GetString("nats.password")

		nopts := nats.GetDefaultOptions()
		nopts.Servers = []string{fmt.Sprintf("%s:%d", url, port)}
		nopts.User = username
		nopts.Password = password
		nopts.Timeout = time.Second * 10

		disconnectedHdlr := func(conn *nats.Conn) {
			log.Println("connection to nats broker closed")
			connClosed <- struct{}{}
		}

		reconnectHdlr := func(conn *nats.Conn) {
			log.Println("reconnected to nats broker")
		}
		nopts.ReconnectedCB = reconnectHdlr

		errorHdlr := func(conn *nats.Conn, sub *nats.Subscription, err error) {
			log.Printf("Error Handler called (%s): %s", sub.Subject, err)
		}
		nopts.AsyncErrorCB = errorHdlr

		regNatsOpts := nopts
		brNatsOpts := nopts
		trNatsOpts := nopts
		regNatsOpts.DisconnectedCB = disconnectedHdlr
		regNatsOpts.Name = "remoteRotator.web:registry"
		brNatsOpts.Name = "remoteRotator.web:broker"
		trNatsOpts.Name = "remoteRotator.web:transport"

		regTimeout := registry.Timeout(time.Second * 2)
		trTimeout := transport.Timeout(time.Second * 2)

		reg = natsReg.NewRegistry(natsReg.Options(regNatsOpts), regTimeout)
		tr = natsTr.NewTransport(natsTr.Options(trNatsOpts), trTimeout)
		br = natsBroker.NewBroker(natsBroker.Options(brNatsOpts))
		cl = client.NewClient(
			client.Broker(br),
			client.Transport(tr),
			client.Registry(reg),
			client.PoolSize(1),
			client.PoolTTL(time.Hour*8760), // one year - don't TTL our connection
			client.Selector(static.NewSelector()),
		)

		if err := cl.Init(); err != nil {
			log.Println(err)
			return
		}
	}

	cache := &serviceCache{
		ttl:   time.Second * 10,
		cache: make(map[string]time.Time),
	}
	w := webserver{h, cl, cache}

	// will be closed when an error occures in the webserver goroutine
	webserverErrorCh := make(chan struct{})

	// launch webserver
	go w.ListenHTTP(viper.GetString("web.host"), viper.GetInt("web.port"), webserverErrorCh)

	// watch the registry in a seperate thread for changes
	if sbTransport == "nats" {
		// at startup query the registry and add all found rotators
		if err := w.listAndAddRotators(); err != nil {
			log.Println(err)
		}
		// from now on watch the registry in a separate go-routine for changes
		go w.watchRegistry()
		// check regularily if the proxy objects are still alive
		go w.checkTimeout()
	}

	// ticker for check lan registry
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
		case <-connClosed:
			rotators := w.Rotators()
			for _, r := range rotators {
				r.Close()
			}
		case <-webserverErrorCh:
			fmt.Println("web server crashed")
			return
		case <-ticker.C:
			switch sbTransport {
			case "lan":
				go w.update()
			}
		case u := <-bcast:
			ev := hub.Event{
				Name:        hub.UpdateHeading,
				RotatorName: u.rotatorName, //MISSING
				Heading:     u.heading,
			}
			w.BroadcastToWsClients(ev)
		}
	}
}

type serviceCache struct {
	sync.Mutex
	ttl   time.Duration
	cache map[string]time.Time
}

type webserver struct {
	*hub.Hub
	cli   client.Client
	cache *serviceCache
}

type update struct {
	rotatorName string
	heading     rotator.Heading
}

var bcast = make(chan update, 10)

var ev = func(r rotator.Rotator, heading rotator.Heading) {
	bcast <- update{r.Name(), heading}
}

//extract the service's name from its fully qualified service name (FQSN)
func nameFromFQSN(serviceName string) string {
	splitted := strings.Split(serviceName, ".")
	name := splitted[len(splitted)-1]
	return strings.Replace(name, "_", " ", -1)
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
	serviceName := sbProxy.ServiceName(strings.Replace(rotatorServiceName, " ", "_", -1))

	// create new rotator proxy object
	r, err := sbProxy.New(done, cli, eh, name, serviceName)
	if err != nil {
		close(doneCh)
		return fmt.Errorf("unable to create proxy object: %v", err)
	}

	if err := w.AddRotator(r); err != nil {
		close(doneCh)
		return fmt.Errorf("unable to add proxy objects: %v", err)
	}

	go func() {
		<-doneCh
		// fmt.Println("disposing:", r.Name())
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
		if !isRotator(service.Name) {
			continue
		}
		if err := w.addRotator(service.Name); err != nil {
			log.Println(err)
		}
	}

	return nil
}

// isRotator checks a serviceName string if it is a shackbus rotator
func isRotator(serviceName string) bool {

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
	}

	for {
		res, err := watcher.Next()
		if err != nil {
			log.Println("watch error:", err)
		}

		if !isRotator(res.Service.Name) {
			continue
		}

		switch res.Action {

		case "create", "update":
			if err := w.addRotator(res.Service.Name); err != nil {
				log.Println(err)
			}
			w.cache.Lock()
			w.cache.cache[res.Service.Name] = time.Now()
			w.cache.Unlock()

		case "delete":
			rotatorName := nameFromFQSN(res.Service.Name)
			r, exists := w.Rotator(rotatorName)
			if !exists {
				continue
			}
			r.Close()

			w.cache.Lock()
			delete(w.cache.cache, res.Service.Name)
			w.cache.Unlock()
		}
	}
}

// used only for lan rotators / discovery
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

// checkTimeout checks every second if the existing proxy objects
// are still alive. Dead objects will be removed.
func (w *webserver) checkTimeout() {

	tick := time.Tick(time.Second)

	for {
		<-tick
		w.cache.Lock()
		for service, timeout := range w.cache.cache {
			if time.Since(timeout) >= w.cache.ttl {
				rotatorName := nameFromFQSN(service)
				r, exists := w.Rotator(rotatorName)
				if !exists {
					continue
				}
				r.Close()
				delete(w.cache.cache, service)
			}
		}
		w.cache.Unlock()
	}
}
