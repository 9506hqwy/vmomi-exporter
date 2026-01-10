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

func GetEntity(ctx context.Context, rootEntities []Entity, entityTypes []ManagedEntityType, withRoot bool) (*[]Entity, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	roots := []types.ManagedObjectReference{}
	for _, e := range rootEntities {
		mor := types.ManagedObjectReference{
			Type:  string(e.Type),
			Value: e.Id,
		}
		roots = append(roots, mor)
	}

	moTypes := []string{}
	for _, t := range entityTypes {
		moTypes = append(moTypes, string(t))
	}

	entities, err := getEntity(ctx, c, roots, moTypes, withRoot)
	if err != nil {
		return nil, err
	}

	info := toEntitiesFromManageds(entities)
	return info, nil
}

func GetEntityFromRoot(ctx context.Context, entityTypes []ManagedEntityType) (*[]Entity, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	moTypes := []string{}
	for _, t := range entityTypes {
		moTypes = append(moTypes, string(t))
	}

	roots := []types.ManagedObjectReference{c.ServiceContent.RootFolder}
	entities, err := getEntity(ctx, c, roots, moTypes, false)
	if err != nil {
		return nil, err
	}

	info := toEntitiesFromManageds(entities)
	return info, nil
}

func getEntity(ctx context.Context, c *vim25.Client, roots []types.ManagedObjectReference, types []string, withRoot bool) (*[]mo.ManagedEntity, error) {
	objects, err := px.Retrieve(ctx, c, roots, types, []string{"name"}, withRoot)
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

func toEntitiesFromManageds(es *[]mo.ManagedEntity) *[]Entity {
	entities := []Entity{}
	for _, e := range *es {
		entities = append(entities, toEntityFromManaged(e))
	}

	return &entities
}

func toEntityFromManaged(e mo.ManagedEntity) Entity {
	return Entity{
		Id:   e.Reference().Value,
		Name: e.Name,
		Type: ManagedEntityType(e.Reference().Type),
	}
}
