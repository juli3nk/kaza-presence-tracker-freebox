#!/usr/bin/env bashio
# shellcheck shell=bash
# shellcheck disable=SC2155

PT_CONFIG_PATH="/etc/presence-tracker"


debug="$(bashio::config 'debug')"
state_path="$(bashio::config 'state_path')"
mqtt_enabled="$(bashio::config 'mqtt.enabled')"
mqtt_host=""
mqtt_port=""
mqtt_username=""
mqtt_password=""

mkdir -p "$state_path"

if bashio::config.true 'mqtt.enabled'; then
  if bashio::config.true 'mqtt.autodiscover'; then
    bashio::log.info "mqtt.autodiscover is defined in options, attempting autodiscovery..."
    if ! bashio::services.available "mqtt"; then
      bashio::exit.nok "No internal MQTT service found. Please install Mosquitto broker"
    fi

    bashio::log.info "... MQTT service found, fetching server detail ..."
    mqtt_host="$(bashio::services mqtt "host")"
    mqtt_port="$(bashio::services mqtt "port")"
    mqtt_username="$(bashio::services mqtt "username")"
    mqtt_password="$(bashio::services mqtt "password")"
  else
    mqtt_host="$(bashio::config 'mqtt.host')"
    mqtt_port="$(bashio::config 'mqtt.port')"
    mqtt_username="$(bashio::config 'mqtt.username')"
    mqtt_password="$(bashio::config 'mqtt.password')"
  fi
fi

bashio::var.json \
	state_path "$state_path" \
  mqtt_enabled "$mqtt_enabled" \
	mqtt_host "$mqtt_host" \
	mqtt_port "$mqtt_port" \
	mqtt_username "$mqtt_username" \
	mqtt_password "$mqtt_password" \
	| tempio \
		-template "${PT_CONFIG_PATH}/templates/config.gtpl" \
		-out "${PT_CONFIG_PATH}/config.hcl"


bashio::log.info "Starting the app"
/usr/local/bin/presence-tracker \
	-config "${PT_CONFIG_PATH}/config.hcl" \
	-debug "$debug" \
	|| bashio::log.fatal "The app has crashed. Are you sure you entered the correct config options?"
