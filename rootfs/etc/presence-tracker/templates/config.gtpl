state_path = "{{ .state_path }}"

mqtt {
  enabled = {{ .mqtt_enabled }}
  {{- if .mqtt_host }}
  host = "{{ .mqtt_host }}"
  {{ end }}
  {{- if .mqtt_port }}
  port = {{ .mqtt_port }}
  {{ end }}
  {{- if .mqtt_username }}
  username = "{{ .mqtt_username }}"
  {{ end }}
  {{- if .mqtt_password }}
  password = "{{ .mqtt_password }}"
  {{ end }}
}
