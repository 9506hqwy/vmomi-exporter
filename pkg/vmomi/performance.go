package vmomi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

const initIntervalKey = int32(0)
const empty = int(0)
const empty32 = int32(0)
const first = int(0)
const sampling = int32(0)

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
	EntityID   string
	EntityName string
	Instance   string
	CounterID  int32
}

func GetInstanceInfo(
	ctx context.Context,
	entityTypes []ManagedEntityType,
) (*[]InstanceInfo, error) {
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
	for _, t := range entityTypes {
		moTypes = append(moTypes, string(t))
	}

	roots := []types.ManagedObjectReference{c.ServiceContent.RootFolder}
	entities, err := getEntities(ctx, c, roots, moTypes, false)
	if err != nil {
		return nil, err
	}

	pm := performance.NewManager(c)
	specs, err := createQuerySpecs(ctx, pm, p.HistoricalInterval, entities, nil)
	if err != nil {
		return nil, err
	}

	info := ToInstanceInfoList(entities, specs)
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
		Value: entity.ID,
	}

	pm := performance.NewManager(c)
	intervals, err := getIntervalIDs(ctx, pm, p.HistoricalInterval, mor)
	if err != nil {
		return nil, err
	}

	return intervals, nil
}

func ToInstanceInfoList(
	entities *[]mo.ManagedEntity,
	specs *[]types.PerfQuerySpec,
) []InstanceInfo {
	info := []InstanceInfo{}
	for _, spec := range *specs {
		entityName := findEntityName(entities, spec.Entity)
		for _, metricID := range spec.MetricId {
			info = append(info, *ToInstanceInfo(&metricID, spec.Entity, entityName))
		}
	}

	return info
}

func ToInstanceInfo(
	c *types.PerfMetricId,
	mor types.ManagedObjectReference,
	name string,
) *InstanceInfo {
	return &InstanceInfo{
		EntityType: ManagedEntityType(mor.Type),
		EntityID:   mor.Value,
		EntityName: name,
		Instance:   c.Instance,
		CounterID:  c.CounterId,
	}
}

func Query(
	ctx context.Context,
	rootEntities *[]Entity,
	moTypes []string,
	counters []CounterInfo,
) ([]Metric, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	cnts := ComplementCounterInfoList(ctx, *p, counters)

	roots := toRootManagedObjectReference(c, rootEntities)

	entities, err := getEntities(ctx, c, roots, moTypes, rootEntities != nil)
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

	return ToMetrics(ctx, p, entities, &entityMetrics)
}

func QueryEntity(
	ctx context.Context,
	entity Entity,
	interval int32,
	counterID int32,
) ([]Metric, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	p, err := getPerformanceManager(ctx, c)
	if err != nil {
		return nil, err
	}

	roots := []types.ManagedObjectReference{c.ServiceContent.RootFolder}
	entities, err := getEntities(ctx, c, roots, []string{string(entity.Type)}, false)
	if err != nil {
		return nil, err
	}

	found, err := findRootEntity(entities, entity)
	if err != nil {
		return nil, err
	}

	pm := performance.NewManager(c)
	specs, err := createQuerySpec(ctx, pm, found, interval, &[]CounterInfo{{ID: counterID}})
	if err != nil {
		return nil, err
	}

	if specs == nil {
		return nil, errors.New("not found counter")
	}

	entityMetrics, err := pm.Query(ctx, []types.PerfQuerySpec{*specs})
	if err != nil {
		return nil, err
	}

	return ToMetrics(ctx, p, entities, &entityMetrics)
}

func ToMetrics(
	ctx context.Context,
	p *mo.PerformanceManager,
	entities *[]mo.ManagedEntity,
	entityMetrics *[]types.BasePerfEntityMetricBase,
) ([]Metric, error) {
	metrics := []Metric{}
	for _, s := range *entityMetrics {
		m, err := ToMetric(p, entities, s)
		if err != nil {
			slog.WarnContext(ctx, "Could not convert", "metric", s)
			continue
		}

		metrics = append(metrics, m...)
	}

	return metrics, nil
}

func ToMetric(
	p *mo.PerformanceManager,
	entities *[]mo.ManagedEntity,
	s types.BasePerfEntityMetricBase,
) ([]Metric, error) {
	entityMetric, ok := s.(*types.PerfEntityMetric)
	if !ok {
		return nil, errors.New("invalid metric type")
	}

	entityRef := entityMetric.Entity
	entity := Entity{
		ID:   entityRef.Value,
		Name: findEntityName(entities, entityRef),
		Type: ManagedEntityType(entityRef.Type),
	}

	metrics := []Metric{}
	for _, v := range entityMetric.Value {
		metric, err := toMetricFromManaged(p, entity, entityMetric, v)
		if err != nil {
			return nil, err
		} else if metric != nil {
			metrics = append(metrics, *metric)
		}
	}

	return metrics, nil
}

func createQuerySpecs(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	entities *[]mo.ManagedEntity,
	counters *[]CounterInfo,
) (*[]types.PerfQuerySpec, error) {
	querySpecs := []types.PerfQuerySpec{}
	intervalIDCache := map[string]int32{}
	for _, entity := range *entities {
		intervalID := findIntervalID(ctx, pm, intervalIDs, intervalIDCache, entity)
		if intervalID == nil {
			// Not support current and historical.
			continue
		}

		createQuerySpec, err := createQuerySpec(ctx, pm, &entity, *intervalID, counters)
		if err != nil {
			return nil, err
		}

		if createQuerySpec == nil {
			continue
		}

		querySpecs = append(querySpecs, *createQuerySpec)
	}

	slog.DebugContext(ctx, "Completed", "intervals", intervalIDCache)
	return &querySpecs, nil
}

