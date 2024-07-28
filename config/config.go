package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Host struct {
	BaseURL   string `json:"baseurl"`
	Password  string `json:"password"`
	Path      string `json:"path"`
	FullURL   string
	SslSecure bool   `json:"sslSecure"`
}

type Config struct {
	PrimaryHost    Host   `json:"primaryhost"`
	SecondaryHosts []Host `json:"secondaryHosts"`
	UpdateGravity  bool   `json:"updateGravity"`
	RunOnce        bool   `json:"runOnce"`
	IntervalMinutes int   `json:"intervalMinutes"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal the config file. You might have a json formatting issue: %v", err)
	}
	return &config, nil
}

func (cfg *Config) SecondaryHostsAsStringSlice() []string {
	hosts := make([]string, len(cfg.SecondaryHosts))
	for i, host := range cfg.SecondaryHosts {
		hosts[i] = host.BaseURL
	}
	return hosts
}
