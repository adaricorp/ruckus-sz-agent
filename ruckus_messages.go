package main

import (
	"log/slog"
	"strconv"
	"strings"
)

func parseChannelWidth(width uint32) int {
	switch width {
	case 0:
		return 20
	case 2:
		return 40
	case 4:
		return 80
	default:
		return int(width)
	}
}

func parseMAC(mac string) string {
	return strings.ToLower(mac)
}

func parseVLAN(vid int32) string {
	if vid == 0 {
		return ""
	}

	return strconv.Itoa(int(vid))
}

func parseAPConnectionStatus(status string) bool {
	return strings.ToLower(status) == "connect"
}

func parseUpDown(s string) bool {
	return strings.ToLower(s) == "up"
}

func parseLatLong(latlong string, idx int) string {
	if latlong == "" {
		return ""
	}

	l := strings.Split(latlong, ",")
	if len(l) == 2 {
		return l[idx]
	}

	slog.Error("Error parsing latlong", "latlong", latlong, "index", idx)

	return ""
}

func parseLatitude(latlong string) string {
	return parseLatLong(latlong, 0)
}

func parseLongitude(latlong string) string {
	return parseLatLong(latlong, 1)
}

func parsePhyLink(phyLink string, idx int) string {
	l := strings.Split(phyLink, " ")
	if len(l) <= 3 && idx < len(l) {
		return l[idx]
	}

	slog.Error("Error parsing phy link", "phylink", phyLink, "index", idx)

	return ""
}

func parseDisconnectReason(reasonCode string) string {
	if reason, exists := WiFiReasonCodes[reasonCode]; exists {
		return reason
	}

	return reasonCode
}

func parsePhyLinkStatus(phyLink string) string {
	return parsePhyLink(phyLink, 0)
}

func parsePhyLinkSpeed(phyLink string) string {
	return parsePhyLink(phyLink, 1)
}

func parsePhyLinkDuplex(phyLink string) string {
	return parsePhyLink(phyLink, 2)
}
