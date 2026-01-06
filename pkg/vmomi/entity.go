package vmomi

import (
	"context"

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	px "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/propertyex"
	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

type ManagedEntityType string

type Entity struct {
	Id   string
	Name string
	Type ManagedEntityType
}

const (
	ManagedEntityTypeClusterComputeResource         = ManagedEntityType("ClusterComputeResource")
	ManagedEntityTypeComputeResource                = ManagedEntityType("ComputeResource")
	ManagedEntityTypeDatacenter                     = ManagedEntityType("Datacenter")
	ManagedEntityTypeDatastore                      = ManagedEntityType("Datastore")
	ManagedEntityTypeDistributedVirtualSwitch       = ManagedEntityType("DistributedVirtualSwitch")
	ManagedEntityTypeFolder                         = ManagedEntityType("Folder")
	ManagedEntityTypeHostSystem                     = ManagedEntityType("HostSystem")
	ManagedEntityTypeNetwork                        = ManagedEntityType("Network")
	ManagedEntityTypeResourcePool                   = ManagedEntityType("ResourcePool")
	ManagedEntityTypeStoragePod                     = ManagedEntityType("StoragePod")
	ManagedEntityTypeVirtualApp                     = ManagedEntityType("VirtualApp")
	ManagedEntityTypeVirtualMachine                 = ManagedEntityType("VirtualMachine")
	ManagedEntityTypeVmwareDistributedVirtualSwitch = ManagedEntityType("VmwareDistributedVirtualSwitch")
)

func ManagedEntityTypeValues() []ManagedEntityType {
	return []ManagedEntityType{
		ManagedEntityTypeClusterComputeResource,
		ManagedEntityTypeComputeResource,
		ManagedEntityTypeDatacenter,
		ManagedEntityTypeDatastore,
		ManagedEntityTypeDistributedVirtualSwitch,
		ManagedEntityTypeFolder,
		ManagedEntityTypeHostSystem,
		ManagedEntityTypeNetwork,
		ManagedEntityTypeResourcePool,
		ManagedEntityTypeStoragePod,
		ManagedEntityTypeVirtualApp,
		ManagedEntityTypeVirtualMachine,
	}
}

func ManagedEntityTypeStrings() []string {
	values := []string{}
	for _, v := range ManagedEntityTypeValues() {
		values = append(values, string(v))
	}
	return values
}

func GetEntity(ctx context.Context, types []ManagedEntityType) (*[]Entity, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	moTypes := []string{}
	for _, t := range types {
		moTypes = append(moTypes, string(t))
	}

	entities, err := getEntity(ctx, c, moTypes)
	if err != nil {
		return nil, err
	}

	info := []Entity{}
	for _, e := range *entities {
		info = append(info, Entity{
			Id:   e.Reference().Value,
			Name: e.Name,
			Type: ManagedEntityType(e.Reference().Type),
		})
	}

	return &info, nil
}

func getEntity(ctx context.Context, c *vim25.Client, types []string) (*[]mo.ManagedEntity, error) {
	objects, err := px.RetrieveFromRoot(ctx, c, types, []string{"name"})
	if err != nil {
		return nil, err
	}

	entities := []mo.ManagedEntity{}
	if err := mo.LoadObjectContent(objects, &entities); err != nil {
		return nil, err
	}

	return &entities, nil
}

func findEntityName(entities *[]mo.ManagedEntity, mor types.ManagedObjectReference) string {
	for _, e := range *entities {
		if e.Reference().Value == mor.Value && e.Reference().Type == mor.Type {
			return e.Name
		}
	}

	return ""
}
