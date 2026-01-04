package vmomi

import (
	"context"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	px "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/propertyex"
	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

type Entity struct {
	Id   string
	Name string
	Type ManagedEntityType
}

type Metric struct {
	Entity    Entity
	Counter   CounterInfo
	Instance  string
	Timestamp time.Time
	Value     int64
}

type InstanceInfo struct {
	EntityType ManagedEntityType
	EntityId   string
	EntityName string
	Instance   string
	CounterId  int32
}

func GetInstanceInfo(ctx context.Context, types []ManagedEntityType) (*[]InstanceInfo, error) {
	url, user, password, noVerifySSL, err := GetTarget(ctx)
	if err != nil {
		return nil, err
	}

	c, err := sx.Login(ctx, url, user, password, noVerifySSL)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	pm := performance.NewManager(c)

	moTypes := []string{}
	for _, t := range types {
		moTypes = append(moTypes, string(t))
	}

	entities, err := getEntity(ctx, c, moTypes)
	if err != nil {
		return nil, err
	}

	specs, err := createQuerySpecs(ctx, pm, entities, nil)
	if err != nil {
		return nil, err
	}

	info := []InstanceInfo{}
	for _, spec := range *specs {
		entityName := findEntityName(entities, spec.Entity)
		for _, metricId := range spec.MetricId {

			info = append(info, *ToInstanceInfo(&metricId, spec.Entity, entityName))
		}
	}

	return &info, nil

}

func ToInstanceInfo(c *types.PerfMetricId, mor types.ManagedObjectReference, name string) *InstanceInfo {
	return &InstanceInfo{
		EntityType: ManagedEntityType(mor.Type),
		EntityId:   mor.Value,
		EntityName: name,
		Instance:   c.Instance,
		CounterId:  c.CounterId,
	}
}

func Query(ctx context.Context, moTypes []string, counters []CounterInfo) ([]Metric, error) {
	url, user, password, noVerifySSL, err := GetTarget(ctx)
	if err != nil {
		return nil, err
	}

	c, err := sx.Login(ctx, url, user, password, noVerifySSL)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	pc := property.DefaultCollector(c)
	pm := performance.NewManager(c)

	var p mo.PerformanceManager
	err = pc.RetrieveOne(ctx, *c.ServiceContent.PerfManager, nil, &p)
	if err != nil {
		return nil, err
	}

	cnts := []CounterInfo{}
	for _, c := range counters {
		c := ComplementCounterInfo(p, c)
		// TODO: logging
		if c != nil {
			cnts = append(cnts, *c)
		}
	}

	entities, err := getEntity(ctx, c, moTypes)
	if err != nil {
		return nil, err
	}

	specs, err := createQuerySpecs(ctx, pm, entities, &cnts)
	if err != nil {
		return nil, err
	}

	entityMetrics, err := pm.Query(ctx, *specs)
	if err != nil {
		return nil, err
	}

	metrics := []Metric{}
	for _, s := range entityMetrics {
		entityMetric, ok := s.(*types.PerfEntityMetric)
		if !ok {
			continue
		}

		entityRef := entityMetric.Entity
		entity := Entity{
			Id:   entityRef.Value,
			Name: findEntityName(entities, entityRef),
			Type: ManagedEntityType(entityRef.Type),
		}

		for _, v := range entityMetric.Value {
			metricSeries, ok := v.(*types.PerfMetricIntSeries)
			if !ok {
				continue
			}

			cnt := findCounter(p, metricSeries.Id.CounterId)
			if cnt == nil {
				continue
			}

			for idx, val := range metricSeries.Value {
				sampling := entityMetric.SampleInfo[idx]

				metric := Metric{
					Entity:    entity,
					Counter:   *cnt,
					Instance:  metricSeries.Id.Instance,
					Timestamp: sampling.Timestamp,
					Value:     val,
				}

				metrics = append(metrics, metric)
			}
		}

	}

	return metrics, nil
}

func createQuerySpecs(ctx context.Context, pm *performance.Manager, entities *[]mo.ManagedEntity, counters *[]CounterInfo) (*[]types.PerfQuerySpec, error) {
	querySpecs := []types.PerfQuerySpec{}
	for _, entity := range *entities {
		createQuerySpec, err := createQuerySpec(ctx, pm, &entity, counters)
		if err != nil {
			//return nil, err
			// TODO: interval
			continue
		}

		querySpecs = append(querySpecs, *createQuerySpec)
	}

	return &querySpecs, nil
}

func createQuerySpec(ctx context.Context, pm *performance.Manager, e *mo.ManagedEntity, counters *[]CounterInfo) (*types.PerfQuerySpec, error) {
	metrics, err := pm.AvailableMetric(ctx, e.Reference(), 0)
	if err != nil {
		return nil, err
	}

	ids := []types.PerfMetricId{}
	for _, m := range metrics {
		if counters != nil {
			for _, c := range *counters {
				if m.CounterId == c.Key {
					ids = append(ids, m)
				}
			}
		} else {
			ids = append(ids, m)
		}
	}

	spec := types.PerfQuerySpec{
		Entity:    e.Reference(),
		MaxSample: 1,
		MetricId:  ids,
	}

	return &spec, nil
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

func findCounter(p mo.PerformanceManager, counterId int32) *CounterInfo {
	for _, c := range p.PerfCounter {
		if c.Key == counterId {
			return ToCounterInfo(&c)
		}
	}

	return nil
}
