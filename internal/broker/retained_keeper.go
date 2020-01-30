package broker

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

type RetainedKeeper struct {
	sync.RWMutex
	retainedMessages map[string]mqtt.Message
}

func newRetainedKeeper() *RetainedKeeper {
	return &RetainedKeeper{
		RWMutex:          sync.RWMutex{},
		retainedMessages: make(map[string]mqtt.Message),
	}
}

func (keeper *RetainedKeeper) addMessage(msg mqtt.Message) {
	keeper.Lock()
	defer keeper.Unlock()
	if msg.Retained() {
		keeper.retainedMessages[msg.Topic()] = msg
	} else {
		delete(keeper.retainedMessages, msg.Topic())
	}
}

func (keeper *RetainedKeeper) GetAll() []mqtt.Message {
	keeper.RLock()
	defer keeper.RUnlock()
	res := make([]mqtt.Message, 0, len(keeper.retainedMessages))
	for _, msg := range keeper.retainedMessages {
		res = append(res, msg)
	}
	return res
}