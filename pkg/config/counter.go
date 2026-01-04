package config

import (
	"go.yaml.in/yaml/v4"
)

var defaultCounters = []Counter{
	{
		Group:  "cpu",
		Name:   "usage",
		Rollup: "average",
	},
	{
		Group:  "cpu",
		Name:   "usagemhz",
		Rollup: "average",
	},
	{
		Group:  "mem",
		Name:   "usage",
		Rollup: "average",
	},
}

type Counter struct {
	Group  string `yaml:"group"`
	Name   string `yaml:"name"`
	Rollup string `yaml:"rollup"`
}

type CounterConfig struct {
	Counters []Counter `yaml:"counters"`
}

func EncodeCounters(c *[]Counter) (string, error) {
	cc := CounterConfig{
		Counters: *c,
	}

	buf, err := yaml.Marshal(&cc)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func DefaultCounterConfig() *CounterConfig {
	return &CounterConfig{
		Counters: defaultCounters,
	}
}
