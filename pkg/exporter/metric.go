package exporter

import (
	"context"
	"fmt"
	"strings"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LabelCounterKey     = "counter_key"
	LabelCounterStat    = "counter_stat"
	LabelCounterUnit    = "counter_unit"
	LabelEntityId       = "entity_id"
	LabelEntityName     = "entity_name"
	LabelEntityType     = "entity_type"
	LabelEntityInstance = "entity_instance"
)

type PerfGauge struct {
	Id    int32
	Gauge prometheus.GaugeVec
}

func GetPerfGauge(ctx context.Context) ([]PerfGauge, error) {
	metrics := []PerfGauge{}

	info, err := vmomi.GetCounterInfo(ctx)
	if err == nil {
		for _, i := range *info {
			metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: ToPerfGaugeId(&i),
				Help: i.NameSummary,
				ConstLabels: prometheus.Labels{
					LabelCounterKey:  fmt.Sprintf("%v", i.Key),
					LabelCounterStat: i.Stats,
					LabelCounterUnit: i.Unit,
				},
			}, []string{LabelEntityId, LabelEntityName, LabelEntityType, LabelEntityInstance})
			gauge := PerfGauge{
				Id:    i.Key,
				Gauge: *metric,
			}

			metrics = append(metrics, gauge)
		}
	}

	return metrics, nil
}

func ToPerfGaugeId(c *vmomi.CounterInfo) string {
	name := fmt.Sprintf("%v_%v_%v", c.Group, c.Name, c.Rollup)
	return strings.ReplaceAll(name, ".", "_")
}
