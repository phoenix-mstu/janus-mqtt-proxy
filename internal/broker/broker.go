package broker

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"sync"
)

type BrokerConnection struct {
	sync.RWMutex
	broker mqtt.Client
	clients []Client
	retainedKeeper *RetainedKeeper
}

func StartBrokerConnection(host, login, password string) *BrokerConnection {
	srvCon := BrokerConnection{
		RWMutex: sync.RWMutex{},
		retainedKeeper: newRetainedKeeper(),
	}
	srvCon.broker = mqtt.NewClient(mqtt.NewClientOptions().
		AddBroker(host).
		SetUsername(login).
		SetPassword(password).
		SetDefaultPublishHandler(srvCon.messageHandler).
		SetOnConnectHandler(func(client mqtt.Client) {
			if token := srvCon.broker.Subscribe("#", 0, nil); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
				os.Exit(1)
			}
		}).
		SetOrderMatters(true))

	if token := srvCon.broker.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return &srvCon
}

func (srvCon *BrokerConnection) Subscribe() *Client{
	srvCon.Lock()
	defer srvCon.Unlock()
	client := Client{
		broker:  srvCon.broker,
		cOut:    make(chan mqtt.Message),
		retained: srvCon.retainedKeeper,
	}
	srvCon.clients = append(srvCon.clients, client)
	return &client
}

func (srvCon *BrokerConnection) Unsubscribe(client *Client) {
	srvCon.Lock()
	defer srvCon.Unlock()
	for i, c := range srvCon.clients {
		if c == *client {
			lastId := len(srvCon.clients)-1
			srvCon.clients[i] = srvCon.clients[lastId]
			srvCon.clients = srvCon.clients[:lastId]
			break
		}
	}
}

func (srvCon *BrokerConnection) messageHandler(client mqtt.Client, msg mqtt.Message) {
	srvCon.Lock()
	defer srvCon.Unlock()
	srvCon.retainedKeeper.addMessage(msg)
	for _, client := range srvCon.clients {
		select {
		case client.cOut <- msg: // write if someone listens
		default:
		}
	}
}