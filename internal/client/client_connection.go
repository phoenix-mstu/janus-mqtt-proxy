package client

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/phoenix-mstu/janus-mqtt-proxy/internal/broker"
	"github.com/phoenix-mstu/janus-mqtt-proxy/internal/client_message_sender"
	"github.com/phoenix-mstu/janus-mqtt-proxy/internal/config"
	"github.com/phoenix-mstu/janus-mqtt-proxy/internal/subscriptions"
	"log"
	"net"
)

type subscription struct {
	route string
	qos byte
}

type client struct {
	cms			  *client_message_sender.ClientMessageSender
	conn          net.Conn
	configs       []config.CompiledFiltersConfig
	brokerFilters []subscriptions.Filter
	clientFilters []subscriptions.Filter

	subscriptions []subscription
	brokerClient  *broker.Client
	isConnected bool
	username string
}

func ServeClientConnection(conn net.Conn, configs []config.CompiledFiltersConfig, brokerClient *broker.Client) {
	c := client{
		conn:         conn,
		configs:      configs,
		brokerClient: brokerClient,
		isConnected:  false,
		cms:          client_message_sender.NewClientMessageSender(conn),
	}
	go c.serveBroker()
	c.serveIncoming()
	c.cms.Close()
}

func (c *client) printf(format string, v ...interface{}) {
	log.Printf("[%s/%s] %s", c.conn.RemoteAddr().String(), c.username, fmt.Sprintf(format, v...))
}

func (c *client) serveBroker() {
	for msg := range c.brokerClient.Receive() {
		topic, payload, qos, ok := filterForClient(c.clientFilters, c.subscriptions, msg.Topic(), msg.Payload())
		if ok {
			c.cms.Publish(topic, min(qos, msg.Qos()), msg.Retained(), payload)
		}
	}
}

// https://www.hivemq.com/blog/mqtt-essentials-part-6-mqtt-quality-of-service-levels/
func (c *client) sendToBroker(MessageID uint16, topic string, qos byte, retained bool, payload interface{}) {
	if token := c.brokerClient.Send(topic, qos, retained, payload); !token.Wait() || token.Error() != nil {
		// some error during the send.
		// Just skip it, client must send the message again
		return
	}
	switch qos {
	case 0:
		// do nothing
	case 1:
		packet := packets.NewControlPacket(packets.Puback).(*packets.PubackPacket)
		packet.MessageID = MessageID
		c.cms.SendPacket(packet)
	case 2:
		// todo implement qos=2
	}
}

func (c *client) login(username, password string) bool {
	for _, conf := range c.configs {
		if conf.Login == username && conf.Password == password {
			c.isConnected = true
			c.username = username
			c.brokerFilters = conf.BrokerFilters
			c.clientFilters = conf.ClientFilters
			return true
		}
	}
	return false
}

func (c *client) subscribe(qos byte, topics []string) {
	var subs []subscription
	for _, topic := range topics {
		subs = append(subs, subscription{
			route: topic,
			qos:   qos,
		})
	}
	c.subscriptions = append(c.subscriptions, subs...)
	for _, msg := range c.brokerClient.GetRetained() {
		topic, payload, qos, ok := filterForClient(c.clientFilters, subs, msg.Topic(), msg.Payload())
		if ok {
			c.cms.Publish(topic, min(qos, msg.Qos()), msg.Retained(), payload)
		}
	}
}

func (c *client) unsubscribe(topics []string) {
	var result []subscription
	for _, subscription := range c.subscriptions {
		toDelete := false
		for _, topic := range topics {
			toDelete = toDelete || topic == subscription.route
		}
		if !toDelete {
			result = append(result, subscription)
		}
	}
	c.subscriptions = result
}

func (c *client) serveIncoming() {
	c.printf("accepted connection: %v\n", c.conn.RemoteAddr())
	for true {

		cp, err := packets.ReadPacket(c.conn)
		if err != nil {
			c.printf("Cant read packet")
			return
		}

		switch packet := cp.(type) {
		case *packets.ConnectPacket:
			c.printf("ConnectPacket, username: %v", packet.Username)
			response := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
			if c.login(packet.Username, string(packet.Password)) {
				c.printf("Connected")
				c.cms.SendPacket(response)
			} else {
				c.printf("Conn failed. Wrong auth")
				response.ReturnCode = 4 // bad password
				c.cms.SendPacket(response)
				return
			}
			continue
		default:
			if !c.isConnected {
				c.printf("Wrong packet, closing")
				return
			}
		}

		switch packet := cp.(type) {
		case *packets.SubscribePacket:
			c.printf("SubscribePacket %v", packet.Topics)
			response := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
			response.MessageID = packet.MessageID
			c.cms.SendPacket(response)
			c.subscribe(packet.Qos, packet.Topics)

		case *packets.UnsubscribePacket:
			c.printf("UnsubscribePacket")
			response := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
			response.MessageID = packet.MessageID
			c.cms.SendPacket(response)
			c.unsubscribe(packet.Topics)

		case *packets.PingreqPacket:
			c.printf("PingreqPacket")
			response := packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket)
			c.cms.SendPacket(response)

		case *packets.PublishPacket:
			c.printf("PublishPacket")
			if topic, payload, ok := filterForBroker(c.brokerFilters, packet.TopicName, packet.Payload); ok {
				go c.sendToBroker(packet.MessageID, topic, packet.Qos, packet.Retain, payload)
			}

		case *packets.PubackPacket:
			c.printf("Puback")
			c.cms.ProcessControlMessage(packet.MessageID, packets.Puback)

		case *packets.PubrelPacket:
			c.printf("Pubrel")
			c.cms.ProcessControlMessage(packet.MessageID, packets.Pubrel)

		case *packets.DisconnectPacket:
			c.printf("DisconnectPacket")
			return

		default:
			c.printf("Unexpected packet")
		}
	}
}
