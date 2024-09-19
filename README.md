# ruckus-sz-agent

This tool can ingest streaming network telemetry and events from Ruckus SmartZone (SZ)
and push this data to a time series database (TSDB).

The supported TSDB APIs are:

  * Prometheus remote write (for metrics)
  * Loki gRPC (for events)

## Requirements

### MQTT broker

The following guide documents how to configure a Mosquitto MQTT broker
to receive data from a Ruckus SZ:

https://docs.commscope.com/bundle/sz-700-gpbmqttinterfacegettingstartedguide-sz300,sz100,sz100d,vsz,vszd/page/GUID-8EF93C42-A3B4-4D22-9D69-8DE3CA4136B9.html

### Ruckus SZ configuration

The following guide documents how to configure a Ruckus SZ to send data to the MQTT broker:

https://docs.commscope.com/bundle/sz-700-gpbmqttinterfacegettingstartedguide-sz300,sz100,sz100d,vsz,vszd/page/GUID-79F075AA-2AD7-4E94-8B9F-4A21393286AC.html

The following message types are supported:

  * ApClientStatus
  * ApReport
  * ApStatus
  * ApWiredClient
  * Event
  * SzApConf
  * SzSystemConf
  * SzZoneConf

## Downloading

Download prebuilt binaries from [GitHub](https://github.com/adaricorp/ruckus-sz-agent/releases/latest).

## Running

To connect ruckus-sz-agent with an MQTT server running at 127.0.0.1:1883 with no username/password
and push network telemetry to a prometheus and loki endpoint:

```
ruckus_sz_agent \
    --mqtt-server mqtt://127.0.0.1:1883 \
    --prometheus-remote-write-uri http://127.0.0.1:9090/api/v1/write \
    --loki-server 127.0.0.1:9096
```

Either prometheus or loki can be disabled by setting their destination to an empty string, e.g:

```
ruckus_sz_agent \
    --mqtt-server mqtt://127.0.0.1:1883 \
    --prometheus-remote-write-uri http://127.0.0.1:9090/api/v1/write \
    --loki-server ""
```

It is also possible to configure ruckus-sz-agent by using environment variables:

```
RUCKUS_SZ_AGENT_MQTT_SERVER="mqtt://127.0.0.1:1883" ruckus_sz_agent
```

## Metrics

### Prometheus

All metric names for prometheus start with `ruckus_`.

### Loki

All event streams for loki have `{service="ruckus"}` as a label.
