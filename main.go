package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	pb "github.com/adaricorp/ruckus-sz-proto"

	"github.com/eclipse/paho.golang/paho"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"google.golang.org/protobuf/proto"
)

const (
	binName = "ruckus_sz_agent"
	timeout = 10 * time.Second
)

var (
	version = "dev"
	date    = "unknown"

	pahoDebugLogger           *log.Logger
	pahoErrorLogger           *log.Logger
	logLevel                  *string
	slogLevel                 *slog.LevelVar = new(slog.LevelVar)
	lokiServer                *string
	lokiMetricId              *string
	mqttURI                   *string
	mqttTopic                 *string
	mqttClientID              *string
	mqttQoS                   *int
	mqttKeepAlive             *int
	mqttSessionExpiryInterval *int
	mqttCleanStart            *bool
	mqttUsername              *string
	mqttPassword              *string
	prometheusRemoteWriteURI  *string

	loki lokiWriteClient
	prom prometheusWriteClient
)

// Print program usage
func printUsage(fs ff.Flags) {
	fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
	os.Exit(1)
}

// Print program version
func printVersion() {
	fmt.Printf("%s v%s built on %s\n", binName, version, date)
	os.Exit(0)
}

func init() {
	fs := ff.NewFlagSet(binName)
	displayVersion := fs.BoolLong("version", "Print version")
	logLevel = fs.StringEnumLong(
		"log-level",
		"Log level: debug, info, warn, error",
		"info",
		"debug",
		"error",
		"warn",
	)
	lokiServer = fs.StringLong("loki-server", "localhost:9096", "Loki server to connect to")
	lokiMetricId = fs.StringLong(
		"loki-metric-id",
		"ruckus",
		"Name to uniquely identify event streams in loki",
	)
	mqttURI = fs.StringLong(
		"mqtt-server",
		"mqtt://127.0.0.1:1883",
		"MQTT server to connect to",
	)
	mqttTopic = fs.StringLong(
		"mqtt-topic",
		"sci-topic",
		"MQTT topic to subscribe to",
	)
	mqttClientID = fs.StringLong(
		"mqtt-client-id",
		binName,
		"Unique client id for the MQTT connection",
	)
	mqttQoS = fs.IntLong(
		"mqtt-qos",
		2,
		"Quality of service level for the MQTT connection",
	)
	mqttKeepAlive = fs.IntLong(
		"mqtt-keep-alive",
		20,
		"Keepalive period in seconds for the MQTT connection",
	)
	mqttSessionExpiryInterval = fs.IntLong(
		"mqtt-session-expiry-interval",
		60,
		"Session expiry interval in seconds for the MQTT connection",
	)
	mqttCleanStart = fs.BoolLong(
		"mqtt-clean-start",
		"Setting this to true will clear the session on the first MQTT connection",
	)
	mqttUsername = fs.StringLong("mqtt-username", "", "The username for the MQTT connection")
	mqttPassword = fs.StringLong("mqtt-password", "", "The password for the MQTT connection")
	prometheusRemoteWriteURI = fs.StringLong(
		"prometheus-remote-write-uri",
		"http://localhost:9090/api/v1/write",
		"Prometheus URI to remote write metrics to",
	)

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix(strings.ToUpper(binName)),
		ff.WithEnvVarSplit(" "),
	)
	if err != nil {
		printUsage(fs)
	}

	if *displayVersion {
		printVersion()
	}

	switch *logLevel {
	case "debug":
		slogLevel.Set(slog.LevelDebug)
	case "info":
		slogLevel.Set(slog.LevelInfo)
	case "warn":
		slogLevel.Set(slog.LevelWarn)
	case "error":
		slogLevel.Set(slog.LevelError)
	}

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slogLevel,
		}),
	)
	slog.SetDefault(logger)

	pahoDebugLogger = slog.NewLogLogger(logger.Handler(), slog.LevelDebug)
	if *logLevel == "debug" {
		pahoErrorLogger = slog.NewLogLogger(logger.Handler(), slog.LevelDebug)
	}
}

