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
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

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

type IntervalID struct {
	ID      int32
	Current bool
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

	serverClock, err := methods.GetCurrentTime(ctx, c)
	if err != nil {
		return nil, err
	}

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
	specs, err := createQuerySpecs(ctx, serverClock, pm, p.HistoricalInterval, entities, nil)
	if err != nil {
		return nil, err
	}

	info := ToInstanceInfoList(entities, specs)
	return &info, nil
}

func GetIntervalInfo(ctx context.Context, entity Entity) ([]IntervalID, error) {
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

	serverClock, err := methods.GetCurrentTime(ctx, c)
	if err != nil {
		return nil, err
	}

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
	specs, err := createQuerySpecs(ctx, serverClock, pm, p.HistoricalInterval, entities, &cnts)
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

	serverClock, err := methods.GetCurrentTime(ctx, c)
	if err != nil {
		return nil, err
	}

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

	specs, err := createQuerySpecEntity(
		ctx,
		serverClock,
		pm,
		found,
		interval,
		counterID,
	)
	if err != nil {
		return nil, err
	}

	entityMetrics, err := pm.Query(ctx, *specs)
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
	serverClock *time.Time,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	entities *[]mo.ManagedEntity,
	counters *[]CounterInfo,
) (*[]types.PerfQuerySpec, error) {
	querySpecs := []types.PerfQuerySpec{}
	intervalIDCache := map[string]IntervalID{}
	for _, entity := range *entities {
		intervalID := findIntervalID(ctx, pm, intervalIDs, intervalIDCache, entity)
		if intervalID == nil {
			// Not support current and historical.
			continue
		}

		createQuerySpec, err := createQuerySpec(
			ctx,
			serverClock,
			pm,
			&entity,
			*intervalID,
			counters,
		)
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

func createQuerySpecEntity(
	ctx context.Context,
	serverClock *time.Time,
	pm *performance.Manager,
	entity *mo.ManagedEntity,
	interval int32,
	counterID int32,
) (*[]types.PerfQuerySpec, error) {
	intervalID := IntervalID{
		ID:      interval,
		Current: true,
	}

	specs, err := createQuerySpec(
		ctx,
		serverClock,
		pm,
		entity,
		intervalID,
		&[]CounterInfo{{ID: counterID}},
	)
	if err != nil {
		return nil, err
	} else if specs == nil {
		return nil, errors.New("not found counter")
	}

	querySpecs := []types.PerfQuerySpec{*specs}
	return &querySpecs, nil
}

func createQuerySpec(
	ctx context.Context,
	serverClock *time.Time,
	pm *performance.Manager,
	e *mo.ManagedEntity,
	intervalID IntervalID,
	counters *[]CounterInfo,
) (*types.PerfQuerySpec, error) {
	metrics, err := pm.AvailableMetric(ctx, e.Reference(), intervalID.ID)
	if err != nil {
		return nil, err
	}

	ids := filterPerfMetricID(counters, metrics)

	if len(ids) == empty {
		return nil, nil
	}

	var startTime *time.Time
	if !intervalID.Current {
		// Limit sampling using period.
		// Use 30min because datastore min period is 30min.
		t := serverClock.Add(-30 * time.Minute)
		startTime = &t
	}

	spec := types.PerfQuerySpec{
		Entity:     e.Reference(),
		MaxSample:  sampling,
		MetricId:   ids,
		IntervalId: intervalID.ID,
		StartTime:  startTime,
	}

	return &spec, nil
}

func getIntervalID(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	e types.ManagedObjectReference,
) (*IntervalID, error) {
	intervals, err := getIntervalIDs(ctx, pm, intervalIDs, e)
	if err != nil {
		return nil, err
	}

	var intervalID *IntervalID
	for _, interval := range intervals {
		if intervalID == nil || interval.ID < intervalID.ID {
			intervalID = &interval
		}
	}

	return intervalID, nil
}

func getIntervalIDs(
	ctx context.Context,
	pm *performance.Manager,
	intervalIDs []types.PerfInterval,
	e types.ManagedObjectReference,
) ([]IntervalID, error) {
	summary, err := pm.ProviderSummary(ctx, e)
	if err != nil {
		return nil, err
	}

	intervals := []IntervalID{}

	if summary.CurrentSupported {
		i := IntervalID{
			ID:      summary.RefreshRate,
			Current: true,
		}
		intervals = append(intervals, i)
	}

	if summary.SummarySupported {
		for _, interval := range intervalIDs {
			i := IntervalID{
				ID:      interval.SamplingPeriod,
				Current: false,
			}
			intervals = append(intervals, i)
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
	intervalIDCache map[string]IntervalID,
	entity mo.ManagedEntity,
) *IntervalID {
	moType := entity.Reference().Type

	if _, ok := intervalIDCache[moType]; !ok {
		intervalID, err := getIntervalID(ctx, pm, intervalIDs, entity.Reference())
		if err != nil {
			slog.WarnContext(ctx, "Could not get interval", "error", err, "type", moType)
		}

		if intervalID == nil {
			intervalID = &IntervalID{
				ID:      empty32,
				Current: true,
			}
		}

		intervalIDCache[moType] = *intervalID
	}

	if intervalIDCache[moType].ID == empty32 {
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
