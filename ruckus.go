package main

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"
	"github.com/pkg/errors"

	dto "github.com/prometheus/client_model/go"
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

		"zone_name": message.GetZoneName(),

		"ap_group_name": message.GetApGroupName(),
		"ap_mac":        parseMAC(message.GetApMac()),

		"client_mac": parseMAC(message.GetClientMac()),
	}

	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	// Label order is important for loki to generate a consistent key for stream
	labelOrder := []string{"service", "severity", "category", "event_type"}

	event := newLogEntry(
		labelMap,
		labelOrder,
		metadataMap,
		timestamp,
		message.GetDescription(),
	)

	if err := loki.write(event); err != nil {
		return errors.Wrapf(err, "Error writing event to loki")
	}

	return nil
}

func handleApStatus(systemID string, message *pb.APStatus) error {
	apStatusData := message.GetApStatusData()
	apSystem := apStatusData.GetAPSystem()

	timestamp := time.Unix(int64(message.GetSampleTime()), 0)

	apInfoLabels := map[string]string{
		"ap_group_name": message.GetApgroupName(),
		"ipv4_address":  apSystem.GetIp(),
		"ipv6_address":  apSystem.GetIpv6(),
		"fw_version":    apSystem.GetFwVersion(),
		"location":      apSystem.GetLocation(),
		"model":         apSystem.GetModel(),
		"serial_number": apSystem.GetSerialNumber(),
	}

	apMetrics := map[string]interface{}{
		"ruckus_ap_boots_total":                     apSystem.GetTotalBootCount(),
		"ruckus_ap_rejoins_total":                   apSystem.GetRejoinCount(),
		"ruckus_ap_controller_disconnections_total": apSystem.GetLossConnectBootCnt(),
		"ruckus_ap_uptime_seconds_total":            apSystem.GetUptime(),
		"ruckus_ap_current_temperature":             apSystem.GetCurrentTemperature(),
		"ruckus_ap_state":                           int(apSystem.GetApState()),
		"ruckus_ap_clients_connected":               apSystem.GetTotalConnectedClient(),
		"ruckus_ap_memory_used_percentage":          100 - apSystem.GetFreeMemoryPercentage(),
		"ruckus_ap_storage_used_percentage":         100 - apSystem.GetFreeStoragePercentage(),
		"ruckus_ap_cpu_used_percentage":             apSystem.GetCpuPercentage(),
		"ruckus_ap_status":                          apSystem.GetApStatus(),
		"ruckus_ap_lan_rx_bytes_total":              apSystem.GetLanStatsRxBytes(),
		"ruckus_ap_lan_tx_bytes_total":              apSystem.GetLanStatsTxBytes(),
		"ruckus_ap_lan_rx_packets_total":            apSystem.GetLanStatsRxPkts(),
		"ruckus_ap_lan_tx_packets_total":            apSystem.GetLanStatsTxPkts(),
		"ruckus_ap_lan_rx_error_packets_total":      apSystem.GetLanStatsRxErrorPkts(),
		"ruckus_ap_lan_tx_error_packets_total":      apSystem.GetLanStatsTxErrorPkts(),
		"ruckus_ap_lan_rx_drop_packets_total":       apSystem.GetLanStatsRxDroppedPkts(),
		"ruckus_ap_lan_tx_drop_packets_total":       apSystem.GetLanStatsTxDroppedPkts(),

		"ruckus_ap_info": 1,
	}

	labelMap := map[string]map[string]string{
		"ruckus_ap_info": apInfoLabels,
		"default":        {},
	}

	metricsFamily := map[string]*dto.MetricFamily{}

	if errs := appendMetrics(timestamp, apMetrics, labelMap, metricsFamily); len(errs) >= 1 {
		for _, err := range errs {
			slog.Error("Error while appending metrics", "error", err)
		}
	}

	for _, radio := range apStatusData.GetAPRadio() {
		radioLabels := map[string]string{
			"radio_id": strconv.Itoa(int(radio.GetRadioId())),
		}

		radioInfoLabels := map[string]string{
			"band":       radio.GetBand(),
			"radio_mode": radio.GetRadioMode(),
			"tx_power":   radio.GetTxPower(),
		}

		radioMetrics := map[string]interface{}{
			"ruckus_radio_channel":                               radio.GetChannel(),
			"ruckus_radio_rx_phy_errors_total":                   radio.GetPhyError(),
			"ruckus_radio_noise_floor_dbm":                       radio.GetNoiseFloor(),
			"ruckus_radio_rx_bytes_total":                        radio.GetRxBytes(),
			"ruckus_radio_tx_bytes_total":                        radio.GetTxBytes(),
			"ruckus_radio_rx_packets_total":                      radio.GetRxFrames(),
			"ruckus_radio_tx_packets_total":                      radio.GetTxFrames(),
			"ruckus_radio_retries_total":                         radio.GetRetry(),
			"ruckus_radio_drops_total":                           radio.GetDrop(),
			"ruckus_radio_channel_utilization_percentage":        radio.GetTotal(),
			"ruckus_radio_channel_busy_time_percentage":          radio.GetBusy(),
			"ruckus_radio_rx_channel_availability_percentage":    radio.GetRx(),
			"ruckus_radio_tx_channel_availability_percentage":    radio.GetTx(),
			"ruckus_radio_channel_width_mhz":                     parseChannelWidth(radio.GetChannelWidth()),
			"ruckus_radio_latency_microseconds":                  radio.GetLatency(),
			"ruckus_radio_capacity_megabits":                     radio.GetCapacity(),
			"ruckus_radio_connection_auth_failures_total":        radio.GetConnectionAuthFailureCount(),
			"ruckus_radio_connection_association_failures_total": radio.GetConnectionAssocFailureCount(),
			"ruckus_radio_connection_failures_total":             radio.GetConnectionTotalFailureCount(),
			"ruckus_radio_connections_total":                     radio.GetConnectionTotalCount(),
			"ruckus_radio_channel_changes_total":                 radio.GetNumOfChannelChange(),
			"ruckus_radio_latency_flagged":                       radio.GetIsLatencyFlagged(),
			"ruckus_radio_connection_failure_flagged":            radio.GetIsConnectionFailureFlagged(),
			"ruckus_radio_airtime_flagged":                       radio.GetIsAirtimeFlagged(),
			"ruckus_radio_enabled":                               radio.GetIsRadioEnabled(),
			"ruckus_radio_eirp_dbm":                              radio.GetEirp(),
			"ruckus_radio_tx_rts_total":                          radio.GetTxRtsCnt(),
			"ruckus_radio_failed_clients_total":                  radio.GetTotalFailureClientCount(),
			"ruckus_radio_clients_connected":                     radio.GetTotalClientCnts(),
			"ruckus_radio_rx_multicast_bytes_total":              radio.GetRxMcastBytes(),
			"ruckus_radio_tx_multicast_bytes_total":              radio.GetTxMcastBytes(),
			"ruckus_radio_rx_error_packets_total":                radio.GetRxErrorPkts(),
			"ruckus_radio_tx_error_packets_total":                radio.GetTxErrorPkts(),
			"ruckus_radio_tx_retry_bytes_total":                  radio.GetTxRetryBytes(),
			"ruckus_radio_tx_drop_bytes_total":                   radio.GetTxDropBytes(),
			"ruckus_radio_rx_drop_bytes_total":                   radio.GetRxDropBytes(),
			"ruckus_radio_authenticated_clients_total":           radio.GetNumAuthClients(),
			"ruckus_radio_antenna_gain_dbi":                      radio.GetAntennaGain(),
			"ruckus_radio_beacon_period":                         radio.GetBeaconPeriod(),
			"ruckus_radio_rx_decrypt_crc_errors_total":           radio.GetRxDecryptCrcError(),
			"ruckus_radio_snr_db":                                radio.GetRssi(),

			"ruckus_radio_info": 1,
		}

		labelMap := map[string]map[string]string{
			"ruckus_radio_info": radioInfoLabels,
			"default":           radioLabels,
		}

		if errs := appendMetrics(
			timestamp,
			radioMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err)
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"zone_name": apSystem.GetZoneName(),

		"ap_name": apSystem.GetDeviceName(),
		"ap_mac":  parseMAC(apSystem.GetAp()),
	}

	if err := prom.write(metricsFamily, apLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleApClient(systemID string, message *pb.APClientStats) error {
	timestamp := time.Unix(int64(message.GetSampleTime()), 0)

	metricsFamily := map[string]*dto.MetricFamily{}

	for _, client := range message.GetClients() {
		clientRadio := client.GetRadio()
		clientWlan := clientRadio.GetWlan()

		clientLabels := map[string]string{
			"radio_id":   strconv.Itoa(int(clientRadio.GetRadioId())),
			"ssid":       clientWlan.GetSsid(),
			"vlan":       parseVLAN(client.GetVlan()),
			"client_mac": parseMAC(client.GetClientMac()),
		}

		clientInfoLabels := map[string]string{
			"ipv4_address": client.GetIpAddress(),
			"ipv6_address": client.GetIpv6Address(),
			"os_type":      client.GetOsType(),
			"hostname":     client.GetHostname(),
			"username":     client.GetUsername(),
		}

		clientMetrics := map[string]interface{}{
			"ruckus_client_wireless_vlan":                         client.GetVlan(),
			"ruckus_client_wireless_snr_db":                       client.GetRssi(),
			"ruckus_client_wireless_rssi_dbm":                     client.GetReceiveSignalStrength(),
			"ruckus_client_wireless_noise_floor_dbm":              client.GetNoiseFloor(),
			"ruckus_client_wireless_rx_bytes_total":               client.GetRxBytes(),
			"ruckus_client_wireless_tx_bytes_total":               client.GetTxBytes(),
			"ruckus_client_wireless_rx_packets_total":             client.GetRxFrames(),
			"ruckus_client_wireless_tx_packets_total":             client.GetTxFrames(),
			"ruckus_client_wireless_throughput_estimate_kilobits": client.GetThroughputEst(),
			"ruckus_client_wireless_tx_drop_data_frames_total":    client.GetTxDropDataFrames(),
			"ruckus_client_wireless_tx_drop_mgmt_frames_total":    client.GetTxDropMgmtFrames(),
			"ruckus_client_wireless_rx_crc_error_frames_total":    client.GetRxCRCErrFrames(),
			"ruckus_client_wireless_median_tx_mcs_rate_kilobits":  client.GetMedianTxMCSRate(),
			"ruckus_client_wireless_median_rx_mcs_rate_kilobits":  client.GetMedianRxMCSRate(),
			"ruckus_client_wireless_rx_errors_total":              client.GetRxError(),
			"ruckus_client_wireless_tx_errors_total":              client.GetTxError(),
			"ruckus_client_wireless_reassociations_total":         client.GetReassocCount(),
			"ruckus_client_wireless_tx_retry_bytes_total":         client.GetTxRetryBytes(),
			"ruckus_client_wireless_rx_drop_packets_total":        client.GetRxDropPkts(),

			"ruckus_client_wireless_info": float64(1),
		}

		labelMap := map[string]map[string]string{
			"ruckus_client_wireless_info": clientInfoLabels,
			"default":                     clientLabels,
		}

		if errs := appendMetrics(
			timestamp,
			clientMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err)
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"zone_name": message.GetZoneName(),

		"ap_name": message.GetDeviceName(),
		"ap_mac":  parseMAC(message.GetAp()),
	}

	if err := prom.write(metricsFamily, apLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleApWiredClient(systemID string, message *pb.APWiredClientStats) error {
	timestamp := time.Unix(int64(message.GetSampleTime()), 0)

	metricsFamily := map[string]*dto.MetricFamily{}

	for _, client := range message.GetClients() {
		clientLabels := map[string]string{
			"port":       client.GetEthIF(),
			"vlan":       parseVLAN(client.GetVlan()),
			"client_mac": parseMAC(client.GetClientMac()),
		}

		clientInfoLabels := map[string]string{
			"ipv4_address": client.GetIpAddress(),
			"ipv6_address": client.GetIpv6Address(),
			"hostname":     client.GetHostname(),
		}

		clientMetrics := map[string]interface{}{
			"ruckus_client_wired_vlan":                     client.GetVlan(),
			"ruckus_client_wired_rx_bytes_total":           client.GetRxBytes(),
			"ruckus_client_wired_tx_bytes_total":           client.GetTxBytes(),
			"ruckus_client_wired_rx_packets_total":         client.GetRxFrames(),
			"ruckus_client_wired_tx_packets_total":         client.GetTxFrames(),
			"ruckus_client_wired_rx_drop_packets_total":    client.GetRxDrop(),
			"ruckus_client_wired_tx_drop_packets_total":    client.GetTxDrop(),
			"ruckus_client_wired_rx_multicast_bytes_total": client.GetRxMcast(),
			"ruckus_client_wired_tx_multicast_bytes_total": client.GetTxMcast(),
			"ruckus_client_wired_rx_broadcast_bytes_total": client.GetRxBcast(),
			"ruckus_client_wired_tx_broadcast_bytes_total": client.GetTxBcast(),
			"ruckus_client_wired_rx_unicast_bytes_total":   client.GetRxUcast(),
			"ruckus_client_wired_tx_unicast_bytes_total":   client.GetTxUcast(),

			"ruckus_client_wired_info": float64(1),
		}

		labelMap := map[string]map[string]string{
			"ruckus_client_wired_info": clientInfoLabels,
			"default":                  clientLabels,
		}

		if errs := appendMetrics(
			timestamp,
			clientMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err)
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"zone_id": message.GetZoneId(),

		"ap_mac": parseMAC(message.GetApmac()),
	}

	if err := prom.write(metricsFamily, apLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleApReport(systemID string, message *pb.APReportStats) error {
	metricsFamily := map[string]*dto.MetricFamily{}

	for _, client := range message.GetBinClient() {
		timestamp := time.Unix(int64(client.GetTime()), 0)

		clientLabels := map[string]string{
			"radio_id":   strconv.Itoa(int(client.GetRadioId())),
			"ssid":       client.GetSsid(),
			"vlan":       strconv.Itoa(int(client.GetClientVlan())),
			"client_mac": parseMAC(client.GetClientMac()),
		}

		clientMetrics := map[string]interface{}{
			"ruckus_client_wireless_ttc_authentication":    client.GetClientAuthTTC(),
			"ruckus_client_wireless_ttc_association":       client.GetClientAssocTTC(),
			"ruckus_client_wireless_ttc_eap":               client.GetClientEapTTC(),
			"ruckus_client_wireless_ttc_radius":            client.GetClientRadiusTTC(),
			"ruckus_client_wireless_ttc_dhcp":              client.GetClientDhcpTTC(),
			"ruckus_client_wireless_band_capability":       client.GetBandCap(),
			"ruckus_client_wireless_vht_capability":        client.GetVHTCap(),
			"ruckus_client_wireless_stream_capability":     client.GetStreamCap(),
			"ruckus_client_wireless_btm_capability":        client.GetStreamCap(),
			"ruckus_client_wireless_wifi6_capability":      client.GetWiFi6Cap(),
			"ruckus_client_wireless_roaming_failure_total": client.GetRoamingFailureCount(),
			"ruckus_client_wireless_roaming_success_total": client.GetRoamingSuccessCount(),
		}

		labelMap := map[string]map[string]string{
			"default": clientLabels,
		}

		if errs := appendMetrics(
			timestamp,
			clientMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err)
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"zone_name": message.GetZoneName(),

		"ap_name": message.GetDeviceName(),
		"ap_mac":  parseMAC(message.GetAp()),
	}

	if err := prom.write(metricsFamily, apLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleApConfigurationMessage(systemID string, message *pb.ConfigurationMessage) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	var apConfig []ApConfig
	err := json.Unmarshal([]byte(clusterInfo.GetAps()), &apConfig)
	if err != nil {
		slog.Error("Failed to ap configuration to JSON", "error", err)
	} else {
		for _, ap := range apConfig {
			apLabels := map[string]string{
				"zone_id": ap.ZoneID,

				"ap_name": ap.DeviceName,
				"ap_mac":  parseMAC(ap.ApID),
			}

			apInfoLabels := map[string]string{
				"blade_id":              ap.BladeID,
				"gps_position":          ap.GpsInfo,
				"ipv4_address":          ap.IP,
				"description":           ap.Description,
				"ipv4_address_external": ap.ExtIP,
				"serial_number":         ap.Serial,
				"model":                 ap.Model,
				"location":              ap.Location,
				"fw_version":            ap.FwVersion,
			}

			apMetrics := map[string]interface{}{
				"ruckus_cluster_ap_connected":           parseAPConnectionStatus(ap.ConnectionStatus),
				"ruckus_cluster_ap_last_seen_timestamp": ap.LastSeen,

				"ruckus_cluster_ap_info": 1,
			}

			regState, err := strconv.Atoi(ap.RegistrationState)
			if err == nil {
				apMetrics["ruckus_cluster_ap_registration_state"] = regState
			} else {
				slog.Error(
					"Failed to convert ap registration state from string",
					"value",
					ap.RegistrationState,
					"error",
					err,
				)
			}

			labelMap := map[string]map[string]string{
				"ruckus_cluster_ap_info": apInfoLabels,
				"default":                apLabels,
			}

			if errs := appendMetrics(
				timestamp,
				apMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err)
				}
			}
		}
	}

	clusterLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, clusterLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSystemConfigurationMessage(systemID string, message *pb.ConfigurationMessage) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	domainZones := map[string]map[string]string{}
	for _, tenant := range clusterInfo.GetTenantInfos() {
		adminDomain := tenant.GetAdminDomain()
		adminDomainName := adminDomain.GetDomainName()
		if domainZones[adminDomainName] == nil {
			domainZones[adminDomainName] = map[string]string{}
		}
		for _, zone := range adminDomain.GetZoneInfos() {
			domainZones[adminDomainName][zone.GetZoneId()] = zone.GetZoneName()
		}

		for _, domain := range adminDomain.GetSubDomainInfos() {
			domainName := domain.GetDomainName()
			if domainZones[domainName] == nil {
				domainZones[domainName] = map[string]string{}
			}
			for _, zone := range domain.GetZoneInfos() {
				domainZones[domainName][zone.GetZoneId()] = zone.GetZoneName()
			}
		}
	}

	for domainName, zones := range domainZones {
		for zoneID, zoneName := range zones {
			zoneLabels := map[string]string{
				"domain_name": domainName,

				"zone_id":   zoneID,
				"zone_name": zoneName,
			}

			zoneMetrics := map[string]interface{}{
				"ruckus_zone_info": 1,
			}

			labelMap := map[string]map[string]string{
				"ruckus_zone_info": zoneLabels,
			}

			if errs := appendMetrics(
				timestamp,
				zoneMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err)
				}
			}
		}
	}

	var controlBladesConfig []ControlBladeConfig
	err := json.Unmarshal([]byte(clusterInfo.GetControlBlades()), &controlBladesConfig)
	if err != nil {
		slog.Error("Failed to convert blade configuration to JSON", "error", err)
	} else {
		for _, c := range controlBladesConfig {
			bladeLabels := map[string]string{
				"blade_name": c.Name,
			}

			bladeIdInfoLabels := map[string]string{
				"blade_id": c.Key,
			}

			bladeInfoLabels := map[string]string{
				"hostname":      c.HostName,
				"fw_version":    c.Firmware,
				"cp_version":    c.CpVersion,
				"model":         c.Model,
				"serial_number": c.SerialNumber,
				"mac_address":   c.Mac,
				"ipv4_address":  c.IP,
				"state":         c.BladeStateEnum,
			}

			bladeMetrics := map[string]interface{}{
				"ruckus_blade_clients_connected":    c.ClientCount,
				"ruckus_blade_aps_associated":       c.AssociateAPs,
				"ruckus_blade_aps_connected":        c.ApCount,
				"ruckus_blade_uptime_seconds_total": c.UptimeSec,

				"ruckus_blade_id_info": 1,
				"ruckus_blade_info":    1,
			}

			labelMap := map[string]map[string]string{
				"ruckus_blade_id_info": bladeIdInfoLabels,
				"ruckus_blade_info":    bladeInfoLabels,
				"default":              bladeLabels,
			}

			if errs := appendMetrics(
				timestamp,
				bladeMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err)
				}
			}
		}
	}

	var bladePerformance ControllerUtilization
	err = json.Unmarshal([]byte(clusterInfo.GetControllerUtilizations()), &bladePerformance)
	if err != nil {
		slog.Error("Failed to convert controller utilization to JSON", "error", err)
	} else {
		for _, blade := range bladePerformance.Cbutils {
			for _, q := range blade.QuarterStats {
				t, err := strconv.Atoi(q.QuarterStartTime)
				if err != nil {
					slog.Error(
						"Failed to convert timestamp from string",
						"timestamp",
						q.QuarterStartTime,
						"error",
						err,
					)
					continue
				}

				quarterStartTime := time.UnixMilli(int64(t))

				memoryUsedPerc, err := strconv.ParseFloat(q.MemoryPerc, 64)
				if err != nil {
					slog.Error(
						"Failed to convert memory percentage from string",
						"value",
						q.MemoryPerc,
						"error",
						err,
					)
					continue
				}
				diskUsedPerc, err := strconv.ParseFloat(q.DiskPerc, 64)
				if err != nil {
					slog.Error(
						"Failed to convert disk percentage from string",
						"value",
						q.DiskPerc,
						"error",
						err,
					)
					continue
				}
				cpuUsedPerc, err := strconv.ParseFloat(q.CPUPerc, 64)
				if err != nil {
					slog.Error(
						"Failed to convert cpu percentage from string",
						"value",
						q.CPUPerc,
						"error",
						err,
					)
					continue
				}

				bladeLabels := map[string]string{
					"blade_id": blade.ControlID,
				}

				bladeMetrics := map[string]interface{}{
					"ruckus_blade_memory_used_percentage":  memoryUsedPerc,
					"ruckus_blade_storage_used_percentage": diskUsedPerc,
					"ruckus_blade_cpu_used_percentage":     cpuUsedPerc,
				}

				labelMap := map[string]map[string]string{
					"default": bladeLabels,
				}

				if errs := appendMetrics(
					quarterStartTime,
					bladeMetrics,
					labelMap,
					metricsFamily,
				); len(errs) >= 1 {
					for _, err := range errs {
						slog.Error(
							"Error while appending metrics",
							"error",
							err,
						)
					}
				}
			}
		}
	}

	var clusterSummary SystemSummary
	err = json.Unmarshal([]byte(clusterInfo.GetSystemSummary()), &clusterSummary)
	if err != nil {
		slog.Error("Failed to convert system summary to JSON", "error", err)
	} else {
		clusterInfoLabels := map[string]string{
			"fw_version":    clusterSummary.Version,
			"cp_version":    clusterSummary.CpVersion,
			"ap_version":    clusterSummary.ApVersion,
			"model":         clusterSummary.Model,
			"serial_number": clusterSummary.SerialNumber,
			"ipv4_address":  clusterSummary.IPAddress,
			"state":         clusterSummary.ClusterState,
			"blade_name":    clusterSummary.BladeName,
			"blade_state":   clusterSummary.BladeState,
		}

		clusterMetrics := map[string]interface{}{
			"ruckus_cluster_control_blades":                clusterSummary.NumOfControlBlades,
			"ruckus_cluster_control_blades_in_service":     clusterSummary.NumOfInServiceControlBlades,
			"ruckus_cluster_control_blades_out_of_service": clusterSummary.NumOfOutOfServiceControlBlades,
			"ruckus_cluster_data_blades":                   clusterSummary.NumOfDataBlades,
			"ruckus_cluster_data_blades_in_service":        clusterSummary.NumOfInServiceDataBlades,
			"ruckus_cluster_data_blades_out_of_service":    clusterSummary.NumOfOutOfServiceDataBlades,
			"ruckus_cluster_aps_connected":                 clusterSummary.NumOfConnectedAPs,
			"ruckus_cluster_aps_disconnected":              clusterSummary.NumOfDisconnectedAPs,
			"ruckus_cluster_uptime_seconds_total":          clusterSummary.Uptime,

			"ruckus_cluster_info": 1,
		}

		maxAP, err := strconv.Atoi(clusterSummary.MaxSupportedAps)
		if err == nil {
			clusterMetrics["ruckus_cluster_aps_max"] = maxAP
		} else {
			slog.Error(
				"Failed to convert max supported APs from string",
				"value",
				clusterSummary.MaxSupportedAps,
				"error",
				err,
			)
		}

		l := strings.Split(clusterSummary.ControllerLicenseSummary, "/")
		if len(l) == 2 {
			used, err1 := strconv.Atoi(l[0])
			avail, err2 := strconv.Atoi(l[1])
			if err1 == nil && err2 == nil && avail != 0 {
				v := float64(used) / float64(avail)
				clusterMetrics["ruckus_cluster_licence_used_percentage"] = v * 100
			}
			if err1 != nil {
				slog.Error(
					"Failed to convert used controller licenses from string",
					"value",
					l[0],
					"error",
					err1,
				)
			}
			if err2 != nil {
				slog.Error(
					"Failed to convert available controller licenses from string",
					"value",
					l[1],
					"error",
					err2,
				)
			}
		} else {
			slog.Error(
				"Failed to parse controller license summary",
				"value",
				clusterSummary.MaxSupportedAps,
				"error",
				err,
			)
		}

		labelMap := map[string]map[string]string{
			"ruckus_cluster_info": clusterInfoLabels,
			"default":             {},
		}

		if errs := appendMetrics(
			timestamp,
			clusterMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err)
			}
		}
	}

	clusterLabels := map[string]string{
		"system_id": systemID,

		"cluster_name": clusterSummary.ClusterName,
	}

	if err := prom.write(metricsFamily, clusterLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}
