package presets

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type ServiceConfiguration struct {
	Name     string   `json:"name"`
	NameLong string   `json:"nameLong"`
	Image    string   `json:"image"`
	MinMem   int      `json:"minMem"`
	RecMem   int      `json:"recMem"`
	Env      []Env    `json:"env"`
	Ports    []Port   `json:"ports"`
	Volumes  []Volume `json:"volumes"`
}

type Env struct {
	Name        string  `json:"name"`
	Key         string  `json:"key"`
	Required    bool    `json:"required"`
	Description string  `json:"description"`
	Default     string  `json:"default"`
	Values      []Value `json:"values"`
}

type Value struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Port struct {
	Host      int64  `json:"host"`
	Container int64  `json:"container"`
	Protocol  string `json:"protocol"`
}

type Volume struct {
	Host        string `json:"host"`
	Destination string `json:"destination"`
}

func GetServiceConfigurations() (map[string]ServiceConfiguration, error) {
	// Open the JSON file
	jsonFile, err := os.Open("./service/presets/service-configurations.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open json file: %v", err)
	}
	defer jsonFile.Close()

	// Read the JSON file content
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read json file: %v", err)
	}

	// Unmarshal the JSON content into ServiceConfiguration struct
	var configs map[string]ServiceConfiguration
	if err := json.Unmarshal(byteValue, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json content: %v", err)
	}

	return configs, nil
}

func GetServiceConfiguration(serviceNameID string) (ServiceConfiguration, error) {
	configs, err := GetServiceConfigurations()
	if err != nil {
		return ServiceConfiguration{}, fmt.Errorf("failed to get server configs: %s: %v", serviceNameID, err)
	}

	serviceConfig, exists := configs[serviceNameID]
	if !exists {
		return ServiceConfiguration{}, fmt.Errorf("config for this service name id not found: %s: %v", serviceNameID, err)
	}

	return serviceConfig, nil
}
