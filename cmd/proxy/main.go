package main

import (
	"github.com/phoenix-mstu/go-modifying-mqtt-proxy/internal/broker"
	"github.com/phoenix-mstu/go-modifying-mqtt-proxy/internal/client"
	"github.com/phoenix-mstu/go-modifying-mqtt-proxy/internal/config"
	"log"
	"net"
	"os"
)

func serveClient(conn net.Conn, configs []config.CompiledFiltersConfig, brokerConnection *broker.BrokerConnection)  {
	brokerClient := brokerConnection.Subscribe()
	defer brokerConnection.Unsubscribe(brokerClient)
	defer conn.Close()
	client.ServeClientConnection(conn, configs, brokerClient)
}

func main() {
	log.Print("Proxy started")

	configPath := os.Getenv("MQTT_PROXY_CONFIG_PATH")
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if configPath == "" {
		println("Usage: go-modifying-mqtt-proxy /path/to/c.yaml")
		os.Exit(0)
	}

	c := config.ReadConfigFile(configPath)

	listener, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		log.Printf("Can't start listener")
		os.Exit(-1)
	}

	brokerConnection := broker.StartBrokerConnection(c.BrokerHost, c.BrokerLogin, c.BrokerPassword)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			os.Exit(-1)
		}
		go serveClient(clientConn, c.Clients, brokerConnection)
	}
}