func createQuerySpec(
	ctx context.Context,
	pm *performance.Manager,
	e *mo.ManagedEntity,
	intervalID int32,
	counters *[]CounterInfo,
) (*types.PerfQuerySpec, error) {
	metrics, err := pm.AvailableMetric(ctx, e.Reference(), intervalID)
	if err != nil {
		return nil, err
	}

	ids := filterPerfMetricID(counters, metrics)

	if len(ids) == empty {
		return nil, nil
	}

	spec := types.PerfQuerySpec{
		Entity:     e.Reference(),
		MaxSample:  sampling,
		MetricId:   ids,
		IntervalId: intervalID,
	}

	return &spec, nil
}

func getIntervalID(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	e types.ManagedObjectReference,
) (*int32, error) {
	intervals, err := getIntervalIDs(ctx, pm, intervalIDs, e)
	if err != nil {
		return nil, err
	}

	intervalID := initIntervalKey
	for _, interval := range intervals {
		if intervalID == initIntervalKey || interval < intervalID {
			intervalID = interval
		}
	}

	return &intervalID, nil
}

func getIntervalIDs(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	e types.ManagedObjectReference,
) ([]int32, error) {
	summary, err := pm.ProviderSummary(ctx, e)
	if err != nil {
		return nil, err
	}

	intervals := []int32{}

	if summary.CurrentSupported {
		intervals = append(intervals, summary.RefreshRate)
	}

	if summary.SummarySupported {
		for _, interval := range intervalIDs {
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

func findCounter(p mo.PerformanceManager, counterID int32) *CounterInfo {
	for _, c := range p.PerfCounter {
		if c.Key == counterID {
			return ToCounterInfo(&c)
		}
	}

	return nil
}

func findIntervalID(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	intervalIDCache map[string]int32,
	entity mo.ManagedEntity,
) *int32 {
	moType := entity.Reference().Type

	if _, ok := intervalIDCache[moType]; !ok {
		intervalID, err := getIntervalID(ctx, pm, intervalIDs, entity.Reference())
		if err != nil {
			slog.WarnContext(ctx, "Could not get interval", "error", err)
			return nil
		}

		intervalIDCache[moType] = *intervalID
	}

	if intervalIDCache[moType] == empty32 {
		// Not support current and historical.
		return nil
	}

	id := intervalIDCache[moType]
	return &id
}

func findRootEntity(entities *[]mo.ManagedEntity, root Entity) (*mo.ManagedEntity, error) {
	var found *mo.ManagedEntity
	for _, e := range *entities {
		if e.Reference().Value == root.ID {
			found = &e
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("entity not found %v(%v)", root.Type, root.ID)
	}

	return found, nil
}

func toRootManagedObjectReference(
	c *vim25.Client,
	rootEntities *[]Entity,
) []types.ManagedObjectReference {
	roots := []types.ManagedObjectReference{}
	if rootEntities == nil {
		roots = append(roots, c.ServiceContent.RootFolder)
	} else {
		for _, e := range *rootEntities {
			mor := types.ManagedObjectReference{
				Type:  string(e.Type),
				Value: e.ID,
			}
			roots = append(roots, mor)
		}
	}

	return roots
}

func toMetricFromManaged(
	p *mo.PerformanceManager,
	entity Entity,
	entityMetric *types.PerfEntityMetric,
	v types.BasePerfMetricSeries,
) (*Metric, error) {
	metricSeries, ok := v.(*types.PerfMetricIntSeries)
	if !ok {
		return nil, errors.New("invalid metric series type")
	}

	if len(entityMetric.SampleInfo) == empty {
		return nil, nil
	}

	// Find the latest value.
	// Because MaxSample is ignored for historical statistics.
	sampling, value := latestSampling(entityMetric, metricSeries)

	cnt := findCounter(*p, metricSeries.Id.CounterId)
	if cnt == nil {
		return nil, fmt.Errorf("not found counter %v", metricSeries.Id.CounterId)
	}

	metric := Metric{
		Entity:    entity,
		Counter:   *cnt,
		Instance:  metricSeries.Id.Instance,
		Timestamp: sampling.Timestamp,
		Value:     value,
		Interval:  sampling.Interval,
	}

	return &metric, nil
}

func filterPerfMetricID(
	counters *[]CounterInfo,
	metrics performance.MetricList,
) []types.PerfMetricId {
	if counters == nil {
		return metrics
	}

	ids := []types.PerfMetricId{}
	for _, m := range metrics {
		for _, c := range *counters {
			if m.CounterId == c.ID {
				ids = append(ids, m)
			}
		}
	}

	return ids
}

func latestSampling(
	entityMetric *types.PerfEntityMetric,
	metricSeries *types.PerfMetricIntSeries,
) (types.PerfSampleInfo, int64) {
	sampling := entityMetric.SampleInfo[first]
	value := metricSeries.Value[first]
	for idx, s := range entityMetric.SampleInfo {
		if s.Timestamp.After(sampling.Timestamp) {
			sampling = s
			value = metricSeries.Value[idx]
		}
	}

	return sampling, value
}
