package main

import (
	"encoding/json"
	"log/slog"
	"net"
	"strconv"
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"
	"github.com/pkg/errors"

	dto "github.com/prometheus/client_model/go"
)

func handleSwitchStatus(systemID string, ts int64, message *pb.SwitchStatus) error {
	timestamp := time.UnixMilli(ts)

	switchInfoLabels := map[string]string{
		"ipv4_address":     message.GetIpAddress(),
		"fw_version":       message.GetFirmware(),
		"sw_version":       message.GetSwitchSWVersion(),
		"model":            message.GetModel(),
		"serial_number":    message.GetSerialNumber(),
		"mode":             message.GetSwitchMode(),
		"switch_type":      message.GetModules(),
		"active_partition": message.GetPartitionInUse(),
	}

	switchMetrics := map[string]interface{}{
		"ruckus_switch_status":                 parseSwitchStatus(message.GetStatus()),
		"ruckus_switch_sync_status":            int(message.GetLocalsyncStatus().Number()),
		"ruckus_switch_uptime_seconds_total":   parseUptime(message.GetUptime()),
		"ruckus_switch_port_count":             message.GetNumOfPorts(),
		"ruckus_switch_last_backup_timestamp":  message.GetLastBackup(),
		"ruckus_switch_cpu_used_percentage":    message.GetCpu(),
		"ruckus_switch_memory_used_percentage": message.GetMemory(),
		"ruckus_switch_alert_count":            message.GetAlerts(),
		"ruckus_switch_unit_count":             message.GetNumOfUnits(),
		"ruckus_switch_operational_status":     message.GetOperational(),
		"ruckus_switch_poe_total_milliwatts":   message.GetPoeTotal(),
		"ruckus_switch_poe_used_milliwatts":    message.GetPoeUtilization(),

		"ruckus_switch_info": 1,
	}

	labelMap := map[string]map[string]string{
		"ruckus_switch_info": switchInfoLabels,
		"default":            {},
	}

	metricsFamily := map[string]*dto.MetricFamily{}

	if errs := appendMetrics(timestamp, switchMetrics, labelMap, metricsFamily); len(errs) >= 1 {
		for _, err := range errs {
			slog.Error("Error while appending metrics", "error", err.Error())
		}
	}

	if message.GetPowerSupplyGroups() != "" {
		var psuStatus []SwitchPowerSupplyStatus
		err := json.Unmarshal([]byte(message.GetPowerSupplyGroups()), &psuStatus)
		if err != nil {
			instJSONUnparseableCounter.WithLabelValues(systemID, "switch_status_psu").Inc()
			slog.Error("Failed to convert switch psu status to JSON", "error", err)
		} else {
			for i, unit := range psuStatus {
				if i == 0 {
					// Skip first entry which is a PSU summary for the
					// entire stack, we are only interested in the PSU
					// status for individual units in the switch stack
					continue
				}

				for _, slot := range unit.PowerSupplySlotList {
					psuLabels := map[string]string{
						"unit_serial": unit.SerialNumber,
						"psu_id":      strconv.Itoa(slot.SlotNumber),
					}

					psuInfoLabels := map[string]string{
						"type": slot.Type,
					}

					psuMetrics := map[string]interface{}{
						"ruckus_switch_unit_psu_status": parseOk(slot.Status),

						"ruckus_switch_unit_psu_info": 1,
					}

					psuLabelMap := map[string]map[string]string{
						"ruckus_switch_unit_psu_info": psuInfoLabels,
						"default":                     psuLabels,
					}

					if errs := appendMetrics(
						timestamp,
						psuMetrics,
						psuLabelMap,
						metricsFamily,
					); len(errs) >= 1 {
						for _, err := range errs {
							slog.Error(
								"Error while appending metrics",
								"error",
								err.Error(),
							)
						}
					}
				}
			}
		}
	}

	if message.GetFanGroups() != "" {
		var fanStatus []SwitchFanStatus
		err := json.Unmarshal([]byte(message.GetFanGroups()), &fanStatus)
		if err != nil {
			instJSONUnparseableCounter.WithLabelValues(systemID, "switch_status_fan").Inc()
			slog.Error("Failed to convert switch fan status to JSON", "error", err)
		} else {
			for _, unit := range fanStatus {
				for _, slot := range unit.FanSlotList {
					fanLabels := map[string]string{
						"unit_serial": unit.SerialNumber,
						"fan_id":      strconv.Itoa(slot.SlotNumber),
					}

					fanMetrics := map[string]interface{}{
						"ruckus_switch_unit_fan_status": parseOk(slot.Status),
					}

					fanLabelMap := map[string]map[string]string{
						"default": fanLabels,
					}

					if errs := appendMetrics(
						timestamp,
						fanMetrics,
						fanLabelMap,
						metricsFamily,
					); len(errs) >= 1 {
						for _, err := range errs {
							slog.Error(
								"Error while appending metrics",
								"error",
								err.Error(),
							)
						}
					}
				}
			}
		}
	}

	if message.GetTemperatureGroups() != "" {
		var temperatureStatus []SwitchTemperatureStatus
		err := json.Unmarshal([]byte(message.GetTemperatureGroups()), &temperatureStatus)
		if err != nil {
			instJSONUnparseableCounter.WithLabelValues(systemID, "switch_status_temperature").Inc()
			slog.Error("Failed to convert switch temperature status to JSON", "error", err)
		} else {
			for _, unit := range temperatureStatus {
				for _, slot := range unit.TemperatureSlotList {
					temperatureLabels := map[string]string{
						"unit_serial": unit.SerialNumber,
						"sensor_id":   strconv.Itoa(slot.SlotNumber),
					}

					temperatureMetrics := map[string]interface{}{
						"ruckus_switch_unit_temperature_celsius": slot.TemperatureValue,
					}

					temperatureLabelMap := map[string]map[string]string{
						"default": temperatureLabels,
					}

					if errs := appendMetrics(
						timestamp,
						temperatureMetrics,
						temperatureLabelMap,
						metricsFamily,
					); len(errs) >= 1 {
						for _, err := range errs {
							slog.Error(
								"Error while appending metrics",
								"error",
								err.Error(),
							)
						}
					}
				}
			}
		}
	}

	labels := map[string]string{
		"system_id": systemID,

		"domain_name":       message.GetDomainName(),
		"switch_group_name": message.GetSwitchGroupLevelOneName(),

		"switch_name": message.GetSwitchName(),
		"switch_mac":  parseMAC(message.GetId()),
	}

	if err := prom.write(metricsFamily, labels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchUnitStatus(systemID string, ts int64, switchUnits []*pb.SwitchUnitStatus) error {
	timestamp := time.UnixMilli(ts)

	metricsFamily := map[string]*dto.MetricFamily{}

	for _, switchUnit := range switchUnits {
		switchUnitLabels := map[string]string{
			"domain_name":       switchUnit.GetDomainName(),
			"switch_group_name": switchUnit.GetSwitchGroupLevelOneName(),

			"switch_name": switchUnit.GetUnitName(),
			"switch_mac":  parseMAC(switchUnit.GetSwitchId()),

			"unit_id":     strconv.Itoa(int(switchUnit.GetUnitId())),
			"unit_serial": switchUnit.GetId(),
		}

		switchUnitInfoLabels := map[string]string{
			"status": switchUnit.GetUnitStatus(),
			"state":  switchUnit.GetUnitState(),
		}

		switchUnitMetrics := map[string]interface{}{
			"ruckus_switch_unit_poe_total_milliwatts": switchUnit.GetPoeTotal(),
			"ruckus_switch_unit_poe_used_milliwatts":  switchUnit.GetPoeUtilization(),

			"ruckus_switch_unit_info": 1,
		}

		labelMap := map[string]map[string]string{
			"ruckus_switch_unit_info": switchUnitInfoLabels,
			"default":                 switchUnitLabels,
		}

		if errs := appendMetrics(timestamp, switchUnitMetrics, labelMap, metricsFamily); len(
			errs,
		) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	systemLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, systemLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchPortStatus(systemID string, ts int64, ports []*pb.PortStatus) error {
	timestamp := time.UnixMilli(ts)

	metricsFamily := map[string]*dto.MetricFamily{}

	for _, port := range ports {
		portLabels := map[string]string{
			"domain_name":       port.GetDomainName(),
			"switch_group_name": port.GetSwitchGroupLevelOneName(),

			"switch_name": port.GetSwitchName(),
			"switch_mac":  parseMAC(port.GetSwitchId()),

			"unit_serial": port.GetSwitchUnitId(),

			"port": port.GetPortIdentifier(),
		}

		portInfoLabels := map[string]string{
			"type":        port.GetType(),
			"description": port.GetName(),
			"port_mac":    parseMAC(port.GetPortMac()),

			"acl_in":  port.GetInAclConfigName(),
			"acl_out": port.GetOutAclConfigName(),

			"stp_status": port.GetSpanningTreeStatus(),

			"stack_enabled": strconv.FormatBool(port.GetUsedInFormingStack()),
			"lag_enabled":   strconv.FormatBool(port.GetIsLagMember()),
			"lldp_enabled":  strconv.FormatBool(port.GetLldpEnabled()),
			"poe_enabled":   strconv.FormatBool(port.GetPoeEnabled()),

			"speed":            port.GetPortSpeed(),
			"transceiver_type": port.GetOpticsType(),
			"phy_capability":   port.GetPortSpeedCapacity(),
		}

		lldpInfoLabels := map[string]string{
			"lldp_neighbor_name": port.GetNeighborName(),
			"lldp_neighbor_mac":  parseMAC(port.GetNeighborMacAddress()),
		}

		lagInfoLabels := map[string]string{
			"lag_id":   strconv.Itoa(int(port.GetLagId())),
			"lag_name": port.GetLagName(),

			"lag_status":       port.GetLagStatus(),
			"lag_admin_status": port.GetLagAdminStatus(),
		}

		poeInfoLabels := map[string]string{
			"poe_type":                                port.GetPoeType(),
			"poe_pd_class":                            port.GetPoePdClass(),
			"poe_pd_class_b":                          port.GetPoePdClassB(),
			"poe_lldp_max_power_request_milliwatts":   port.GetPoeLldpMaxPowerRequest(),
			"poe_lldp_max_power_request_a_milliwatts": port.GetPoeLldpMaxPowerRequestA(),
			"poe_lldp_max_power_request_b_milliwatts": port.GetPoeLldpMaxPowerRequestB(),
			"poe_2_pair_max_power_milliwatts":         port.GetPoe2PairMaxPower(),
			"poe_4_pair_max_power_milliwatts":         port.GetPoe4PairMaxPower(),
			"poe_overdrive_mode":                      port.GetPoeOverdriveMode().String(),
			"poe_port_cabability":                     port.GetPoePortCapability().String(),
		}

		portMetrics := map[string]interface{}{
			"ruckus_switch_port_phy_status":    parseUpDown(port.GetStatus()),
			"ruckus_switch_port_admin_status":  parseUpDown(port.GetAdminStatus()),
			"ruckus_switch_port_warning_state": port.GetIsInWarningState(),

			"ruckus_switch_port_stp_status": parseSTPStatus(port.GetSpanningTreeStatus()),

			"ruckus_switch_port_rx_bytes_total":             port.GetRx(),
			"ruckus_switch_port_rx_broadcast_packets_total": port.GetBroadcastIn(),
			"ruckus_switch_port_rx_multicast_packets_total": port.GetMulticastIn(),
			"ruckus_switch_port_tx_bytes_total":             port.GetTx(),
			"ruckus_switch_port_tx_broadcast_packets_total": port.GetBroadcastOut(),
			"ruckus_switch_port_tx_multicast_packets_total": port.GetMulticastOut(),

			"ruckus_switch_port_rx_errors_total":     port.GetInErr(),
			"ruckus_switch_port_rx_crc_errors_total": port.GetCrcErr(),
			"ruckus_switch_port_rx_discards_total":   port.GetInDiscard(),
			"ruckus_switch_port_tx_errors_total":     port.GetOutErr(),

			"ruckus_switch_port_info": 1,
		}

		if port.GetIsLagMember() {
			portMetrics["ruckus_switch_port_lag_status"] = parseUpDown(
				port.GetLagStatus(),
			)
			portMetrics["ruckus_switch_port_lag_admin_status"] = parseUpDown(
				port.GetLagAdminStatus(),
			)
			portMetrics["ruckus_switch_port_lag_info"] = 1
		}

		if port.GetLldpEnabled() {
			portMetrics["ruckus_switch_port_lldp_info"] = 1
		}

		if port.GetPoeEnabled() {
			portMetrics["ruckus_switch_port_poe_class"] = port.GetPoeClass()
			portMetrics["ruckus_switch_port_poe_priority"] = port.GetPoePriority()
			portMetrics["ruckus_switch_port_poe_total_milliwatts"] = port.GetPoeTotal()
			portMetrics["ruckus_switch_port_poe_used_milliwatts"] = port.GetPoeUsed()
			portMetrics["ruckus_switch_port_poe_budget_milliwatts"] = port.GetPoeBudget()
			portMetrics["ruckus_switch_port_poe_info"] = 1
		}

		labelMap := map[string]map[string]string{
			"ruckus_switch_port_info":      portInfoLabels,
			"ruckus_switch_port_lag_info":  lagInfoLabels,
			"ruckus_switch_port_lldp_info": lldpInfoLabels,
			"ruckus_switch_port_poe_info":  poeInfoLabels,
			"default":                      portLabels,
		}

		if port.GetVlanDetailInformation() != "" {
			var vlanDetail []VLANDetail
			err := json.Unmarshal([]byte(port.GetVlanDetailInformation()), &vlanDetail)
			if err != nil {
				instJSONUnparseableCounter.WithLabelValues(systemID, "switch_vlan_detail").Inc()
				slog.Error("Failed to convert switch vlan detail to JSON", "error", err)
			} else {
				for _, vlan := range vlanDetail {
					var vlanType string
					if vlan.VlanID == port.GetUnTaggedVlan() {
						vlanType = "untagged"
					} else {
						vlanType = "tagged"
					}

					vlanInfoLabels := map[string]string{
						"vlan":      vlan.VlanID,
						"vlan_name": vlan.VlanName,
						"vlan_type": vlanType,
					}

					vlanMetrics := map[string]interface{}{
						"ruckus_switch_port_vlan_info": 1,
					}

					vlanLabelMap := map[string]map[string]string{
						"ruckus_switch_port_vlan_info": vlanInfoLabels,
						"default":                      portLabels,
					}

					if errs := appendMetrics(timestamp, vlanMetrics, vlanLabelMap, metricsFamily); len(errs) >= 1 {
						for _, err := range errs {
							slog.Error("Error while appending metrics", "error", err.Error())
						}
					}
				}
			}
		}

		if errs := appendMetrics(timestamp, portMetrics, labelMap, metricsFamily); len(errs) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	systemLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, systemLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchConnectedDeviceStatus(
	systemID string,
	devices []*pb.ConnectedDeviceStatus,
) error {
	metricsFamily := map[string]*dto.MetricFamily{}

	for _, device := range devices {
		timestamp := time.UnixMilli(device.GetUpdatedTime())

		deviceLabels := map[string]string{
			"domain_name":       device.GetDomainName(),
			"switch_group_name": device.GetSwitchGroupLevelOneName(),

			"switch_name": device.GetSwitchName(),
			"switch_mac":  parseMAC(device.GetSwitchId()),

			"unit_serial": device.GetUnitId(),

			"port": parseLongPortName(device.GetLocalPort()),

			"client_mac": parseMAC(device.GetRemotePortMac()),
		}

		deviceInfoLabels := map[string]string{
			"client_name":        device.GetRemoteDeviceName(),
			"client_port":        device.GetRemotePort(),
			"client_description": device.GetRemotePortDesc(),
		}

		if ip := net.ParseIP(device.GetRemoteDeviceMac()); ip != nil {
			// RemoteDeviceMac sometimes contains the device IP
			deviceInfoLabels["client_ip"] = device.GetRemoteDeviceMac()
		}

		deviceMetrics := map[string]interface{}{
			"ruckus_switch_device_info": 1,
		}

		labelMap := map[string]map[string]string{
			"ruckus_switch_device_info": deviceInfoLabels,
			"default":                   deviceLabels,
		}

		if errs := appendMetrics(timestamp, deviceMetrics, labelMap, metricsFamily); len(
			errs,
		) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	systemLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, systemLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchClientVisibility(
	systemID string,
	clients []*pb.SwitchClientVisibility,
) error {
	metricsFamily := map[string]*dto.MetricFamily{}

	for _, client := range clients {
		if client.GetClientType() == pb.ClientType_ROUTER ||
			client.GetClientType() == pb.ClientType_BRIDGE {
			// Only want clients connected to edge ports, not other L2 or L3 devices
			continue
		}

		timestamp := time.UnixMilli(client.GetUpdatedTime())

		clientLabels := map[string]string{
			"domain_name":       client.GetDomainName(),
			"switch_group_name": client.GetSwitchGroupLevelOneName(),

			"switch_name": client.GetSwitchName(),
			"switch_mac":  parseMAC(client.GetSwitchId()),

			"unit_serial": client.GetUnitId(),

			"port": client.GetSwitchPort(),

			"client_mac": parseMAC(client.GetClientMac()),
		}

		clientInfoLabels := map[string]string{
			"client_name":        client.GetClientName(),
			"client_type":        client.GetClientType().String(),
			"client_description": client.GetClientDesc(),

			"vlan":      client.GetClientVlan(),
			"vlan_name": client.GetVlanName(),

			"username":    client.GetClientUserName(),
			"auth_type":   client.GetClientAuthType().String(),
			"auth_status": client.GetClientAuthStatus().String(),

			"ipv4_address": client.GetClientIpv4Addr(),
			"ipv6_address": client.GetClientIpv6Addr(),

			"ruckus_ap": strconv.FormatBool(client.GetIsRuckusAP()),
		}

		client8021xInfoLabels := map[string]string{
			"ipv4_address": client.GetDot1XIpv4Addr(),
			"ipv6_address": client.GetDot1XIpv6Addr(),
		}

		clientMetrics := map[string]interface{}{
			"ruckus_switch_client_vlan":                 parseIntegerString(client.GetClientVlan()),
			"ruckus_switch_client_auth_status":          int(client.GetClientAuthStatus().Number()),
			"ruckus_switch_client_uptime_seconds_total": parseUptime(client.GetClientUpTime()),

			"ruckus_switch_client_info": 1,
		}

		if client.GetClientAuthType() == pb.ClientAuthType_DOT1X {
			clientMetrics["ruckus_switch_client_8021x_info"] = 1
		}

		labelMap := map[string]map[string]string{
			"ruckus_switch_client_info":       clientInfoLabels,
			"ruckus_switch_client_8021x_info": client8021xInfoLabels,
			"default":                         clientLabels,
		}

		if errs := appendMetrics(timestamp, clientMetrics, labelMap, metricsFamily); len(
			errs,
		) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	systemLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, systemLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchConfigurationMessage(
	systemID string,
	message *pb.SwitchConfigurationMessage,
) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	tenantDomains := map[string][]*pb.IcxDomainMessage{}
	for _, tenant := range clusterInfo.GetTenantInfos() {
		tenantName := tenant.GetTenantName()
		tenantDomains[tenantName] = []*pb.IcxDomainMessage{}

		root := tenant.GetAdminDomain()
		stack := []*pb.IcxDomainMessage{root}

		for len(stack) > 0 {
			for range stack {
				node := stack[0]
				tenantDomains[tenantName] = append(
					tenantDomains[tenantName],
					node,
				)

				stack = stack[1:]
				stack = append(stack, node.GetSubDomainInfos()...)
			}
		}
	}

	for tenantName, domains := range tenantDomains {
		for _, domain := range domains {
			for _, switchGroup := range domain.GetSwitchGroupInfos() {
				switchGroupLabels := map[string]string{
					"tenant_name": tenantName,
					"domain_name": domain.GetDomainName(),

					"switch_group_id":   switchGroup.GetSwitchGroupId(),
					"switch_group_name": switchGroup.GetSwitchGroupName(),
				}

				switchGroupMetrics := map[string]interface{}{
					"ruckus_switch_group_info": 1,
				}

				labelMap := map[string]map[string]string{
					"ruckus_switch_group_info": switchGroupLabels,
				}

				if errs := appendMetrics(
					timestamp,
					switchGroupMetrics,
					labelMap,
					metricsFamily,
				); len(errs) >= 1 {
					for _, err := range errs {
						slog.Error("Error while appending metrics", "error", err.Error())
					}
				}
			}
		}
	}

	systemLabels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, systemLabels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}

func handleSwitchDetailMessage(systemID string, ts int64, message *pb.SwitchDetailMessage) error {
	timestamp := time.UnixMilli(ts)

	switches := message.GetSwitchDetail()

	metricsFamily := map[string]*dto.MetricFamily{}

	for _, sw := range switches {
		switchLabels := map[string]string{
			"switch_group_name": sw.GetGroupName(),
			"switch_group_id":   sw.GetGroupId(),

			"switch_name": sw.GetSwitchName(),
			"switch_mac":  parseMAC(sw.GetId()),
		}

		switchInfoLabels := map[string]string{
			"ipv4_address":  sw.GetIpAddress(),
			"fw_version":    sw.GetFirmwareVersion(),
			"model":         sw.GetModel(),
			"serial_number": sw.GetSerialNumber(),

			"status":              sw.GetStatus(),
			"operational_status":  strconv.FormatBool(sw.GetOperational()),
			"registration_status": sw.GetRegistrationStatus(),
		}

		if sw.GetIsStack() {
			switchInfoLabels["switch_type"] = "stack"
		} else {
			switchInfoLabels["switch_type"] = "switch"
		}

		switchMetrics := map[string]interface{}{
			"ruckus_cluster_switch_status":             parseSwitchStatus(sw.GetStatus()),
			"ruckus_cluster_switch_operational_status": sw.GetOperational(),
			"ruckus_cluster_switch_registration_status": parseRegistrationStatus(
				sw.GetRegistrationStatus(),
			),
			"ruckus_cluster_switch_unit_count": sw.GetNumOfUnits(),
			"ruckus_cluster_switch_port_count": sw.GetNumOfPorts(),

			"ruckus_cluster_switch_info": 1,
		}

		labelMap := map[string]map[string]string{
			"ruckus_cluster_switch_info": switchInfoLabels,
			"default":                    switchLabels,
		}

		if errs := appendMetrics(timestamp, switchMetrics, labelMap, metricsFamily); len(
			errs,
		) >= 1 {
			for _, err := range errs {
				slog.Error("Error while appending metrics", "error", err.Error())
			}
		}
	}

	labels := map[string]string{
		"system_id": systemID,
	}

	if err := prom.write(metricsFamily, labels); err != nil {
		return errors.Wrapf(err, "Error writing metrics to prometheus")
	}

	return nil
}
