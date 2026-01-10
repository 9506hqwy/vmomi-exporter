package propertyex

import (
	"github.com/vmware/govmomi/vim25/types"
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
	datastore := initSpec("ComputeResource", "datastore", cache)
	setSelectSet(datastore, createDatastoreLower, cache)

	host := initSpec("ComputeResource", "host", cache)
	setSelectSet(host, createHostSystemLower, cache)

	network := initSpec("ComputeResource", "network", cache)
	setSelectSet(network, createNetworkLower, cache)

	resourcePool := initSpec("ComputeResource", "resourcePool", cache)
	setSelectSet(resourcePool, createResourcePoolLower, cache)

	return []types.BaseSelectionSpec{
		datastore,
		host,
		network,
		resourcePool,
	}
}

func createDatacenterLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastoreFolder := initSpec("Datacenter", "datastoreFolder", cache)
	setSelectSet(datastoreFolder, createFolderLower, cache)

	hostFolder := initSpec("Datacenter", "hostFolder", cache)
	setSelectSet(hostFolder, createFolderLower, cache)

	networkFolder := initSpec("Datacenter", "networkFolder", cache)
	setSelectSet(networkFolder, createFolderLower, cache)

	vmFolder := initSpec("Datacenter", "vmFolder", cache)
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
		initSpec("Datastore", "vm", cache),
	}
}

func createDatastoreUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	return []types.BaseSelectionSpec{
		initSpec("Datastore", "host", cache),
	}
}

func createDistributedVirtualSwitchLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	portgroup := initSpec("DistributedVirtualSwitch", "portgroup", cache)
	setSelectSet(portgroup, createNetworkLower, cache)

	return []types.BaseSelectionSpec{
		portgroup,
	}
}

func createFolderLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	selectSet := func(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
		specs := make([]types.BaseSelectionSpec, 0)
		specs = append(specs, createFolderLower(cache)...)
		specs = append(specs, createComputeResourceLower(cache)...)
		specs = append(specs, createDatacenterLower(cache)...)
		specs = append(specs, createDatastoreLower(cache)...)
		specs = append(specs, createDistributedVirtualSwitchLower(cache)...)
		specs = append(specs, createNetworkLower(cache)...)

		return specs
	}

	childEntity := initSpec("Folder", "childEntity", cache)
	setSelectSet(childEntity, selectSet, cache)

	return []types.BaseSelectionSpec{
		childEntity,
	}
}

func createHostSystemLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastore := initSpec("HostSystem", "datastore", cache)
	setSelectSet(datastore, createDatastoreLower, cache)

	network := initSpec("HostSystem", "network", cache)
	setSelectSet(network, createNetworkLower, cache)

	vm := initSpec("HostSystem", "vm", cache)

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
		initSpec("Network", "vm", cache),
	}
}

func createNetworkUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	host := initSpec("Network", "host", cache)
	dvs := initSpec("DistributedVirtualPortgroup", "config.distributedVirtualSwitch", cache)

	return []types.BaseSelectionSpec{
		host,
		dvs,
	}
}

func createResourcePoolLower(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	resourcePool := initSpec("ResourcePool", "resourcePool", cache)
	setSelectSet(resourcePool, createResourcePoolLower, cache)

	vm := initSpec("ResourcePool", "vm", cache)

	return []types.BaseSelectionSpec{
		resourcePool,
		vm,
	}
}

func createVirtualMachineUpper(cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
	datastore := initSpec("VirtualMachine", "datastore", cache)
	setSelectSet(datastore, createDatastoreUpper, cache)

	network := initSpec("VirtualMachine", "network", cache)
	setSelectSet(network, createNetworkUpper, cache)

	parentVApp := initSpec("VirtualMachine", "parentVApp", cache)

	resourcePool := initSpec("VirtualMachine", "resourcePool", cache)

	host := initSpec("VirtualMachine", "runtime.host", cache)

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

func initSpec(moType string, path string, cache map[string]*types.TraversalSpec) types.BaseSelectionSpec {
	name := moType + "Spec" + path

	if _, found := cache[name]; found {
		return initSelection(name)
	}

	return initTraversal(name, moType, path, cache)
}

func initTraversal(name string, moType string, path string, cache map[string]*types.TraversalSpec) *types.TraversalSpec {
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

func setSelectSet(spec types.BaseSelectionSpec, selectSetFunc func(map[string]*types.TraversalSpec) []types.BaseSelectionSpec, cache map[string]*types.TraversalSpec) {
	if t, ok := spec.(*types.TraversalSpec); ok {
		t.SelectSet = selectSetFunc(cache)
	}
}

func traverseLower(moType string, cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {
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

func traverseUpper(moType string, cache map[string]*types.TraversalSpec) []types.BaseSelectionSpec {

	switch moType {
	case "ComputeResource":
		return createManagedEntityUpper(cache)
	case "Datacenter":
		return createManagedEntityUpper(cache)
	case "Datastore":
		return append(
			createManagedEntityUpper(cache),
			createDatastoreUpper(cache)...,
		)
	case "DistributedVirtualSwitch":
		return createManagedEntityUpper(cache)
	case "Folder":
		return createManagedEntityUpper(cache)
	case "HostSystem":
		return createManagedEntityUpper(cache)
	case "Network":
		return append(
			createManagedEntityUpper(cache),
			createNetworkUpper(cache)...,
		)
	case "ResourcePool":
		return createManagedEntityUpper(cache)
	case "VirtualMachine":
		return append(
			createManagedEntityUpper(cache),
			createVirtualMachineUpper(cache)...,
		)
	default:
		panic("Not supported")
	}
}
