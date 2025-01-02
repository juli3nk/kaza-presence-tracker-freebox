package mqtt

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	client mqtt.Client
}

func New() *Mqtt {
	return new(Mqtt)
}

func (m *Mqtt) Init(host string, port uint16, username, password string) error {
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
		return err
	}

	m.client = client

	return nil
}

func (m *Mqtt) Publish(topic string, payload []byte) error {
	// pubClient := client
	token := m.client.Publish(topic, 0, false, string(payload))
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return err
	}

	return nil
}

func (m *Mqtt) IsConnected() bool {
	return m.client.IsConnected()
}

func (m *Mqtt) Disconnect(quiesce uint) {
	m.client.Disconnect(quiesce)
}
