package performanceex

import (
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func GetCounter(p *mo.PerformanceManager, id int32) *types.PerfCounterInfo {
	for _, cnt := range p.PerfCounter {
		if cnt.Key == id {
			return &cnt
		}
	}

	return nil
}

func GetCounterTypeLabel(p *mo.PerformanceManager, t types.PerfSummaryType) string {
	for _, cnt := range p.Description.CounterType {
		key := types.PerfSummaryType(cnt.GetElementDescription().Key)
		if key == t {
			return cnt.GetElementDescription().Label
		}
	}

	return ""
}

func GetStatTypeLabel(p *mo.PerformanceManager, t types.PerfStatsType) string {
	for _, stat := range p.Description.StatsType {
		key := types.PerfStatsType(stat.GetElementDescription().Key)
		if key == t {
			return stat.GetElementDescription().Label
		}
	}

	return ""
}
