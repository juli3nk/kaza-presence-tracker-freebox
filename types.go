package main

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
