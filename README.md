# VMOMI Exporter

This repository provides Prometheus exporter for vSphere infrastructure.
It collects and exposes vSphere performance metrics for monitoring.

## Features

- Collects vSphere performance counters
- Flexible configuration for target entities and metrics
- Exposes metrics at `/metrics` for Prometheus scraping
- Includes exporter process and Go runtime metrics

### Labels

Expose metrics with follow labels.

| Label            | Description                     |
| :--------------- | :------------------------------ |
| counter_id       | Internal identifier for counter |
| counter_stat     | Kind of statistics for counter  |
| counter_unit     | Unit of value for counter       |
| counter_interval | Interval for counter            |
| entity_id        | Internal identifier for entiry  |
| entity_name      | Display name for entity         |
| entity_type      | Kind for entity                 |
| entity_instance  | Instance of entity for counter  |

## TODO

- More configuration options
- Multiple concurrent access support
- Timeout support
- Metrics acquirement splitting.
- Session caching
- Other than performance metrics

## Build

Build binary.

```sh
go build -o bin/vmomi-exporter ./cmd/vmomi-exporter
```

Or build container image.

```sh
docker build -t vmomi-exporter .
```

Add `Z` option at bind mount operation in *Dockerfile* if using podman with SELinux.

## Usage

Run application.

```sh
$ ./bin/vmomi-exporter -h
VMOMI Exporter

Usage:
  vmomi-exporter [flags]
  vmomi-exporter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      VMOMI Exporter Config
  counter     VMOMI Exporter Counter
  help        Help about any command
  instance    VMOMI Exporter Instance
  interval    VMOMI Exporter Interval
  perf        VMOMI Exporter Performance

Flags:
      --config string      Config file path.
      --exporter string    Exporter URL. (default "127.0.0.1:9247")
  -h, --help               help for vmomi-exporter
      --log-level string   Log level. (default "INFO")
      --no-verify-ssl      Skip SSL verification.
      --password string    vSphere server password.
      --url string         vSphere server URL. (default "https://127.0.0.1/sdk")
      --user string        vSphere server username.
  -v, --version            version for vmomi-exporter

Use "vmomi-exporter [command] --help" for more information about a command.
```

Set environment variable instead of arguments.

| Argument        | Environment Variable                |
| :-------------- | :---------------------------------- |
| --config        | VMOMI_EXPORTER_CONFIG               |
| --exporter      | VMOMI_EXPORTER_URL                  |
| --log-level     | VMOMI_EXPORTER_LOG_LEVEL            |
| --no-verify-ssl | VMOMI_EXPORTER_TARGET_NO_VERIFY_SSL |
| --password      | VMOMI_EXPORTER_TARGET_PASSWORD      |
| --url           | VMOMI_EXPORTER_TARGET_URL           |
| --user          | VMOMI_EXPORTER_TARGET_USER          |

Or run container.

```sh
docker run -d \
    -e VMOMI_EXPORTER_TARGET_URL=<URL> \
    -e VMOMI_EXPORTER_TARGET_USER=<USER> \
    -e VMOMI_EXPORTER_TARGET_PASSWORD=<PASSWORD> \
    -p 9247:9247 \
    vmomi-exporter
```

### Subcommands

- `config`: Show current configuration
- `counter`: List available performance counters
- `instance`: List available performance instances
- `interval`: List available performance counter interval
- `perf`: Show performace value

## Configuration

Configure the exporter using the `--config` option. See [examples/all.yaml](./examples/all.yaml) for a full example.

### Default Configuration

```yaml
counters:
 - group: cpu
   name: usage
   rollup: average
 - group: cpu
   name: usagemhz
   rollup: average
 - group: mem
   name: usage
   rollup: average

objects:
 - type: HostSystem
 - type: VirtualMachine
```

### Definition

`counters` defines the counter described on [PerformanceManager][PerformanceManager].
`objects` defined the type of [ManagedEntity][ManagedEntity] described on [ManagedObjectReference][ManagedObjectReference].

| key             | valye                                                       |
| :-------------- | :---------------------------------------------------------- |
| counters        | List counters.                                              |
| counters.group  | `groupInfo` in [PerfCounterInfo][PerfCounterInfo].          |
| counters.name   | `nameInfo` in [PerfCounterInfo][PerfCounterInfo].           |
| counters.rollup | `rollupType` in [PerfCounterInfo][PerfCounterInfo].         |
| objects         | List target objects.                                        |
| objects.type    | `type` in [ManagedObjectReference][ManagedObjectReference]. |

[PerformanceManager]: https://developer.broadcom.com/xapis/vsphere-web-services-api/latest/vim.PerformanceManager.html
[PerfCounterInfo]: https://developer.broadcom.com/xapis/vsphere-web-services-api/latest/vim.PerformanceManager.CounterInfo.html
[ManagedObjectReference]: https://developer.broadcom.com/xapis/vsphere-web-services-api/latest/vmodl.ManagedObjectReference.html
[ManagedEntity]: https://developer.broadcom.com/xapis/vsphere-web-services-api/latest/vim.ManagedEntity.html

`vmomi-exporter counter` command acquires all counters from target environment.

## NOTES

- In large environment, occur error.
  - Update `config.vpxd.stats.maxQueryMetrics` in vCenter Server.
  - see [Performance charts are empty and displays the error: Request processing is restricted by administrator](https://knowledge.broadcom.com/external/article/301449/performance-charts-are-empty-and-display.html)
