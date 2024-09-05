package main

import (
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
)

var (
	instMQTTMessageCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_messages_total",
		},
	)
	instMQTTBytesCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_bytes_total",
		},
	)
	instMQTTUnparseableMessageCounter = prometheus.NewCounter(
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

	prometheus.MustRegister(instMQTTMessageCounter)
	prometheus.MustRegister(instMQTTBytesCounter)
	prometheus.MustRegister(instMQTTUnparseableMessageCounter)
	prometheus.MustRegister(instProcessedMessageCounter)
	prometheus.MustRegister(instUnparseableMessageCounter)
	prometheus.MustRegister(instUnhandledMessageCounter)
}
