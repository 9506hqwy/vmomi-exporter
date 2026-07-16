package propertyex

import (
	"context"

	"github.com/vmware/govmomi/vim25/types"
)

type IgnoreDatastoreVMKey struct{}
type IgnoreNetworkVMKey struct{}

const (
	ComputeResourceName          = "ComputeResource"
	DatacenterName               = "Datacenter"
	DatastoreName                = "Datastore"
	DistributedVirtualSwitchName = "DistributedVirtualSwitch"
	FolderName                   = "Folder"
	HostSystemName               = "HostSystem"
	NetworkName                  = "Network"
	ResourcePoolName             = "ResourcePool"
	VirtualMachineName           = "VirtualMachine"
)

const (
	DatastoreProperty    = "datastore"
	HostProperty         = "host"
	NetworkProperty      = "network"
	ResourcePoolProperty = "resourcePool"
	VMProperty           = "vm"
)

func TraverseChild(
	ctx context.Context,
	mo types.ManagedObjectReference,
	withMo bool,
) types.ObjectSpec {
	cache := make(map[string]*types.TraversalSpec)
	return types.ObjectSpec{
		Obj:       mo,
		SelectSet: traverseLower(ctx, mo.Type, cache),
		Skip:      types.NewBool(!withMo),
	}
}

func TraverseParent(
	ctx context.Context,
	mo types.ManagedObjectReference,
	withMo bool,
) types.ObjectSpec {
	cache := make(map[string]*types.TraversalSpec)
	return types.ObjectSpec{
		Obj:       mo,
		SelectSet: traverseUpper(ctx, mo.Type, cache),
		Skip:      types.NewBool(withMo),
	}
}

func createComputeResourceLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	datastore := initSpec(ctx, ComputeResourceName, DatastoreProperty, cache)
	setSelectSet(ctx, datastore, createDatastoreLower, cache)

	host := initSpec(ctx, ComputeResourceName, HostProperty, cache)
	setSelectSet(ctx, host, createHostSystemLower, cache)

	network := initSpec(ctx, ComputeResourceName, NetworkProperty, cache)
	setSelectSet(ctx, network, createNetworkLower, cache)

	resourcePool := initSpec(ctx, ComputeResourceName, ResourcePoolProperty, cache)
	setSelectSet(ctx, resourcePool, createResourcePoolLower, cache)

	return []types.BaseSelectionSpec{
		datastore,
		host,
		network,
		resourcePool,
	}
}

func createDatacenterLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	datastoreFolder := initSpec(ctx, DatacenterName, "datastoreFolder", cache)
	setSelectSet(ctx, datastoreFolder, createFolderLower, cache)

	hostFolder := initSpec(ctx, DatacenterName, "hostFolder", cache)
	setSelectSet(ctx, hostFolder, createFolderLower, cache)

	networkFolder := initSpec(ctx, DatacenterName, "networkFolder", cache)
	setSelectSet(ctx, networkFolder, createFolderLower, cache)

	vmFolder := initSpec(ctx, DatacenterName, "vmFolder", cache)
	setSelectSet(ctx, vmFolder, createFolderLower, cache)

	return []types.BaseSelectionSpec{
		datastoreFolder,
		hostFolder,
		networkFolder,
		vmFolder,
	}
}

func createDatastoreLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	skipDatastoreVMKey, ok := ctx.Value(IgnoreDatastoreVMKey{}).(bool)
	if ok && skipDatastoreVMKey {
		return []types.BaseSelectionSpec{}
	}

	return []types.BaseSelectionSpec{
		initSpec(ctx, DatastoreName, VMProperty, cache),
	}
}

func createDatastoreUpper(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	return []types.BaseSelectionSpec{
		initSpec(ctx, DatastoreName, HostProperty, cache),
	}
}

func createDistributedVirtualSwitchLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	portgroup := initSpec(ctx, DistributedVirtualSwitchName, "portgroup", cache)
	setSelectSet(ctx, portgroup, createNetworkLower, cache)

	return []types.BaseSelectionSpec{
		portgroup,
	}
}

func createFolderLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	selectSet := func(
		ctx context.Context,
		cache map[string]*types.TraversalSpec,
	) []types.BaseSelectionSpec {
		specs := []types.BaseSelectionSpec{}
		specs = append(specs, createFolderLower(ctx, cache)...)
		specs = append(specs, createComputeResourceLower(ctx, cache)...)
		specs = append(specs, createDatacenterLower(ctx, cache)...)
		specs = append(specs, createDatastoreLower(ctx, cache)...)
		specs = append(specs, createDistributedVirtualSwitchLower(ctx, cache)...)
		specs = append(specs, createNetworkLower(ctx, cache)...)

		return specs
	}

	childEntity := initSpec(ctx, FolderName, "childEntity", cache)
	setSelectSet(ctx, childEntity, selectSet, cache)

	return []types.BaseSelectionSpec{
		childEntity,
	}
}

func createHostSystemLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	datastore := initSpec(ctx, HostSystemName, DatastoreProperty, cache)
	setSelectSet(ctx, datastore, createDatastoreLower, cache)

	network := initSpec(ctx, HostSystemName, NetworkProperty, cache)
	setSelectSet(ctx, network, createNetworkLower, cache)

	vm := initSpec(ctx, HostSystemName, VMProperty, cache)

	return []types.BaseSelectionSpec{
		datastore,
		network,
		vm,
	}
}

