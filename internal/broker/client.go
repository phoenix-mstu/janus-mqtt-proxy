package broker

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	broker mqtt.Client
	cOut chan mqtt.Message
	retained *RetainedKeeper
}

func (c *Client) Send(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	return c.broker.Publish(topic, qos, retained, payload)
}

func (c *Client) Receive() <-chan mqtt.Message {
	return c.cOut
}

func (c *Client) GetRetained() []mqtt.Message {
	return c.retained.GetAll()
}