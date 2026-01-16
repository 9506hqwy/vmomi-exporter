package propertyex

import (
	"github.com/vmware/govmomi/vim25/types"
)

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

func TraverseChild(mo types.ManagedObjectReference, withMo bool) types.ObjectSpec {
	cache := make(map[string]*types.TraversalSpec)
	return types.ObjectSpec{
		Obj:       mo,
		SelectSet: traverseLower(mo.Type, cache),
		Skip:      types.NewBool(!withMo),
	}
}

func TraverseParent(mo types.ManagedObjectReference, withMo bool) types.ObjectSpec {
	cache := make(map[string]*types.TraversalSpec)
	return types.ObjectSpec{
		Obj:       mo,
		SelectSet: traverseUpper(mo.Type, cache),
		Skip:      types.NewBool(withMo),
	}
}

func createComputeResourceLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastore := initSpec(ComputeResourceName, DatastoreProperty, cache)
	setSelectSet(datastore, createDatastoreLower, cache)

	host := initSpec(ComputeResourceName, HostProperty, cache)
	setSelectSet(host, createHostSystemLower, cache)

	network := initSpec(ComputeResourceName, NetworkProperty, cache)
	setSelectSet(network, createNetworkLower, cache)

	resourcePool := initSpec(ComputeResourceName, ResourcePoolProperty, cache)
	setSelectSet(resourcePool, createResourcePoolLower, cache)

	return []types.BaseSelectionSpec{
		datastore,
		host,
		network,
		resourcePool,
	}
}

func createDatacenterLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastoreFolder := initSpec(DatacenterName, "datastoreFolder", cache)
	setSelectSet(datastoreFolder, createFolderLower, cache)

	hostFolder := initSpec(DatacenterName, "hostFolder", cache)
	setSelectSet(hostFolder, createFolderLower, cache)

	networkFolder := initSpec(DatacenterName, "networkFolder", cache)
	setSelectSet(networkFolder, createFolderLower, cache)

	vmFolder := initSpec(DatacenterName, "vmFolder", cache)
	setSelectSet(vmFolder, createFolderLower, cache)

	return []types.BaseSelectionSpec{
		datastoreFolder,
		hostFolder,
		networkFolder,
		vmFolder,
	}
}

func createDatastoreLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	return []types.BaseSelectionSpec{
		initSpec(DatastoreName, VMProperty, cache),
	}
}

func createDatastoreUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	return []types.BaseSelectionSpec{
		initSpec(DatastoreName, HostProperty, cache),
	}
}

func createDistributedVirtualSwitchLower(
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	portgroup := initSpec(DistributedVirtualSwitchName, "portgroup", cache)
	setSelectSet(portgroup, createNetworkLower, cache)

	return []types.BaseSelectionSpec{
		portgroup,
	}
}

func createFolderLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	selectSet := func(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
		specs := []types.BaseSelectionSpec{}
		specs = append(specs, createFolderLower(cache)...)
		specs = append(specs, createComputeResourceLower(cache)...)
		specs = append(specs, createDatacenterLower(cache)...)
		specs = append(specs, createDatastoreLower(cache)...)
		specs = append(specs, createDistributedVirtualSwitchLower(cache)...)
		specs = append(specs, createNetworkLower(cache)...)

		return specs
	}

	childEntity := initSpec(FolderName, "childEntity", cache)
	setSelectSet(childEntity, selectSet, cache)

	return []types.BaseSelectionSpec{
		childEntity,
	}
}

func createHostSystemLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastore := initSpec(HostSystemName, DatastoreProperty, cache)
	setSelectSet(datastore, createDatastoreLower, cache)

	network := initSpec(HostSystemName, NetworkProperty, cache)
	setSelectSet(network, createNetworkLower, cache)

	vm := initSpec(HostSystemName, VMProperty, cache)

	return []types.BaseSelectionSpec{
		datastore,
		network,
		vm,
	}
}

func createManagedEntityUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	parent := initSpec("ManagedEntity", "parent", cache)
	setSelectSet(parent, createManagedEntityUpper, cache)

	return []types.BaseSelectionSpec{
		parent,
	}
}

func createNetworkLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	return []types.BaseSelectionSpec{
		initSpec(NetworkName, VMProperty, cache),
	}
}

func createNetworkUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	host := initSpec(NetworkName, HostProperty, cache)
	dvs := initSpec("DistributedVirtualPortgroup", "config.distributedVirtualSwitch", cache)

	return []types.BaseSelectionSpec{
		host,
		dvs,
	}
}

func createResourcePoolLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	resourcePool := initSpec(ResourcePoolName, ResourcePoolProperty, cache)
	setSelectSet(resourcePool, createResourcePoolLower, cache)

	vm := initSpec(ResourcePoolName, VMProperty, cache)

	return []types.BaseSelectionSpec{
		resourcePool,
		vm,
	}
}

func createVirtualMachineUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastore := initSpec(VirtualMachineName, DatastoreProperty, cache)
	setSelectSet(datastore, createDatastoreUpper, cache)

	network := initSpec(VirtualMachineName, NetworkProperty, cache)
	setSelectSet(network, createNetworkUpper, cache)

	parentVApp := initSpec(VirtualMachineName, "parentVApp", cache)

	resourcePool := initSpec(VirtualMachineName, ResourcePoolProperty, cache)

	host := initSpec(VirtualMachineName, "runtime.host", cache)

	return []types.BaseSelectionSpec{
		datastore,
		network,
		parentVApp,
		resourcePool,
		host,
	}
}

func initSelection(name string) *types.SelectionSpec {
	return &types.SelectionSpec{
		Name: name,
	}
}

func initSpec(
	moType string,
	path string,
	cache map[string]*types.TraversalSpec,
) types.BaseSelectionSpec {
	name := moType + "Spec" + path

	if _, found := cache[name]; found {
		return initSelection(name)
	}

	return initTraversal(name, moType, path, cache)
}

func initTraversal(
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
	spec types.BaseSelectionSpec,
	selectSetFunc func(map[string]*types.TraversalSpec) []types.BaseSelectionSpec,
	cache map[string]*types.TraversalSpec,
) {
	if t, ok := spec.(*types.TraversalSpec); ok {
		t.SelectSet = selectSetFunc(cache)
	}
}

//revive:disable:cyclomatic

func traverseLower(
	moType string,
	cache map[string]*types.TraversalSpec,
) []types.BaseSelectionSpec {
	switch moType {
	case "ComputeResource":
		return createComputeResourceLower(cache)
	case "Datacenter":
		return createDatacenterLower(cache)
	case "Datastore":
		return createDatastoreLower(cache)
	case "DistributedVirtualSwitch":
		return createDistributedVirtualSwitchLower(cache)
	case "Folder":
		return createFolderLower(cache)
	case "HostSystem":
		return createHostSystemLower(cache)
	case "Network":
		return createNetworkLower(cache)
	case "ResourcePool":
		return createResourcePoolLower(cache)
	case "VirtualMachine":
		return nil
	default:
		panic("Not supported")
	}
}

//revive:enable:cyclomatic

func traverseUpper(
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
		return createManagedEntityUpper(cache)
	case "Datastore":
		return append(
			createManagedEntityUpper(cache),
			createDatastoreUpper(cache)...,
		)
	case "Network":
		return append(
			createManagedEntityUpper(cache),
			createNetworkUpper(cache)...,
		)
	case "VirtualMachine":
		return append(
			createManagedEntityUpper(cache),
			createVirtualMachineUpper(cache)...,
		)
	default:
		panic("Not supported")
	}
}
