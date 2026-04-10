package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func newMQTTConnection(
	uri string,
	username string,
	password string,
	topic string,
	qos byte,
	errorLogger *log.Logger,
	debugLogger *log.Logger,
	clientConfig paho.ClientConfig,
) (*autopaho.ConnectionManager, error) {
	mqttURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse mqtt uri %s: %w",
			uri,
			err,
		)
	}

	mqttConfig := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{mqttURL},
		ConnectUsername:               username,
		ConnectPassword:               []byte(password),
		KeepAlive:                     uint16(*mqttKeepAlive),
		CleanStartOnInitialConnection: *mqttCleanStart,
		SessionExpiryInterval:         uint32(*mqttSessionExpiryInterval),
		Errors:                        errorLogger,
		PahoErrors:                    errorLogger,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			instMQTTConnectionCounter.WithLabelValues(uri).Inc()
			slog.Info("Connected to MQTT server", "server", uri)

			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: topic, QoS: qos},
				},
			}); err != nil {
				slog.Error(
					"Failed to subscribe to MQTT topic",
					"topic",
					topic,
					"error",
					err,
				)
			}
			slog.Info("Subscribed to MQTT topic", "topic", topic)
		},
		OnConnectError: func(err error) {
			slog.Error("Error connecting to MQTT server", "uri", uri, "error", err)
		},
		ClientConfig: clientConfig,
	}

	if debugLogger != nil {
		mqttConfig.Debug = debugLogger
		mqttConfig.PahoDebug = debugLogger
	}

	mqttContext := context.Background()

	mqttConnection, err := autopaho.NewConnection(mqttContext, mqttConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create mqtt connection: %w", err)
	}

	return mqttConnection, nil
}
