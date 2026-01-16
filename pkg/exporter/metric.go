package exporter

import (
	"context"
	"fmt"
	"strings"

	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	LabelCounterID       = "counter_id"
	LabelCounterStat     = "counter_stat"
	LabelCounterUnit     = "counter_unit"
	LabelCounterInterval = "counter_interval"
	LabelEntityID        = "entity_id"
	LabelEntityName      = "entity_name"
	LabelEntityType      = "entity_type"
	LabelEntityInstance  = "entity_instance"
)

type PerfGauge struct {
	ID    int32
	Gauge prometheus.GaugeVec
}

func GetPerfGauge(ctx context.Context) ([]PerfGauge, error) {
	info, err := vmomi.GetCounterInfo(ctx)
	if err != nil {
		return nil, err
	}

	metrics := []PerfGauge{}

	for _, i := range *info {
		metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: ToPerfGaugeID(&i),
			Help: i.NameSummary,
			ConstLabels: prometheus.Labels{
				LabelCounterID:   fmt.Sprintf("%v", i.ID),
				LabelCounterStat: i.Stats,
				LabelCounterUnit: i.Unit,
			},
		}, []string{
			LabelCounterInterval,
			LabelEntityID,
			LabelEntityName,
			LabelEntityType,
			LabelEntityInstance,
		})
		gauge := PerfGauge{
			ID:    i.ID,
			Gauge: *metric,
		}

		metrics = append(metrics, gauge)
	}

	return metrics, nil
}

func ToPerfGaugeID(c *vmomi.CounterInfo) string {
	name := fmt.Sprintf("%v_%v_%v", c.Group, c.Name, c.Rollup)
	return strings.ReplaceAll(name, ".", "_")
}