func createManagedEntityUpper(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	parent := initSpec(ctx, "ManagedEntity", "parent", cache)
	setSelectSet(ctx, parent, createManagedEntityUpper, cache)

	return []types.BaseSelectionSpec{
		parent,
	}
}

func createNetworkLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	skipNetworkVMKey, ok := ctx.Value(IgnoreNetworkVMKey{}).(bool)
	if ok && skipNetworkVMKey {
		return []types.BaseSelectionSpec{}
	}

	return []types.BaseSelectionSpec{
		initSpec(ctx, NetworkName, VMProperty, cache),
	}
}

func createNetworkUpper(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	host := initSpec(ctx, NetworkName, HostProperty, cache)
	dvs := initSpec(ctx, "DistributedVirtualPortgroup", "config.distributedVirtualSwitch", cache)

	return []types.BaseSelectionSpec{
		host,
		dvs,
	}
}

func createResourcePoolLower(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	resourcePool := initSpec(ctx, ResourcePoolName, ResourcePoolProperty, cache)
	setSelectSet(ctx, resourcePool, createResourcePoolLower, cache)

	vm := initSpec(ctx, ResourcePoolName, VMProperty, cache)

	return []types.BaseSelectionSpec{
		resourcePool,
		vm,
	}
}

func createVirtualMachineUpper(
	ctx context.Context,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	datastore := initSpec(ctx, VirtualMachineName, DatastoreProperty, cache)
	setSelectSet(ctx, datastore, createDatastoreUpper, cache)

	network := initSpec(ctx, VirtualMachineName, NetworkProperty, cache)
	setSelectSet(ctx, network, createNetworkUpper, cache)

	parentVApp := initSpec(ctx, VirtualMachineName, "parentVApp", cache)

	resourcePool := initSpec(ctx, VirtualMachineName, ResourcePoolProperty, cache)

	host := initSpec(ctx, VirtualMachineName, "runtime.host", cache)

	return []types.BaseSelectionSpec{
		datastore,
		network,
		parentVApp,
		resourcePool,
		host,
	}
}

func initSelection(
	_ context.Context,
	name string,
) *types.SelectionSpec {
	return &types.SelectionSpec{
		Name: name,
	}
}

func initSpec(
	ctx context.Context,
	moType string,
	path string,
	cache map[string]*types.TraversalSpec,
) types.BaseSelectionSpec {
	name := moType + "Spec" + path

	if _, found := cache[name]; found {
		return initSelection(ctx, name)
	}

	return initTraversal(ctx, name, moType, path, cache)
}

func initTraversal(
	_ context.Context,
	name string,
	moType string,
	path string,
	cache map[string]*types.TraversalSpec,
) *types.TraversalSpec {
	spec := &types.TraversalSpec{
		SelectionSpec: types.SelectionSpec{
			Name: name,
		},
		Type: moType,
		Path: path,
		Skip: types.NewBool(false),
	}
	cache[name] = spec
	return spec
}

func setSelectSet(
	ctx context.Context,
	spec types.BaseSelectionSpec,
	selectSetFunc func(context.Context, map[string]*types.TraversalSpec) []types.BaseSelectionSpec,
	cache map[string]*types.TraversalSpec,
) {
	if t, ok := spec.(*types.TraversalSpec); ok {
		t.SelectSet = selectSetFunc(ctx, cache)
	}
}

//revive:disable:cyclomatic

func traverseLower(
	ctx context.Context,
	moType string,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	switch moType {
	case "ComputeResource":
		return createComputeResourceLower(ctx, cache)
	case "Datacenter":
		return createDatacenterLower(ctx, cache)
	case "Datastore":
		return createDatastoreLower(ctx, cache)
	case "DistributedVirtualSwitch":
		return createDistributedVirtualSwitchLower(ctx, cache)
	case "Folder":
		return createFolderLower(ctx, cache)
	case "HostSystem":
		return createHostSystemLower(ctx, cache)
	case "Network":
		return createNetworkLower(ctx, cache)
	case "ResourcePool":
		return createResourcePoolLower(ctx, cache)
	case "VirtualMachine":
		return nil
	default:
		panic("Not supported")
	}
}

//revive:enable:cyclomatic

func traverseUpper(
	ctx context.Context,
	moType string,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	switch moType {
	case
		"ComputeResource",
		"Datacenter",
		"DistributedVirtualSwitch",
		"Folder",
		"HostSystem",
		"ResourcePool":
		return createManagedEntityUpper(ctx, cache)
	case "Datastore":
		return append(
			createManagedEntityUpper(ctx, cache),
			createDatastoreUpper(ctx, cache)...,
		)
	case "Network":
		return append(
			createManagedEntityUpper(ctx, cache),
			createNetworkUpper(ctx, cache)...,
		)
	case "VirtualMachine":
		return append(
			createManagedEntityUpper(ctx, cache),
			createVirtualMachineUpper(ctx, cache)...,
		)
	default:
		panic("Not supported")
	}
}
