package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/juli3nk/go-freebox"
	"github.com/juli3nk/kaza-presence-tracker-freebox/internal/config"
	"github.com/juli3nk/kaza-presence-tracker-freebox/internal/fbxapp"
	"github.com/juli3nk/kaza-presence-tracker-freebox/internal/mqtt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	flgConfig string
	flgDebug  bool
)

func init() {
	flag.StringVar(&flgConfig, "config", "/srv/presence-tracker/conf/config.hcl", "config file path")
	flag.BoolVar(&flgDebug, "debug", false, "enable debug log")

	flag.Parse()
}

func main() {
	mqttCli := mqtt.New()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flgDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	appID := "xyz.kaza.homepresence"

	cfg, err := config.New(flgConfig)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "config").
			Err(err).
			Send()
	}

	devices, err := freebox.Discover(freebox.DiscoverProtocolHTTP)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-discover").
			Err(err).
			Send()
	}
	dev := &devices[0]

	app, err := fbxapp.New(appID, dev, cfg.StatePath)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-app-new").
			Err(err).
			Send()
	}

	if err := app.Create("Presence Tracker", "0.1.0", "Kaza"); err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-app-create").
			Err(err).
			Send()
	}

	sessionToken, err := app.GetSessionToken()
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-get-session-token").
			Err(err).
			Send()
	}

	// MQTT
	if cfg.Mqtt.Enabled {
		if err = mqttCli.Init(cfg.Mqtt.Host, cfg.Mqtt.Port, cfg.Mqtt.Username, cfg.Mqtt.Password); err != nil {
			log.Fatal().
				Str("func", "main").
				Str("exec", "mqttInit").
				Err(err).
				Send()
		}
		log.Debug().Msg("connected to MQTT")
	}

	states := map[string]string{}

	probe := func() {
		log.Debug().Msg("probing")

		if cfg.Mqtt.Enabled && !mqttCli.IsConnected() {
			if err = mqttCli.Init(cfg.Mqtt.Host, cfg.Mqtt.Port, cfg.Mqtt.Username, cfg.Mqtt.Password); err != nil {
				log.Error().
					Str("func", "main").
					Str("exec", "mqttInit").
					Err(err).
					Send()
				return
			}
			log.Debug().Msg("connected to MQTT")
		}

		result, rerr, err := dev.DynamicLease(*sessionToken)
		if err != nil {
			log.Fatal().
				Str("func", "main").
				Str("freebox-device-dynamic-lease", "config").
				Err(err).
				Send()
		}
		if rerr != nil {
			sessionToken, err := app.GetSessionToken()
			if err != nil {
				log.Fatal().
					Str("func", "main").
					Str("exec", "freebox-get-session-token").
					Err(err).
					Send()
			}

			result, _, err = dev.DynamicLease(*sessionToken)
			if err != nil {
				log.Fatal().
					Str("func", "main").
					Str("freebox-device-dynamic-lease", "config").
					Err(err).
					Send()
			}
		}

		tmpStates := make(map[string]stateVal)

		for _, r := range result.Result {
			log.Debug().
				Msgf("%s (%s) %s => %s", r.Hostname, r.Mac, r.Host.HostType, r.Host.AccessPoint.ConnectivityType)

			if r.Host.HostType == "smartphone" {
				id := strings.Replace(r.Host.ID, "ether-", "", 1)
				id = strings.ReplaceAll(id, ":", "")

				status := "not_home"
				if r.Host.AccessPoint.ConnectivityType == "wifi" {
					status = "home"
				}

				tmpStates[id] = stateVal{
					mac:    strings.ToLower(r.Mac),
					name:   r.Hostname,
					status: status,
				}
			}
		}

		for k, v := range tmpStates {
			publish := false

			stateTopic := fmt.Sprintf("%s/state", k)

			if _, ok := states[k]; ok {
				if v.status != states[k] {
					publish = true
				}
			} else {
				publish = true

				if cfg.Mqtt.Enabled {
					topic := fmt.Sprintf("homeassistant/device_tracker/%s/config", k)

					payload := mqttPayload{
						StateTopic:     stateTopic,
						Name:           v.name,
						PayloadHome:    "home",
						PayloadNotHome: "not_home",
						Icon:           "mdi:cellphone",
						UniqueId:       k,
					}
					payloadJson, err := json.Marshal(payload)
					if err != nil {
						log.Error().
							Str("func", "main").
							Str("exec", "mqtt-payload-json").
							Err(err).
							Send()
						break
					}

					log.Info().Msgf("mqtt publish config for device %s", v.name)

					if err := mqttCli.Publish(topic, payloadJson); err != nil {
						log.Error().
							Str("func", "main").
							Str("exec", "mqtt-publish-device-config").
							Err(err).
							Send()
						break
					}
				}
			}

			if cfg.Mqtt.Enabled && publish {
				log.Info().Msgf("mqtt publish state for device %s (%s)", v.name, v.status)

				if err := mqttCli.Publish(stateTopic, []byte(v.status)); err != nil {
					log.Error().
						Str("func", "main").
						Str("exec", "mqtt-publish-device-state").
						Err(err).
						Send()
					break
				}

				states[k] = v.status
			}
		}
	}

	ticker := time.NewTicker(time.Duration(10) * time.Second)
	defer ticker.Stop() // Ensure ticker stops on exit

	done := make(chan struct{}) // Use a `struct{}` for signals (no data)

	probe()

	go func() {
		for {
			select {
			case <-done:
				return // Exit goroutine
			case <-ticker.C:
				probe()
			}
		}
	}()

	// Handle OS interrupts for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Wait for interrupt signal
	<-quit

	// Graceful shutdown
	log.Info().Msg("shutting down gracefully...")

	// Signal the goroutine to exit
	close(done) // Close the channel (non-blocking signal)

	// Stop ticker
	ticker.Stop()

	if cfg.Mqtt.Enabled {
		mqttCli.Disconnect(250)
	}

	log.Info().Msg("shutdown complete")
}
