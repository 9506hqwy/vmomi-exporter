package config

import (
	"go.yaml.in/yaml/v4"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

type Instance struct {
	EntityType vmomi.ManagedEntityType `yaml:"entity_type"`
	EntityID   string                  `yaml:"entity_id"`
	EntityName string                  `yaml:"entity_name"`
	Instance   string                  `yaml:"instance"`
	CounterID  int32                   `yaml:"counter_id"`
}

type InstanceConfig struct {
	Instances []Instance `yaml:"instances"`
}

func EncodeInstances(i *[]Instance) (string, error) {
	cc := InstanceConfig{
		Instances: *i,
	}

	buf, err := yaml.Marshal(&cc)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
