package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/golang/snappy"
	dto "github.com/prometheus/client_model/go"
	prom_config "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/prometheus/prometheus/util/fmtutil"
)

type prometheusWriteClient struct {
	client remote.WriteClient
}

func newPrometheusWriteClient(uri string, timeout time.Duration) (prometheusWriteClient, error) {
	promURI, err := url.Parse(uri)
	if err != nil {
		return prometheusWriteClient{}, fmt.Errorf(
			"failed to parse prometheus remote write uri %s: %w",
			uri,
			err,
		)
	}

	client, err := remote.NewWriteClient(binName, &remote.ClientConfig{
		URL:     &prom_config.URL{URL: promURI},
		Timeout: model.Duration(timeout),
	})
	if err != nil {
		return prometheusWriteClient{}, fmt.Errorf(
			"failed to create prometheus remote write client: %w", err,
		)
	}

	return prometheusWriteClient{
		client: client,
	}, nil
}

func (p prometheusWriteClient) write(metrics map[string]*dto.MetricFamily, labels map[string]string) error {
	writeRequest, err := fmtutil.MetricFamiliesToWriteRequest(metrics, labels)
	if err != nil {
		return fmt.Errorf("unable to format write request: %w", err)
	}

	rawRequest, err := writeRequest.Marshal()
	if err != nil {
		return fmt.Errorf("unable to marshal write request: %w", err)
	}

	compressedRequest := snappy.Encode(nil, rawRequest)

	ctx := context.Background()

	_, err = p.client.Store(ctx, compressedRequest, 0)
	if err != nil {
		return fmt.Errorf("unable to send write request to prometheus: %w", err)
	}

	return nil
}
