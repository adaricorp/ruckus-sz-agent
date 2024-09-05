package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/loki/pkg/push"
	"github.com/grafana/loki/v3/pkg/logproto"
)

func checkTimestamp(t time.Time) error {
	now := time.Now()
	upperLimit := now.Add(*lokiTimestampFuture).Unix()
	lowerLimit := now.Add(*lokiTimestampPast * -1).Unix()

	if t.Unix() >= upperLimit {
		return fmt.Errorf("timestamp is too far in the future")
	}
	if t.Unix() <= lowerLimit {
		return fmt.Errorf("timestamp is too far in the past")
	}

	return nil
}

type logEntry struct {
	labels    string
	metadata  push.LabelsAdapter
	timestamp time.Time
	message   string
}

func newLogEntry(
	labels map[string]string,
	labelOrder []string,
	metadata map[string]string,
	ts time.Time,
	message string,
) (logEntry, error) {
	l := []string{}
	for _, k := range labelOrder {
		v := labels[k]
		if v == "" {
			continue
		}
		l = append(l, fmt.Sprintf(`%s="%s"`, k, v))
	}

	m := push.LabelsAdapter{}
	for k, v := range metadata {
		if v == "" {
			continue
		}
		m = append(m, push.LabelAdapter{
			Name:  k,
			Value: v,
		})
	}

	e := logEntry{
		labels:    "{" + strings.Join(l, ",") + "}",
		metadata:  m,
		timestamp: ts,
		message:   message,
	}

	if err := checkTimestamp(ts); err != nil {
		return e, err
	}

	return e, nil
}

func (e logEntry) String() string {
	return fmt.Sprintf(
		`timestamp='%s' labels='%s' metadata='%s' message='%s'`,
		e.timestamp, e.labels, e.metadata, e.message,
	)
}

func (e logEntry) marshal() logproto.Stream {
	return logproto.Stream{
		Labels: e.labels,
		Entries: []logproto.Entry{
			{
				Timestamp:          e.timestamp,
				Line:               e.message,
				StructuredMetadata: e.metadata,
			},
		},
	}
}
