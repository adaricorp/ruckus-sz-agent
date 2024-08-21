package main

import (
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
)

var (
	instMqttMessageCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_messages_total",
		},
	)
	instMqttBytesCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_bytes_total",
		},
	)
	instMqttUnparseableMessageCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_messages_unparseable_total",
		},
	)
	instProcessedMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "messages_processed_total",
		},
		[]string{"message_type"},
	)
	instUnparseableMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "messages_unparseable_total",
		},
		[]string{"message_type"},
	)
	instUnhandledMessageCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "messages_unhandled_total",
		},
	)
)

func registerInstrumentationMetrics() {
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "up",
			},
			func() float64 { return 1 },
		),
	)

	versionCollector := versioncollector.NewCollector(binName)
	prometheus.MustRegister(versionCollector)

	prometheus.MustRegister(instMqttMessageCounter)
	prometheus.MustRegister(instMqttBytesCounter)
	prometheus.MustRegister(instMqttUnparseableMessageCounter)
	prometheus.MustRegister(instProcessedMessageCounter)
	prometheus.MustRegister(instUnparseableMessageCounter)
	prometheus.MustRegister(instUnhandledMessageCounter)
}
