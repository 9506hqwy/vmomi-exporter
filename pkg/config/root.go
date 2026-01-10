package config

import (
	"go.yaml.in/yaml/v4"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

type Root struct {
	Type vmomi.ManagedEntityType `yaml:"type"`
	Name string                  `yaml:"name"`
}

type RootConfig struct {
	Roots []Root `yaml:"roots"`
}

func EncodeRoots(r *[]Root) (string, error) {
	cc := RootConfig{
		Roots: *r,
	}

	buf, err := yaml.Marshal(&cc)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func DefaultRootConfig() *RootConfig {
	roots := []Root{{
		Type: vmomi.ManagedEntityTypeFolder,
		Name: "",
	}}

	return &RootConfig{
		Roots: roots,
	}
}
