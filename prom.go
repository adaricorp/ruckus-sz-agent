package main

import (
	"context"
	"net/url"
	"time"

	"github.com/golang/snappy"
	"github.com/pkg/errors"
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
		return prometheusWriteClient{}, errors.Wrapf(
			err,
			"Failed to parse prometheus remote write uri: %s",
			uri,
		)
	}

	client, err := remote.NewWriteClient(binName, &remote.ClientConfig{
		URL:     &prom_config.URL{URL: promURI},
		Timeout: model.Duration(timeout),
	})
	if err != nil {
		return prometheusWriteClient{}, errors.Wrapf(
			err,
			"Failed to create prometheus remote write client",
		)
	}

	return prometheusWriteClient{
		client: client,
	}, nil
}

func (p prometheusWriteClient) write(metrics map[string]*dto.MetricFamily, labels map[string]string) error {
	writeRequest, err := fmtutil.MetricFamiliesToWriteRequest(metrics, labels)
	if err != nil {
		return errors.Wrapf(err, "Unable to format write request")
	}

	rawRequest, err := writeRequest.Marshal()
	if err != nil {
		return errors.Wrapf(err, "Unable to marshal write request")
	}

	compressedRequest := snappy.Encode(nil, rawRequest)

	ctx := context.Background()

	err = p.client.Store(ctx, compressedRequest, 0)
	if err != nil {
		return errors.Wrapf(err, "Unable to send write request to prometheus")
	}

	return nil
}
