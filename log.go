package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/loki/pkg/push"
	"github.com/grafana/loki/v3/pkg/logproto"
)

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
) logEntry {
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

	return logEntry{
		labels:    "{" + strings.Join(l, ",") + "}",
		metadata:  m,
		timestamp: ts,
		message:   message,
	}
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
