package propertyex

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func RetrieveFromRoot(ctx context.Context, c *vim25.Client, moTypes []string, pathSet []string) ([]types.ObjectContent, error) {
	pc := property.DefaultCollector(c)

	objs := []types.ObjectSpec{TraverseChild(c.ServiceContent.RootFolder)}

	props := []types.PropertySpec{}
	for _, moType := range moTypes {
		spec := types.PropertySpec{
			Type:    moType,
			PathSet: pathSet,
		}

		props = append(props, spec)
	}

	filter := types.PropertyFilterSpec{
		ObjectSet: objs,
		PropSet:   props,
	}

	req := types.RetrieveProperties{
		SpecSet: []types.PropertyFilterSpec{filter},
	}

	res, err := pc.RetrieveProperties(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Returnval, nil
}
