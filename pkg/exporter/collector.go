package exporter

import (
	"context"
	"fmt"

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

	config, err := config.GetConfig(opt.Context)
	if err != nil {
		panic(err)
	}

	metrics, _ := GetPerfGauge(opt.Context)

	return &vmomiCollector{
		Context: opt.Context,
		Config:  *config,
		metrics: metrics,
	}
}

func (c *vmomiCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		m.Gauge.Describe(ch)
	}
}

func (c *vmomiCollector) Collect(ch chan<- prometheus.Metric) {
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

	metrics, err := vmomi.Query(c.Context, moTypes, counters)
	if err != nil {
		return
	}

	// Reset all metrics to remove all instances.
	resetMetrics(c.metrics)

	// Do not use because expose metrics with timestamp
	//gauge.Gauge.Collect(ch)

	for _, m := range metrics {
		gauge := findPerfGaugeById(c.metrics, m.Counter.Key)
		if gauge == nil {
			continue
		}

		inst := m.Instance
		if inst == "" {
			inst = m.Entity.Name
		}

		gaugeWithLabels := gauge.Gauge.With(prometheus.Labels{
			LabelCounterInterval: fmt.Sprintf("%v", m.Interval),
			LabelEntityId:        m.Entity.Id,
			LabelEntityName:      m.Entity.Name,
			LabelEntityType:      string(m.Entity.Type),
			LabelEntityInstance:  inst,
		})

		gaugeWithLabels.Set(float64(m.Value))

		ch <- prometheus.NewMetricWithTimestamp(m.Timestamp, gaugeWithLabels)
	}
}

func findPerfGaugeById(gauges []PerfGauge, id int32) *PerfGauge {
	for _, g := range gauges {
		if g.Id == id {
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
