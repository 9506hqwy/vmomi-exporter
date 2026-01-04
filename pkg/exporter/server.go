package exporter

import (
	"context"
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/9506hqwy/vmomi-exporter/pkg/flag"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

func Run(ctx context.Context) error {
	_, _, _, _, err := vmomi.GetTarget(ctx)
	if err != nil {
		return err
	}

	exporterUrl, ok := ctx.Value(flag.ExporterUrlKey{}).(string)
	if !ok {
		return errors.New("exporter_url not found in context")
	}

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		NewVmomiCollector(WithVmomiCollectorContext(ctx)),
	)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	return http.ListenAndServe(exporterUrl, nil)
}