func main() {
	var err error

	slog.Info(
		fmt.Sprintf("Starting %s", binName),
		"version",
		version,
		"build_context",
		fmt.Sprintf(
			"go=%s, platform=%s",
			runtime.Version(),
			runtime.GOOS+"/"+runtime.GOARCH,
		),
	)

	prom, err = newPrometheusWriteClient(*prometheusRemoteWriteURI, timeout)
	if err != nil {
		slog.Error(
			"Failed to create prometheus remote write client",
			"error",
			err,
		)
		os.Exit(1)
	}

	loki, err = newLokiWriteClient(*lokiServer, timeout)
	if err != nil {
		slog.Error(
			"Failed to create loki write client",
			"error",
			err,
		)
		os.Exit(1)
	}

	mqttChannel := make(chan *paho.Publish)

	mqttClientConfig := paho.ClientConfig{
		ClientID: *mqttClientID,
		OnPublishReceived: []func(paho.PublishReceived) (bool, error){
			func(pr paho.PublishReceived) (bool, error) {
				mqttChannel <- pr.Packet
				return true, nil
			}},
		OnClientError: func(err error) {
			slog.Error("MQTT client error", "error", err)
		},
		OnServerDisconnect: func(d *paho.Disconnect) {
			if d.Properties != nil {
				slog.Error(
					"MQTT server disconnected",
					"reason",
					d.Properties.ReasonString,
				)
			} else {
				slog.Error(
					"MQTT server disconnected",
					"reason",
					d.ReasonCode,
				)
			}
		},
	}

	mqttConnection, err := newMQTTConnection(
		*mqttURI,
		*mqttUsername,
		*mqttPassword,
		*mqttTopic,
		byte(*mqttQoS),
		pahoErrorLogger,
		pahoDebugLogger,
		mqttClientConfig,
	)
	if err != nil {
		slog.Error("Failed to create MQTT connection", "error", err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		slog.Info("Signal received, exiting")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := mqttConnection.Disconnect(ctx); err != nil {
			slog.Error("Failed to disconnect from MQTT server", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	for message := range mqttChannel {
		sciMessage := &pb.SciMessage{}
		if err := proto.Unmarshal(message.Payload, sciMessage); err != nil {
			slog.Error("Failed to parse MQTT message", "error", err)
			continue
		}

		systemId := sciMessage.GetSciSystemId()

		if event := sciMessage.GetEventMessage(); event != nil {
			if *lokiServer == "" {
				continue
			}
			slog.Debug("Starting to process event message")
			if err := handleEvent(systemId, event); err != nil {
				slog.Error("Error processing event message", "error", err)
			} else {
				slog.Debug("Finished processing event message")
			}
		} else if apStatus := sciMessage.GetApStatus(); apStatus != nil {
			if *prometheusRemoteWriteURI == "" {
				continue
			}
			slog.Debug("Starting to process ap status message")
			if err := handleApStatus(systemId, apStatus); err != nil {
				slog.Error("Error processing ap status message", "error", err)
			}
			slog.Debug("Finished processing ap status message")
		} else if apClient := sciMessage.GetApClient(); apClient != nil {
			if *prometheusRemoteWriteURI == "" {
				continue
			}
			slog.Debug("Starting to process ap client message")
			if err := handleApClient(systemId, apClient); err != nil {
				slog.Error("Error processing ap client message", "error", err)
			}
			slog.Debug("Finished processing ap client message")
		} else if apWiredClient := sciMessage.GetApWiredClient(); apWiredClient != nil {
			if *prometheusRemoteWriteURI == "" {
				continue
			}
			slog.Debug("Starting to process ap wired client message")
			if err := handleApWiredClient(systemId, apWiredClient); err != nil {
				slog.Error("Error processing ap wired client message", "error", err)
			}
			slog.Debug("Finished processing ap wired client message")
		} else if apReport := sciMessage.GetApReport(); apReport != nil {
			if *prometheusRemoteWriteURI == "" {
				continue
			}
			slog.Debug("Starting to process ap report message")
			if err := handleApReport(systemId, apReport); err != nil {
				slog.Error("Error processing ap report message", "error", err)
			}
			slog.Debug("Finished processing ap report message")
		} else if configMessage := sciMessage.GetConfigurationMessage(); configMessage != nil {
			if *prometheusRemoteWriteURI == "" {
				continue
			}
			clusterMessage := configMessage.GetClusterInfo()

			if clusterMessage.GetAps() != "" {
				slog.Debug("Starting to process cluster ap configuration message")
				if err := handleApConfigurationMessage(systemId, configMessage); err != nil {
					slog.Error("Error processing cluster ap configuration message", "error", err)
				}
				slog.Debug("Finished processing cluster ap configuration message")
			}

			if clusterMessage.GetControlBlades() != "" {
				slog.Debug("Starting to process cluster configuration message")
				if err := handleSystemConfigurationMessage(systemId, configMessage); err != nil {
					slog.Error("Error processing cluster configuration message", "error", err)
				}
				slog.Debug("Finished processing cluster configuration message")
			}
		} else {
			if *logLevel == "debug" {
				jsonSciMessage, err := json.Marshal(sciMessage)
				if err != nil {
					slog.Error("Failed to convert message to JSON", "error", err)
				}
				slog.Debug("Unhandled ruckus message", "message", string(jsonSciMessage))
			}
		}
	}
}
