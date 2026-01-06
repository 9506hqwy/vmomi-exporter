package exporter

import (
	"context"
	"fmt"
	"strings"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LabelCounterId       = "counter_id"
	LabelCounterStat     = "counter_stat"
	LabelCounterUnit     = "counter_unit"
	LabelCounterInterval = "counter_interval"
	LabelEntityId        = "entity_id"
	LabelEntityName      = "entity_name"
	LabelEntityType      = "entity_type"
	LabelEntityInstance  = "entity_instance"
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
					LabelCounterId:   fmt.Sprintf("%v", i.Id),
					LabelCounterStat: i.Stats,
					LabelCounterUnit: i.Unit,
				},
			}, []string{LabelCounterInterval, LabelEntityId, LabelEntityName, LabelEntityType, LabelEntityInstance})
			gauge := PerfGauge{
				Id:    i.Id,
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
