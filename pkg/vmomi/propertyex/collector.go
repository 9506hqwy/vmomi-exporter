package propertyex

import (
	"context"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func Retrieve(
	ctx context.Context,
	c *vim25.Client,
	roots []types.ManagedObjectReference,
	moTypes []string,
	pathSet []string,
	withRoot bool,
) ([]types.ObjectContent, error) {
	pc := property.DefaultCollector(c)

	objs := []types.ObjectSpec{}
	for _, r := range roots {
		objs = append(objs, TraverseChild(r, withRoot))
	}

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
