package exporter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/9506hqwy/vmomi-exporter/pkg/config"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

type VmomiCollectorOptions struct {
	Context context.Context
}

type vmomiCollector struct {
	Context context.Context
	Config  config.Config
	metrics []PerfGauge
}

func defaultGoCollectorOptions() VmomiCollectorOptions {
	return VmomiCollectorOptions{
		Context: nil,
	}
}

func WithVmomiCollectorContext(ctx context.Context) func(o *VmomiCollectorOptions) {
	return func(o *VmomiCollectorOptions) {
		o.Context = ctx
	}
}

func NewVmomiCollector(opts ...func(o *VmomiCollectorOptions)) prometheus.Collector {
	opt := defaultGoCollectorOptions()
	for _, o := range opts {
		o(&opt)
	}

	infoStartedLog(opt.Context)

	cfg, err := config.GetConfig(opt.Context)
	if err != nil {
		errorCompletedLog(opt.Context, err)
		cfg = config.DefaultConfig()
	}

	metrics, err := GetPerfGauge(opt.Context)
	if err != nil {
		errorCompletedLog(opt.Context, err)
		metrics = []PerfGauge{}
	}

	infoCompletedLog(opt.Context, "metric_count", len(metrics))
	return &vmomiCollector{
		Context: opt.Context,
		Config:  *cfg,
		metrics: metrics,
	}
}

func (c *vmomiCollector) Describe(ch chan<- *prometheus.Desc) {
	slog.InfoContext(c.Context, "Started")

	for _, m := range c.metrics {
		m.Gauge.Describe(ch)
	}

	infoCompletedLog(c.Context)
}

func (c *vmomiCollector) Collect(ch chan<- prometheus.Metric) {
	infoStartedLog(c.Context)

	roots, err := ToEntityFromRoot(c.Context, c.Config.Roots)
	if err != nil {
		errorCompletedLog(c.Context, err)
		return
	}

	moTypes := []string{}
	for _, o := range c.Config.Objects {
		moTypes = append(moTypes, string(*o.Type))
	}

	counters := []vmomi.CounterInfo{}
	for _, o := range c.Config.Counters {
		v := vmomi.CounterInfo{
			Group:  o.Group,
			Name:   o.Name,
			Rollup: o.Rollup,
		}
		counters = append(counters, v)
	}

	metrics, err := vmomi.Query(c.Context, roots, moTypes, counters)
	if err != nil {
		errorCompletedLog(c.Context, err)
		return
	}

	// Reset all metrics to remove all instances.
	resetMetrics(c.metrics)

	// Do not use because expose metrics with timestamp
	// gauge.Gauge.Collect(ch)

	for _, m := range metrics {
		c.sendMetric(ch, m)
	}

	infoCompletedLog(c.Context)
}

func (c *vmomiCollector) sendMetric(ch chan<- prometheus.Metric, m vmomi.Metric) {
	gauge := findPerfGaugeByID(c.metrics, m.Counter.ID)
	if gauge == nil {
		slog.WarnContext(c.Context, "Not found", "counter", m.Counter)
		return
	}

	inst := m.Instance
	if inst == "" {
		inst = m.Entity.Name
	}

	gaugeWithLabels := gauge.Gauge.With(prometheus.Labels{
		LabelCounterInterval: fmt.Sprintf("%v", m.Interval),
		LabelEntityID:        m.Entity.ID,
		LabelEntityName:      m.Entity.Name,
		LabelEntityType:      string(m.Entity.Type),
		LabelEntityInstance:  inst,
	})

	gaugeWithLabels.Set(float64(m.Value))

	ch <- prometheus.NewMetricWithTimestamp(m.Timestamp, gaugeWithLabels)
}

func findPerfGaugeByID(gauges []PerfGauge, id int32) *PerfGauge {
	for _, g := range gauges {
		if g.ID == id {
			return &g
		}
	}

	return nil
}

func resetMetrics(gauges []PerfGauge) {
	for _, g := range gauges {
		g.Gauge.Reset()
	}
}

func infoStartedLog(c context.Context) {
	slog.InfoContext(c, "Started")
}

func infoCompletedLog(c context.Context, args ...any) {
	slog.InfoContext(c, "Completed", args...)
}

func errorCompletedLog(c context.Context, err error) {
	slog.ErrorContext(c, "Completed", "error", err)
}
