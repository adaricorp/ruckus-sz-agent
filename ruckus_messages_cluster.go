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

func handleApConfigurationMessage(systemID string, message *pb.ConfigurationMessage) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	var apConfig []ApConfig
	err := json.Unmarshal([]byte(clusterInfo.GetAps()), &apConfig)
	if err != nil {
		instJSONUnparseableCounter.WithLabelValues(systemID, "ap_configuration").Inc()
		slog.Error("Failed to convert ap configuration to JSON", "error", err)
	} else {
		for _, ap := range apConfig {
			apLabels := map[string]string{
				"zone_id":     ap.ZoneID,
				"ap_group_id": ap.ApGroupID,

				"ap_name": ap.DeviceName,
				"ap_mac":  parseMAC(ap.ApID),
			}

			apInfoLabels := map[string]string{
				"blade_id":              ap.BladeID,
				"latitude":              parseLatitude(ap.GpsInfo),
				"longitude":             parseLongitude(ap.GpsInfo),
				"ipv4_address":          ap.IP,
				"description":           ap.Description,
				"ipv4_address_external": ap.ExtIP,
				"serial_number":         ap.Serial,
				"model":                 ap.Model,
				"location":              ap.Location,
				"fw_version":            ap.FwVersion,

				"status":              strconv.FormatBool(parseAPConnectionStatus(ap.ConnectionStatus)),
				"registration_status": ap.RegistrationState,
			}

			apMetrics := map[string]interface{}{
				"ruckus_cluster_ap_status":              parseAPConnectionStatus(ap.ConnectionStatus),
				"ruckus_cluster_ap_registration_status": parseIntegerString(ap.RegistrationState),
				"ruckus_cluster_ap_last_seen_timestamp": ap.LastSeen,

				"ruckus_cluster_ap_info": 1,
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
					slog.Error("Error while appending metrics", "error", err.Error())
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

	tenantDomains := map[string][]*pb.DomainMessage{}
	for _, tenant := range clusterInfo.GetTenantInfos() {
		tenantName := tenant.GetTenantName()
		tenantDomains[tenantName] = []*pb.DomainMessage{}

		root := tenant.GetAdminDomain()
		stack := []*pb.DomainMessage{root}

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
			domainLabels := map[string]string{
				"tenant_name": tenantName,
				"domain_id":   domain.GetDomainId(),
				"domain_name": domain.GetDomainName(),
			}

			domainMetrics := map[string]interface{}{
				"ruckus_domain_info": 1,
			}

			labelMap := map[string]map[string]string{
				"ruckus_domain_info": domainLabels,
			}

			if errs := appendMetrics(
				timestamp,
				domainMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err.Error())
				}
			}

			for _, zone := range domain.GetZoneInfos() {
				zoneLabels := map[string]string{
					"tenant_name": tenantName,
					"domain_name": domain.GetDomainName(),

					"zone_id":   zone.GetZoneId(),
					"zone_name": zone.GetZoneName(),
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
						slog.Error("Error while appending metrics", "error", err.Error())
					}
				}
			}
		}
	}

	var controlBladesConfig []ControlBladeConfig
	err := json.Unmarshal([]byte(clusterInfo.GetControlBlades()), &controlBladesConfig)
	if err != nil {
		instJSONUnparseableCounter.WithLabelValues(
			systemID,
			"system_configuration_control_blades",
		).Inc()
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
				"ruckus_blade_state":                parseClusterState(c.BladeStateEnum),
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
					slog.Error("Error while appending metrics", "error", err.Error())
				}
			}
		}
	}

	var bladePerformance ControllerUtilization
	err = json.Unmarshal([]byte(clusterInfo.GetControllerUtilizations()), &bladePerformance)
	if err != nil {
		instJSONUnparseableCounter.WithLabelValues(
			systemID,
			"system_configuration_controller_utilization",
		).Inc()
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

				bladeLabels := map[string]string{
					"blade_id": blade.ControlID,
				}

				bladeMetrics := map[string]interface{}{}

				if q.MemoryPerc != "" {
					memoryUsedPerc, err := strconv.ParseFloat(q.MemoryPerc, 64)
					if err != nil {
						slog.Error(
							"Failed to convert memory percentage from string",
							"value",
							q.MemoryPerc,
							"error",
							err,
						)
					} else {
						bladeMetrics["ruckus_blade_memory_used_percentage"] = memoryUsedPerc
					}
				}

				if q.DiskPerc != "" {
					diskUsedPerc, err := strconv.ParseFloat(q.DiskPerc, 64)
					if err != nil {
						slog.Error(
							"Failed to convert disk percentage from string",
							"value",
							q.DiskPerc,
							"error",
							err,
						)
					} else {
						bladeMetrics["ruckus_blade_storage_used_percentage"] = diskUsedPerc
					}
				}

				if q.CPUPerc != "" {
					cpuUsedPerc, err := strconv.ParseFloat(q.CPUPerc, 64)
					if err != nil {
						slog.Error(
							"Failed to convert cpu percentage from string",
							"value",
							q.CPUPerc,
							"error",
							err,
						)
					} else {
						bladeMetrics["ruckus_blade_cpu_used_percentage"] = cpuUsedPerc
					}
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
		instJSONUnparseableCounter.WithLabelValues(
			systemID,
			"system_configuration_system_summary",
		).Inc()
		slog.Error("Failed to convert system summary to JSON", "error", err)
	} else {
		clusterInfoLabels := map[string]string{
			"fw_version":    clusterSummary.Version,
			"cp_version":    clusterSummary.CpVersion,
			"ap_version":    clusterSummary.ApVersion,
			"model":         clusterSummary.Model,
			"serial_number": clusterSummary.SerialNumber,
			"ipv4_address":  clusterSummary.IPAddress,
			"blade_name":    clusterSummary.BladeName,
			"state":         clusterSummary.ClusterState,
			"blade_state":   clusterSummary.BladeState,
		}

		clusterMetrics := map[string]interface{}{
			"ruckus_cluster_state":                         parseClusterState(clusterSummary.ClusterState),
			"ruckus_cluster_blade_state":                   parseClusterState(clusterSummary.ClusterState),
			"ruckus_cluster_control_blades":                clusterSummary.NumOfControlBlades,
			"ruckus_cluster_control_blades_in_service":     clusterSummary.NumOfInServiceControlBlades,
			"ruckus_cluster_control_blades_out_of_service": clusterSummary.NumOfOutOfServiceControlBlades,
			"ruckus_cluster_data_blades":                   clusterSummary.NumOfDataBlades,
			"ruckus_cluster_data_blades_in_service":        clusterSummary.NumOfInServiceDataBlades,
			"ruckus_cluster_data_blades_out_of_service":    clusterSummary.NumOfOutOfServiceDataBlades,
			"ruckus_cluster_aps_connected":                 clusterSummary.NumOfConnectedAPs,
			"ruckus_cluster_aps_disconnected":              clusterSummary.NumOfDisconnectedAPs,
			"ruckus_cluster_aps_max":                       parseIntegerString(clusterSummary.MaxSupportedAps),

			"ruckus_cluster_uptime_seconds_total": clusterSummary.Uptime,

			"ruckus_cluster_info": 1,
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
				slog.Error("Error while appending metrics", "error", err.Error())
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

func handleZoneConfigurationMessage(systemID string, message *pb.ConfigurationMessage) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	var zoneConfig []ZoneConfig
	err := json.Unmarshal([]byte(clusterInfo.GetZones()), &zoneConfig)
	if err != nil {
		instJSONUnparseableCounter.WithLabelValues(systemID, "zone_configuration").Inc()
		slog.Error("Failed to convert zone configuration to JSON", "error", err)
	} else {
		for _, zone := range zoneConfig {
			zoneLabels := map[string]string{
				"zone_id":   zone.Key,
				"zone_name": zone.ZoneName,
			}

			zoneInfoLabels := map[string]string{
				"latitude":            parseLatitude(zone.GpsInfo),
				"longitude":           parseLongitude(zone.GpsInfo),
				"description":         zone.Description,
				"location":            zone.Location,
				"location_additional": zone.LocationAdditionalInfo,
			}

			zoneMetrics := map[string]interface{}{
				"ruckus_cluster_zone_info": 1,
			}

			labelMap := map[string]map[string]string{
				"ruckus_cluster_zone_info": zoneInfoLabels,
				"default":                  zoneLabels,
			}

			if errs := appendMetrics(
				timestamp,
				zoneMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err.Error())
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

func handleApGroupConfigurationMessage(systemID string, message *pb.ConfigurationMessage) error {
	timestamp := time.UnixMilli(int64(message.GetTimestamp()))

	clusterInfo := message.GetClusterInfo()

	metricsFamily := map[string]*dto.MetricFamily{}

	var apGroupConfig []ApGroupConfig
	err := json.Unmarshal([]byte(clusterInfo.GetApGroups()), &apGroupConfig)
	if err != nil {
		instJSONUnparseableCounter.WithLabelValues(systemID, "ap_group_configuration").Inc()
		slog.Error("Failed to convert ap group configuration to JSON", "error", err)
	} else {
		for _, apGroup := range apGroupConfig {
			apGroupLabels := map[string]string{
				"ap_group_id":   apGroup.Key,
				"ap_group_name": apGroup.Name,
			}

			apGroupMetrics := map[string]interface{}{
				"ruckus_ap_group_info": 1,
			}

			labelMap := map[string]map[string]string{
				"ruckus_ap_group_info": apGroupLabels,
			}

			if errs := appendMetrics(
				timestamp,
				apGroupMetrics,
				labelMap,
				metricsFamily,
			); len(errs) >= 1 {
				for _, err := range errs {
					slog.Error("Error while appending metrics", "error", err.Error())
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
