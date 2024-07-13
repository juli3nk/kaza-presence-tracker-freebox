package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/juli3nk/go-freebox"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type stateVal struct {
	mac    string
	name   string
	status string
}

// type hassDevice struct {
// 	Connections []string `json:"connections"`
// 	Name        string   `json:"name"`
// }

type mqttPayload struct {
	StateTopic     string `json:"state_topic"`
	Name           string `json:"name"`
	PayloadHome    string `json:"payload_home"`
	PayloadNotHome string `json:"payload_not_home"`
	//Device              hassDevice `json:"device,omitempty"`
	Icon                string `json:"icon"`
	UniqueId            string `json:"unique_id"`
	JsonAttributesTopic string `json:"json_attributes_topic,omitempty"`
}

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
	var mqttCli mqtt.Client

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flgDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	appID := "xyz.kaza.homepresence"

	cfg, err := NewConfig(flgConfig)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "config").
			Err(err).
			Send()
	}

	st, err := newState(cfg.StatePath)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "newState").
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

	if err := getAppToken(st, dev, appID); err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-get-app-token").
			Err(err).
			Send()
	}

	sessionToken, err := getSessionToken(st, dev, appID)
	if err != nil {
		log.Fatal().
			Str("func", "main").
			Str("exec", "freebox-get-session-token").
			Err(err).
			Send()
	}

	states := map[string]string{}

	// MQTT
	if cfg.Mqtt.Enabled {
		mqttCli, err = mqttInit(cfg.Mqtt.Host, cfg.Mqtt.Port, cfg.Mqtt.Username, cfg.Mqtt.Password)
		if err != nil {
			log.Fatal().
				Str("func", "main").
				Str("exec", "mqttInit").
				Err(err).
				Send()
		}
		log.Debug().Msg("connected to MQTT")
	}

	probe := func() {
		log.Debug().Msg("probing")

		result, rerr, err := dev.DynamicLease(*sessionToken)
		if err != nil {
			log.Fatal().
				Str("func", "main").
				Str("freebox-device-dynamic-lease", "config").
				Err(err).
				Send()
		}
		if rerr != nil {
			if err := st.removeSessionTokenFile(); err != nil {
				log.Fatal().
					Str("func", "main").
					Str("exec", "freebox-remove-session-token").
					Err(err).
					Send()
			}

			sessionToken, err := getSessionToken(st, dev, appID)
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

		if cfg.Mqtt.Enabled && !mqttCli.IsConnected() {
			mqttCli, err = mqttInit(cfg.Mqtt.Host, cfg.Mqtt.Port, cfg.Mqtt.Username, cfg.Mqtt.Password)
			if err != nil {
				log.Error().
					Str("func", "main").
					Str("exec", "mqttInit").
					Err(err).
					Send()
				return
			}
			log.Debug().Msg("connected to MQTT")
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

					mqttPublish(mqttCli, topic, payloadJson)
				}
			}

			if cfg.Mqtt.Enabled && publish {
				log.Info().Msgf("mqtt publish state for device %s (%s)", v.name, v.status)

				mqttPublish(mqttCli, stateTopic, []byte(v.status))
				states[k] = v.status
			}
		}
	}

	ticker := time.NewTicker(time.Duration(10) * time.Second)
	done := make(chan interface{})

	probe()

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				probe()
			}
		}
	}()

	// Exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	if cfg.Mqtt.Enabled {
		mqttCli.Disconnect(250)
	}

	ticker.Stop()
	done <- true
}
