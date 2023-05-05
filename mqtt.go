package main

import (
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

func mqttInit(host string, port uint16, username, password string) (mqtt.Client, error) {
	broker := fmt.Sprintf("tcp://%s:%d", host, port)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetUsername(username)
	opts.SetPassword(password)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	return client, nil
}

func mqttPublish(client mqtt.Client, topic string, payload []byte) {
	pubClient := client
	pubClient.Publish(topic, 0, false, string(payload))
}
