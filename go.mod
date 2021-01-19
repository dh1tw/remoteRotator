module github.com/dh1tw/remoteRotator

go 1.15

require (
	github.com/GeertJohan/go.rice v1.0.2
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/micro/go-micro/plugins/broker/nats/v2 v2.0.0-20210105173217-bf4ab679e18b
	github.com/micro/go-micro/plugins/registry/nats/v2 v2.0.0-20210105173217-bf4ab679e18b
	github.com/micro/go-micro/plugins/transport/nats/v2 v2.0.0-20210105173217-bf4ab679e18b
	github.com/micro/go-micro/v2 v2.9.2-0.20201226154210-35d72660c801
	github.com/micro/mdns v0.3.0
	github.com/nats-io/nats.go v1.10.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	google.golang.org/protobuf v1.23.0
)
