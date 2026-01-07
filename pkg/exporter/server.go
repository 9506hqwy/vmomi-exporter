package exporter

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/9506hqwy/vmomi-exporter/pkg/flag"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

func Run(ctx context.Context) error {
	log.SetFlags(log.Flags() | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	logLevel := getLogLevel(ctx)
	slog.SetLogLoggerLevel(logLevel)

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

	slog.Info("HTTP server started", "url", exporterUrl)
	return http.ListenAndServe(exporterUrl, nil)
}

func getLogLevel(ctx context.Context) slog.Level {
	logLevelStr, ok := ctx.Value(flag.LogLevelKey{}).(string)
	if ok {
		switch logLevelStr {
		case "DEBUG":
			return slog.LevelDebug
		case "WARN":
			return slog.LevelWarn
		case "ERROR":
			return slog.LevelError
		}
	}

	return slog.LevelInfo
}
