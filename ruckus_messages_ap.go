package main

import (
	"log/slog"
	"strconv"
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
)

func handleApStatus(systemID string, message *pb.APStatus) error {
	apStatusData := message.GetApStatusData()
	apSystem := apStatusData.GetAPSystem()

	timestamp := time.Unix(int64(message.GetSampleTime()), 0)

	apInfoLabels := map[string]string{
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
			slog.Error("Error while appending metrics", "error", err.Error())
		}
	}

	for _, port := range apStatusData.GetLanPortStatus() {
		portLabels := map[string]string{
			"port_id": strconv.Itoa(int(port.GetPort())),
			"port":    port.GetInterface(),
		}

		portInfoLabels := map[string]string{
			"speed":          parsePhyLinkSpeed(port.GetPhyLink()),
			"duplex":         parsePhyLinkDuplex(port.GetPhyLink()),
			"phy_capability": port.GetPhyCapability(),
		}

		portMetrics := map[string]interface{}{
			"ruckus_ap_lan_port_logical_state": parseUpDown(port.GetLogicLink()),
			"ruckus_ap_lan_port_phy_state":     parseUpDown(parsePhyLinkStatus(port.GetPhyLink())),

			"ruckus_ap_lan_port_info": 1,
		}

		labelMap := map[string]map[string]string{
			"ruckus_ap_lan_port_info": portInfoLabels,
			"default":                 portLabels,
		}

		if errs := appendMetrics(
			timestamp,
			portMetrics,
			labelMap,
			metricsFamily,
		); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	for _, radio := range apStatusData.GetAPRadio() {
		radioLabels := map[string]string{
			"radio_id": strconv.Itoa(int(radio.GetRadioId())),
			"band":     radio.GetBand(),
		}

		radioInfoLabels := map[string]string{
			"radio_mode": radio.GetRadioMode(),
			"tx_power":   radio.GetTxPower(),
		}

		radioMetrics := map[string]interface{}{
			"ruckus_radio_channel":                            radio.GetChannel(),
			"ruckus_radio_rx_phy_errors":                      radio.GetPhyError(),
			"ruckus_radio_noise_floor_dbm":                    radio.GetNoiseFloor(),
			"ruckus_radio_rx_bytes_total":                     radio.GetRxBytes(),
			"ruckus_radio_tx_bytes_total":                     radio.GetTxBytes(),
			"ruckus_radio_rx_packets_total":                   radio.GetRxFrames(),
			"ruckus_radio_tx_packets_total":                   radio.GetTxFrames(),
			"ruckus_radio_retries_total":                      radio.GetRetry(),
			"ruckus_radio_drops_total":                        radio.GetDrop(),
			"ruckus_radio_channel_utilization_percentage":     radio.GetTotal(),
			"ruckus_radio_channel_busy_time_percentage":       radio.GetBusy(),
			"ruckus_radio_rx_channel_availability_percentage": radio.GetRx(),
			"ruckus_radio_tx_channel_availability_percentage": radio.GetTx(),
			"ruckus_radio_channel_width_mhz": parseChannelWidth(
				radio.GetChannelWidth(),
			),
			"ruckus_radio_latency_microseconds":                     radio.GetLatency(),
			"ruckus_radio_capacity_megabits":                        radio.GetCapacity(),
			"ruckus_radio_connection_authentication_failures_total": radio.GetConnectionAuthFailureCount(),
			"ruckus_radio_connection_association_failures_total":    radio.GetConnectionAssocFailureCount(),
			"ruckus_radio_connection_failures_total":                radio.GetConnectionTotalFailureCount(),
			"ruckus_radio_connections_total":                        radio.GetConnectionTotalCount(),
			"ruckus_radio_channel_changes_total":                    radio.GetNumOfChannelChange(),
			"ruckus_radio_latency_flagged":                          radio.GetIsLatencyFlagged(),
			"ruckus_radio_connection_failures_flagged":              radio.GetIsConnectionFailureFlagged(),
			"ruckus_radio_airtime_flagged":                          radio.GetIsAirtimeFlagged(),
			"ruckus_radio_enabled":                                  radio.GetIsRadioEnabled(),
			"ruckus_radio_eirp_dbm":                                 radio.GetEirp(),
			"ruckus_radio_tx_rts_total":                             radio.GetTxRtsCnt(),
			"ruckus_radio_failed_clients_total":                     radio.GetTotalFailureClientCount(),
			"ruckus_radio_clients_connected":                        radio.GetTotalClientCnts(),
			"ruckus_radio_rx_multicast_bytes_total":                 radio.GetRxMcastBytes(),
			"ruckus_radio_tx_multicast_bytes_total":                 radio.GetTxMcastBytes(),
			"ruckus_radio_rx_error_packets_total":                   radio.GetRxErrorPkts(),
			"ruckus_radio_tx_error_packets_total":                   radio.GetTxErrorPkts(),
			"ruckus_radio_authenticated_clients_total":              radio.GetNumAuthClients(),
			"ruckus_radio_antenna_gain_dbi":                         radio.GetAntennaGain(),
			"ruckus_radio_beacon_period":                            radio.GetBeaconPeriod(),
			"ruckus_radio_rx_decrypt_crc_errors_total":              radio.GetRxDecryptCrcError(),
			"ruckus_radio_snr_db":                                   radio.GetRssi(),

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
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"domain_name":   message.GetDomainName(),
		"zone_name":     message.GetZoneName(),
		"ap_group_name": message.GetApgroupName(),

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
			"ruckus_client_wireless_tx_retry_packets_total":       client.GetTxRetry(),
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
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"domain_name":   message.GetDomainName(),
		"zone_name":     message.GetZoneName(),
		"ap_group_name": message.GetApgroupName(),

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
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"domain_id": message.GetDomainId(),
		"zone_id":   message.GetZoneId(),

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
			"ruckus_client_wireless_ttc_authentication":      client.GetClientAuthTTC(),
			"ruckus_client_wireless_ttc_association":         client.GetClientAssocTTC(),
			"ruckus_client_wireless_ttc_eap":                 client.GetClientEapTTC(),
			"ruckus_client_wireless_ttc_radius":              client.GetClientRadiusTTC(),
			"ruckus_client_wireless_ttc_dhcp":                client.GetClientDhcpTTC(),
			"ruckus_client_wireless_band_capability":         client.GetBandCap(),
			"ruckus_client_wireless_vht_capability":          client.GetVHTCap(),
			"ruckus_client_wireless_stream_capability":       client.GetStreamCap(),
			"ruckus_client_wireless_btm_capability":          client.GetStreamCap(),
			"ruckus_client_wireless_wifi6_capability":        client.GetWiFi6Cap(),
			"ruckus_client_wireless_roaming_failures_total":  client.GetRoamingFailureCount(),
			"ruckus_client_wireless_roaming_successes_total": client.GetRoamingSuccessCount(),
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
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	apLabels := map[string]string{
		"system_id": systemID,

		"domain_name":   message.GetDomainName(),
		"zone_name":     message.GetZoneName(),
		"ap_group_name": message.GetApgroupName(),

		"ap_name": message.GetDeviceName(),
		"ap_mac":  parseMAC(message.GetAp()),
	}

	if err := prom.write(metricsFamily, apLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}
