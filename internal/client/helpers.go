package client

import (
	"github.com/phoenix-mstu/go-modifying-mqtt-proxy/internal/subscriptions"
)

func filterForClient(filters []subscriptions.Filter, subs []subscription, topic string, payload []byte) (string, []byte, byte, bool) {
	for _, filter := range filters {
		fTopic, fPayload, ok := filter.Apply(topic, payload)
		if !ok {
			continue
		}
		for _, subscription := range subs {
			if subscriptions.RouteIncludesTopic(subscription.route, fTopic) {
				return fTopic, fPayload, subscription.qos, true
			}
		}
	}
	return "", []byte{}, 0, false
}

func filterForBroker(filters []subscriptions.Filter, topic string, payload []byte) (string, []byte, bool) {
	for _, filter := range filters {
		if fTopic, fPayload, ok := filter.Apply(topic, payload); ok {
			return fTopic, fPayload, true
		}
	}
	return "", []byte{}, false
}

func min(a, b byte) byte {
	if a <= b {
		return a
	}
	return b
}