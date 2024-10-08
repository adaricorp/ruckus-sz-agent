package main

import (
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
)

var (
	instMQTTConnectionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "mqtt_connections_total",
		},
		[]string{"server"},
	)
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
		[]string{"system_id", "message_type"},
	)
	instMessageErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "messages_errors_total",
		},
		[]string{"system_id", "message_type"},
	)
	instJSONUnparseableCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "json_messages_unparseable_total",
		},
		[]string{"system_id", "message_type"},
	)
	instUnhandledMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "messages_unhandled_total",
		},
		[]string{"system_id"},
	)
	instMetricErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: binName,
			Name:      "metric_errors_total",
		},
		[]string{"metric_name"},
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

	prometheus.MustRegister(instMQTTConnectionCounter)
	prometheus.MustRegister(instMQTTMessageCounter)
	prometheus.MustRegister(instMQTTBytesCounter)
	prometheus.MustRegister(instMQTTUnparseableMessageCounter)
	prometheus.MustRegister(instProcessedMessageCounter)
	prometheus.MustRegister(instMessageErrorCounter)
	prometheus.MustRegister(instJSONUnparseableCounter)
	prometheus.MustRegister(instUnhandledMessageCounter)
	prometheus.MustRegister(instMetricErrorCounter)
}
