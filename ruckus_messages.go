package main

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"
)

type apTenant struct {
	tenantMessage *pb.TenantMessage

	domainMessages []*pb.DomainMessage
}

type switchTenant struct {
	tenantMessage *pb.IcxTenantMessage

	domains []switchDomain
}

type switchDomain struct {
	domainMessage *pb.IcxDomainMessage

	switchGroups []switchGroup
}

type switchGroup struct {
	switchGroupMessage *pb.SwitchGroupMessage

	subSwitchGroupMessages []*pb.SwitchGroupMessage
}

func parseIntegerString(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		slog.Error("Error parsing integer string", "string", s)
		return 0
	}

	return i
}

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

func parseOk(s string) bool {
	return strings.ToLower(s) == "ok"
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

func parseClusterState(s string) bool {
	return strings.ToLower(s) == "in_service"
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

func parseRegistrationStatus(status string) bool {
	return strings.ToLower(status) == "approved"
}

func parseSwitchStatus(status string) bool {
	return strings.ToLower(status) == "online"
}

func parseSTPStatus(status string) bool {
	return strings.ToLower(status) == "forwarding"
}

func parseLongPortName(port string) string {
	return shortPortNameRegex.FindString(port)
}

// Parse uptimes in the following formats:
//   - hh:mm:ss.ms
//   - dd days, hh:mm:ss.ms
func parseUptime(t string) int64 {
	var err error

	p := strings.Split(t, ", ")
	if len(p) != 1 && len(p) != 2 {
		slog.Error("Error parsing uptime", "time", t)

		return 0
	}

	// Parse days
	days := 0
	if len(p) == 2 {
		d := strings.Split(p[0], " ")
		if len(d) != 2 || !(d[1] == "days" || d[1] == "day") {
			slog.Error("Error parsing days in uptime", "time", t, "days", p[0])

			return 0
		}

		days, err = strconv.Atoi(d[0])
		if err != nil {
			slog.Error(
				"Error converting days in uptime to int",
				"error",
				err.Error(),
				"time",
				t,
				"days",
				d[0],
			)

			return 0
		}
	}

	duration := time.Duration(days) * 24 * time.Hour

	// Parse hours/minutes/seconds
	hours, err := time.Parse("15:04:05.00", p[len(p)-1])
	if err != nil {
		slog.Error(
			"Error parsing hours in uptime",
			"error",
			err.Error(),
			"time",
			t,
			"hours",
			p[len(p)-1],
		)

		return 0
	}

	z, _ := time.Parse("15:04:05.00", "00:00:00.00")

	duration += hours.Sub(z)

	return int64(duration / time.Second)
}
