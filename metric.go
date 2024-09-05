package main

import (
	"fmt"
	"maps"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
)

type promMetric struct {
	name      string
	timestamp time.Time
	labels    map[string]string
	value     float64
}

func newPromMetric(
	name string,
	ts time.Time,
	labels map[string]string,
	value interface{},
) (promMetric, error) {
	var f float64

	switch v := value.(type) {
	case int:
		f = float64(v)
	case int16:
		f = float64(v)
	case int32:
		f = float64(v)
	case int64:
		f = float64(v)
	case uint:
		f = float64(v)
	case uint16:
		f = float64(v)
	case uint32:
		f = float64(v)
	case uint64:
		f = float64(v)
	case bool:
		if v {
			f = float64(1)
		} else {
			f = float64(0)
		}
	case float64:
		f = v
	default:
		return promMetric{}, fmt.Errorf("Unknown type for metric %s", name)
	}

	return promMetric{
		name:      name,
		timestamp: ts,
		labels:    labels,
		value:     f,
	}, nil
}

func (m promMetric) marshalMetricFamily() *dto.MetricFamily {
	return &dto.MetricFamily{
		Name:   proto.String(m.name),
		Type:   dto.MetricType_UNTYPED.Enum(),
		Metric: []*dto.Metric{m.marshalMetric()},
	}
}

func (m promMetric) marshalMetric() *dto.Metric {
	labels := []*dto.LabelPair{}

	for k, v := range m.labels {
		labels = append(labels, &dto.LabelPair{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	return &dto.Metric{
		Label: labels,
		Untyped: &dto.Untyped{
			Value: proto.Float64(m.value),
		},
		TimestampMs: proto.Int64(m.timestamp.UnixMilli()),
	}
}

func appendMetrics(
	timestamp time.Time,
	metricMap map[string]interface{},
	labelMap map[string]map[string]string,
	metricsFamily map[string]*dto.MetricFamily,
) []error {
	errs := []error{}

	for k, v := range metricMap {
		labels := map[string]string{}

		// Add default labels if they exist
		maps.Copy(labels, labelMap["default"])

		// Add metric-specific labels if they exist
		maps.Copy(labels, labelMap[k])

		m, err := newPromMetric(k, timestamp, labels, v)
		if err != nil {
			instMetricErrorCounter.WithLabelValues(k).Inc()
			errs = append(errs, errors.Wrapf(err, "Error creating metric"))
			continue
		}

		if metricsFamily[k] == nil {
			metricsFamily[k] = m.marshalMetricFamily()
		} else {
			metricsFamily[k].Metric = append(metricsFamily[k].Metric, m.marshalMetric())
		}
	}

	if len(errs) >= 1 {
		return errs
	}

	return nil
}

func metricSliceToMap(metrics []*dto.MetricFamily) (map[string]*dto.MetricFamily, []error) {
	metricsFamily := map[string]*dto.MetricFamily{}

	errs := []error{}

	for _, metric := range metrics {
		if _, exists := metricsFamily[metric.GetName()]; exists {
			errs = append(
				errs,
				fmt.Errorf("Map already contains metric %s", metric.GetName()),
			)

		}
		metricsFamily[metric.GetName()] = metric
	}

	return metricsFamily, errs
}
