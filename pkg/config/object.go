package config

import (
	"go.yaml.in/yaml/v4"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

var defaultTypes = []vmomi.ManagedEntityType{
	vmomi.ManagedEntityTypeHostSystem,
	vmomi.ManagedEntityTypeVirtualMachine,
}

type Object struct {
	Type *vmomi.ManagedEntityType `yaml:"type,omitempty"`
}

type ObjectConfig struct {
	Objects []Object `yaml:"objects"`
}

func EncodeObjects(o *[]Object) (string, error) {
	cc := ObjectConfig{
		Objects: *o,
	}

	buf, err := yaml.Marshal(&cc)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func DefaultObjectConfig() *ObjectConfig {
	objects := []Object{}
	for _, t := range defaultTypes {
		objects = append(objects, Object{
			Type: &t,
		})
	}

	return &ObjectConfig{
		Objects: objects,
	}
}
