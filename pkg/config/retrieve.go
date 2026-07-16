package config

import (
	"go.yaml.in/yaml/v4"
)

type RetrieveConfig struct {
	IgnorDatastoreVM bool `yaml:"ignore_datastore_vm_relation"`
	IgnoreNetworkVM  bool `yaml:"ignore_network_vm_relation"`
}

func EncodeRetrieveConfig(c *RetrieveConfig) (string, error) {
	buf, err := yaml.Marshal(&c)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func DefaultRetrieveConfig() *RetrieveConfig {
	return &RetrieveConfig{
		IgnorDatastoreVM: false,
		IgnoreNetworkVM:  false,
	}
}
