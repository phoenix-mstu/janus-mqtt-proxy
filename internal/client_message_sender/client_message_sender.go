package client_message_sender

import (
	"github.com/eclipse/paho.mqtt.golang/packets"
	"log"
	"net"
	"sync"
	"time"
)

const MessageDeliveryTimeout = 5 * time.Second
const MessageDeliveryAttempts = 3

type ClientMessageSender struct {
	sync.RWMutex
	wg 			      sync.WaitGroup
	conn          	  net.Conn
	cOutgoing         chan packets.ControlPacket
	sendingMessageIds map[uint16]chan byte // MessageType
	lastSentMessageId uint16
}

func NewClientMessageSender(conn net.Conn) *ClientMessageSender {
	cms := ClientMessageSender{
		RWMutex:           sync.RWMutex{},
		conn:              conn,
		cOutgoing:         make(chan packets.ControlPacket),
		sendingMessageIds: make(map[uint16]chan byte),
		lastSentMessageId: 0,
	}
	go cms.serveOutgoing()
	return &cms
}

func (cms *ClientMessageSender) SendPacket(packet packets.ControlPacket) {
	cms.cOutgoing <- packet
}

func (cms *ClientMessageSender) Publish(topic string, qos byte, retained bool, payload []byte) {
	go cms.publishToClient(topic, qos, retained, payload)
}

func (cms *ClientMessageSender) ProcessControlMessage(messageID uint16, messageType byte) {
	if channel, ok := cms.sendingMessageIds[messageID]; ok {
		channel <- messageType
	}
}

func (cms *ClientMessageSender) Close() {
	cms.Lock()
	close(cms.cOutgoing)
	for _, channel := range cms.sendingMessageIds {
		close(channel)
	}
	cms.Unlock()
	cms.wg.Wait()
}

func (cms *ClientMessageSender) serveOutgoing() {
	cms.wg.Add(1)
	defer cms.wg.Done()
	for packet := range cms.cOutgoing {
		if e := packet.Write(cms.conn); e != nil {
			log.Print("Error writing packet to client")
		}
	}
}

func (cms *ClientMessageSender) preparePublishSession() (uint16, chan byte) {
	cms.Lock()
	defer cms.Unlock()
	for {
		// searching for fee messageId
		cms.lastSentMessageId++
		_, found := cms.sendingMessageIds[cms.lastSentMessageId]
		// must be non zero my spec
		if !found && cms.lastSentMessageId != 0 {
			break
		}
	}
	cms.sendingMessageIds[cms.lastSentMessageId] = make(chan byte)
	return cms.lastSentMessageId, cms.sendingMessageIds[cms.lastSentMessageId]
}

func (cms *ClientMessageSender) finishPublishSession(messageId uint16) {
	cms.Lock()
	defer cms.Unlock()
	close(cms.sendingMessageIds[messageId])
	delete(cms.sendingMessageIds, messageId)
}

func (cms *ClientMessageSender) readByteFromChannel(messageId uint16, channel chan byte, timeout time.Duration) byte {
	select {
	case msg := <- channel:
		return msg
	case <-time.After(timeout):
		return 0
	}
}

func (cms *ClientMessageSender) publishToClient(topic string, qos byte, retained bool, payload []byte) {
	cms.wg.Add(1)
	defer cms.wg.Done()

	messageId, channel := cms.preparePublishSession()
	defer cms.finishPublishSession(messageId)

	packet := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	packet.MessageID = messageId
	packet.Payload = payload
	packet.TopicName = topic
	packet.Retain = retained
	packet.Qos = qos
	cms.cOutgoing <- packet

	switch qos {
	case 0:
		// for qos=0 just send message and quit
		return
	case 1:
		for i := 0; i < MessageDeliveryAttempts; i++ {
			MessageType := cms.readByteFromChannel(messageId, channel, MessageDeliveryTimeout)
			switch MessageType {
			case 0:
				continue
			case packets.Puback:
				break
			default:
				// todo log error
				break
			}
		}
	case 2:
		// todo implement qos=2
	}
}