package vmomi

import (
	"context"
	"errors"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

type Metric struct {
	Entity    Entity
	Counter   CounterInfo
	Instance  string
	Timestamp time.Time
	Value     int64
	Interval  int32
}

type InstanceInfo struct {
	EntityType ManagedEntityType
	EntityId   string
	EntityName string
	Instance   string
	CounterId  int32
}

func GetInstanceInfo(ctx context.Context, types []ManagedEntityType) (*[]InstanceInfo, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	moTypes := []string{}
	for _, t := range types {
		moTypes = append(moTypes, string(t))
	}

	entities, err := getEntity(ctx, c, moTypes)
	if err != nil {
		return nil, err
	}

	pm := performance.NewManager(c)
	specs, err := createQuerySpecs(ctx, pm, p.HistoricalInterval, entities, nil)
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

func GetIntervalInfo(ctx context.Context, entity Entity) ([]int32, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	mor := types.ManagedObjectReference{
		Type:  string(entity.Type),
		Value: entity.Id,
	}

	pm := performance.NewManager(c)
	intervals, err := getIntervalIds(ctx, pm, p.HistoricalInterval, mor)
	if err != nil {
		return nil, err
	}

	return intervals, nil

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
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	cnts := []CounterInfo{}
	for _, c := range counters {
		c := ComplementCounterInfo(*p, c)
		// TODO: logging
		if c != nil {
			cnts = append(cnts, *c)
		}
	}

	entities, err := getEntity(ctx, c, moTypes)
	if err != nil {
		return nil, err
	}

	pm := performance.NewManager(c)
	specs, err := createQuerySpecs(ctx, pm, p.HistoricalInterval, entities, &cnts)
	if err != nil {
		return nil, err
	}

	entityMetrics, err := pm.Query(ctx, *specs)
	if err != nil {
		return nil, err
	}

	metrics := []Metric{}
	for _, s := range entityMetrics {
		m, err := ToMetric(p, entities, s)
		if err != nil {
			continue
		}

		metrics = append(metrics, m...)
	}

	return metrics, nil
}

func QueryEntity(ctx context.Context, entity Entity, interval int32, counterId int32) ([]Metric, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	entities, err := getEntity(ctx, c, []string{string(entity.Type)})
	if err != nil {
		return nil, err
	}

	var mo *mo.ManagedEntity
	for _, e := range *entities {
		if e.Reference().Value == entity.Id {
			mo = &e
			break
		}
	}

	if mo == nil {
		return nil, errors.New("entity not found")
	}

	pm := performance.NewManager(c)
	specs, err := createQuerySpec(ctx, pm, mo, interval, &[]CounterInfo{{Id: counterId}})
	if err != nil {
		return nil, err
	}

	entityMetrics, err := pm.Query(ctx, []types.PerfQuerySpec{*specs})
	if err != nil {
		return nil, err
	}

	metrics := []Metric{}
	for _, s := range entityMetrics {
		m, err := ToMetric(p, entities, s)
		if err != nil {
			continue
		}

		metrics = append(metrics, m...)
	}

	return metrics, nil

}

func ToMetric(p *mo.PerformanceManager, entities *[]mo.ManagedEntity, s types.BasePerfEntityMetricBase) ([]Metric, error) {
	entityMetric, ok := s.(*types.PerfEntityMetric)
	if !ok {
		return nil, errors.New("invalid metric type")
	}

	entityRef := entityMetric.Entity
	entity := Entity{
		Id:   entityRef.Value,
		Name: findEntityName(entities, entityRef),
		Type: ManagedEntityType(entityRef.Type),
	}

	metrics := []Metric{}
	for _, v := range entityMetric.Value {
		metricSeries, ok := v.(*types.PerfMetricIntSeries)
		if !ok {
			continue
		}

		cnt := findCounter(*p, metricSeries.Id.CounterId)
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
				Interval:  sampling.Interval,
			}

			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func createQuerySpecs(ctx context.Context, pm *performance.Manager, intervalIds []types.PerfInterval, entities *[]mo.ManagedEntity, counters *[]CounterInfo) (*[]types.PerfQuerySpec, error) {
	querySpecs := []types.PerfQuerySpec{}
	intervalIdCache := map[string]int32{}
	for _, entity := range *entities {
		moType := entity.Reference().Type

		if _, ok := intervalIdCache[moType]; !ok {
			intervalId, err := getIntervalId(ctx, pm, intervalIds, entity.Reference())
			if err != nil {
				continue
			}

			intervalIdCache[moType] = *intervalId
		}

		createQuerySpec, err := createQuerySpec(ctx, pm, &entity, intervalIdCache[moType], counters)
		if err != nil {
			return nil, err
		}

		querySpecs = append(querySpecs, *createQuerySpec)
	}

	return &querySpecs, nil
}

func createQuerySpec(ctx context.Context, pm *performance.Manager, e *mo.ManagedEntity, intervalId int32, counters *[]CounterInfo) (*types.PerfQuerySpec, error) {
	metrics, err := pm.AvailableMetric(ctx, e.Reference(), intervalId)
	if err != nil {
		return nil, err
	}

	ids := []types.PerfMetricId{}
	for _, m := range metrics {
		if counters != nil {
			for _, c := range *counters {
				if m.CounterId == c.Id {
					ids = append(ids, m)
				}
			}
		} else {
			ids = append(ids, m)
		}
	}

	spec := types.PerfQuerySpec{
		Entity:     e.Reference(),
		MaxSample:  1,
		MetricId:   ids,
		IntervalId: intervalId,
	}

	return &spec, nil
}

func getIntervalId(ctx context.Context, pm *performance.Manager, intervalIds []types.PerfInterval, e types.ManagedObjectReference) (*int32, error) {
	intervals, err := getIntervalIds(ctx, pm, intervalIds, e)
	if err != nil {
		return nil, err
	}

	intervalId := int32(0)
	for _, interval := range intervals {
		if intervalId == 0 || interval < intervalId {
			intervalId = interval
		}
	}

	if intervalId == 0 {
		return nil, errors.New("no supported interval found")
	}

	return &intervalId, nil
}

func getIntervalIds(ctx context.Context, pm *performance.Manager, intervalIds []types.PerfInterval, e types.ManagedObjectReference) ([]int32, error) {
	summary, err := pm.ProviderSummary(ctx, e)
	if err != nil {
		return nil, err
	}

	intervals := []int32{}

	if summary.CurrentSupported {
		intervals = append(intervals, summary.RefreshRate)
	}

	if summary.SummarySupported {
		for _, interval := range intervalIds {
			intervals = append(intervals, interval.SamplingPeriod)
		}
	}

	return intervals, nil
}

func getPerformanceManager(ctx context.Context, c *vim25.Client) (*mo.PerformanceManager, error) {
	pc := property.DefaultCollector(c)

	var p mo.PerformanceManager
	err := pc.RetrieveOne(ctx, *c.ServiceContent.PerfManager, nil, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func findCounter(p mo.PerformanceManager, counterId int32) *CounterInfo {
	for _, c := range p.PerfCounter {
		if c.Key == counterId {
			return ToCounterInfo(&c)
		}
	}

	return nil
}
