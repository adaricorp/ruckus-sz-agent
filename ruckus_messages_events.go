package main

import (
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"
	"github.com/pkg/errors"
)

func handleEvent(systemID string, message *pb.EventMessage) error {
	labelMap := map[string]string{
		"service":    *lokiMetricId,
		"event_type": message.GetEventType(),
		"category":   message.GetMainCategory(),
		"severity":   message.GetSeverity(),
	}

	metadataMap := map[string]string{
		"system_id": systemID,

		"subcategory": message.GetSubCategory(),

		"reason":            message.GetReason(),
		"disconnect_reason": parseDisconnectReason(message.GetDisconnectReason()),

		"domain_name":   message.GetDomainName(),
		"zone_name":     message.GetZoneName(),
		"ap_group_name": message.GetApGroupName(),

		"ap_mac": parseMAC(message.GetApMac()),

		"client_mac": parseMAC(message.GetClientMac()),
	}

	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	// Label order is important for loki to generate a consistent key for stream
	labelOrder := []string{"service", "severity", "category", "event_type"}

	event, err := newLogEntry(
		labelMap,
		labelOrder,
		metadataMap,
		timestamp,
		message.GetDescription(),
	)
	if err != nil {
		return errors.Wrapf(err, "Error creating event (%s)", event)
	}

	if err := loki.write(event); err != nil {
		return errors.Wrapf(err, "Error writing event to loki (%s)", event)
	}

	return nil
}
