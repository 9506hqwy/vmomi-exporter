package vmomi

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	sx "github.com/9506hqwy/vmomi-exporter/pkg/vmomi/sessionex"
)

type CounterInfo struct {
	Id          int32
	Group       string
	Name        string
	NameSummary string
	Rollup      string
	Stats       string
	Unit        string
}

func GetCounterInfo(ctx context.Context) (*[]CounterInfo, error) {
	c, err := login(ctx)
	if err != nil {
		return nil, err
	}

	defer sx.Logout(ctx, c)

	pc := property.DefaultCollector(c)

	var p mo.PerformanceManager
	err = pc.RetrieveOne(ctx, *c.ServiceContent.PerfManager, nil, &p)
	if err != nil {
		return nil, err
	}

	info := []CounterInfo{}
	for _, cnt := range p.PerfCounter {
		info = append(info, *ToCounterInfo(&cnt))
	}

	return &info, nil
}

func ToCounterInfo(c *types.PerfCounterInfo) *CounterInfo {
	return &CounterInfo{
		Id:          c.Key,
		Group:       c.GroupInfo.GetElementDescription().Key,
		Name:        c.NameInfo.GetElementDescription().Key,
		NameSummary: c.NameInfo.GetElementDescription().Summary,
		Rollup:      fmt.Sprintf("%v", c.RollupType),
		Stats:       fmt.Sprintf("%v", c.StatsType),
		Unit:        c.UnitInfo.GetElementDescription().Key,
	}
}

func ComplementCounterInfo(p mo.PerformanceManager, cnt CounterInfo) *CounterInfo {
	for _, c := range p.PerfCounter {
		if c.Key != 0 && c.Key == cnt.Id {
			return ToCounterInfo(&c)
		}

		group := c.GroupInfo.GetElementDescription().Key
		name := c.NameInfo.GetElementDescription().Key
		Rollup := string(c.RollupType)
		if group == cnt.Group && name == cnt.Name && Rollup == cnt.Rollup {
			return ToCounterInfo(&c)
		}

	}

	return nil
}
