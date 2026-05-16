package config

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Onkyo    OnkyoConfig           `yaml:"onkyo"`
	Profiles map[string]ProfileCfg `yaml:"profiles"`
}

type OnkyoConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type ProfileCfg struct {
	Code           string `yaml:"code"`
	VolumeLevel    int    `yaml:"volumeLevel"`
	SubwooferLevel int    `yaml:"subwooferLevel"`
	MaxVolume      int    `yaml:"maxVolume"`
}

func defaults() *Config {
	return &Config{
		Onkyo: OnkyoConfig{
			Host: "10.205.0.163",
			Port: "60128",
		},
		Profiles: map[string]ProfileCfg{
			"tv":      {Code: "12", VolumeLevel: 22, SubwooferLevel: 0, MaxVolume: 28},
			"dj":      {Code: "10", VolumeLevel: 27, SubwooferLevel: -4, MaxVolume: 35},
			"vinyl":   {Code: "22", VolumeLevel: 20, SubwooferLevel: 0, MaxVolume: 30},
			"spotify": {Code: "01", VolumeLevel: 42, SubwooferLevel: 0, MaxVolume: 50},
		},
	}
}

func Load(path string) (*Config, error) {
	if path != "" {
		cfg, err := loadFile(path)
		if err == nil {
			return cfg, nil
		}
	}

	cfg, err := loadFile("/onkyo_config")
	if err == nil {
		return cfg, nil
	}

	return defaults(), nil
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Onkyo.Host == "" {
		cfg.Onkyo.Host = defaults().Onkyo.Host
	}
	if cfg.Onkyo.Port == "" {
		cfg.Onkyo.Port = defaults().Onkyo.Port
	}

	return &cfg, nil
}

func (c *Config) InputCodes() map[string]string {
	codes := make(map[string]string, len(c.Profiles))
	for name, p := range c.Profiles {
		codes[name] = p.Code
	}
	return codes
}

func (c *Config) InputNames() map[string]string {
	names := make(map[string]string, len(c.Profiles))
	for name, p := range c.Profiles {
		names[p.Code] = name
	}
	return names
}

func (c *Config) InputList() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
