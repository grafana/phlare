---
title: "Go (pull mode)"
menuTitle: "Go (pull mode)"
description: "Instrumenting Golang applications for continuous profiling"
weight: 30
---

# How to add Golang profiling to your application

Modern observability systems generally fall into two categories: those where services push data, and those where the
observability system pulls data from services. We find that both approaches are suitable: they both have advantages
and disadvantages, and one of them can be more desirable than another under certain circumstances.

Pyroscope server can operate in both "pull" and "push" modes. The current implementation makes extensive use of Prometheus
scrape and service discovery mechanisms.


## Supported languages and platforms

Any application that exposes data in [`pprof`](https://github.com/google/pprof/blob/master/doc/README.md) format via HTTP can be set as a remote profiling target.


## Scrape configuration

Pyroscope uses exactly the same mechanisms as Prometheus does in order to ensure smooth user experience. Therefore, it
is configured in almost identical way, and [Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config)
can be used as a reference.

If using Pyroscope you will need to add the following content to your `pyroscope/server.yml` Pyroscope config file. See the [Server config documentation](/docs/server-configuration#configuration-file) for more information on where this config is located by default on your system.

If using Phlare ... TODO

```yaml
---
# A list of scrape configurations.
scrape-configs:
  # The job name assigned to scraped profiles by default.
  - job-name: pyroscope

    # How frequently to scrape targets by default.
    scrape-interval: 10s

    # The list of profiles to be scraped from the targets.
    enabled-profiles: [cpu, mem, goroutines, mutex, block]

    # List of labeled statically configured targets for this job.
    static-configs:
      - application: my-application-name
        spy-name: gospy
        targets:
          - hostname:6060
        labels:
          env: dev
```

## Scrape interval

By default, a target is scraped on 10 seconds interval which is specified with `scrape-interval` parameter.
That is, with the default configuration a 10 seconds profile is collected every 10 seconds. Both interval and duration
parameters can be modified to allow sampling strategies, where only a subset of targets is being profiled at a given time
and only a subset of profiling data is being collected.

For example, configuring `scrape-interval` to 60 seconds with the default 10 second profiling duration effectively
means that only 1/6 of the data is collected, because only 1/6 of the targets is being scraped, which in turn:
 - Reduces the overall profiling overhead
 - Decreases resource usage of the Pyroscope server

```yaml
---
scrape-configs:
  - job-name: pyroscope
    scrape-interval: 60s
    enabled-profiles: [cpu, mem, goroutines, mutex, block]
    static-configs:
      - application: my-application-name
        spy-name: gospy
        targets:
          - hostname:6060
        labels:
          env: dev
```

There are two types of profiles in Go:
 - Profiles that accumulate samples during a profiling session: `cpu`, `mutex`, and `block`.
 - Instant profiles that represent the current state: `goroutines` and `mem` – Pyroscope stores the delta of two consecutive "snapshots".

For profiles of the first type, you can override the duration of the profiling session, which allows you to use even more
flexible scenarios. However, setting the profiling duration shorter than the scrape interval degrades the accuracy of the
resulting profiles and may significantly complicate their analysis.

For profiles of the second type, profiling duration is always equal to `scrape-interval`.

The example configuration below instructs Pyroscope to collected cpu, block, and mutex profiles for 30 seconds
every minute, goroutine and memory profiles are collected every minute as well but include all the stack trace samples
emerged within this time window.


```yaml
---
scrape-configs:
  - job-name: pyroscope
    scrape-interval: 60s
    scrape-timeout: 60s
    enabled-profiles: [cpu, mem, goroutines, mutex, block]
    use-delta-profiles: true
    profiles:
      cpu:
        params:
          seconds: [ "30" ]
      block:
        params:
          seconds: [ "30" ]
      mutex:
        params:
          seconds: [ "30" ]
    static-configs:
      - application: my-application-name
        spy-name: gospy
        targets:
          - hostname:6060
        labels:
          env: dev
```

## Service Discovery

Pyroscope creates pull targets based on the discovered labels. At least `__name__` and `__address__` labels must be
present, where `__name__` is the name of the application being profiled.

Optional labels:
* `__scheme__`: If the metrics endpoint is secured then you will need to set this to `https`.
* `__port__`: Scrape the target on the indicated port.
* `__profile_{profile_name}_enabled`: Indicates whether a particular profile should be scraped.
* `__profile_{profile_name}_path`: Specifies URL path exposing pprof profile.
* `__profile_{profile_name}_param_{param_key}`: Overrides scrape URL parameters.

Where `{profile_name}` must be a valid profile configuration name.

At this point, Pyroscope fully supports only [Kubernetes Service Discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config).


## Mutex Profiling

Mutex profiling is useful for finding sources of contention within your application. It helps you to find out which mutexes are being held by which goroutines.

To enable mutex profiling, you need to add the following code to your application:
```go
runtime.SetMutexProfileFraction(rate)
```

`rate` parameter controls the fraction of mutex contention events that are reported in the mutex profile. On average 1/rate events are reported.

## Block Profiling

Block profiling lets you analyze how much time your program spends waiting on the blocking operations such as:
* select
* channel send/receive
* semacquire
* notifyListWait

To enable block profiling, you need to add the following code to your application:
```go
runtime.SetBlockProfileRate(rate)
```

`rate` parameter controls the fraction of goroutine blocking events that are reported in the blocking profile. The profiler aims to sample an average of one blocking event per rate nanoseconds spent blocked.

## godeltaprof

[godeltaprof](https://github.com/pyroscope-io/godeltaprof) is a memory profiler for cumulative profiles(heap, block, mutex).
It is more efficient because it does the delta/merging before producing pprof data, avoiding extra decompression/parsing/allocations/compression.

To start using godeltaprof in pull mode in a Go application, you need to include godeltaprof module in your app:
```bash
go get github.com/pyroscope-io/godeltaprof@latest
```
Integration is very simillar to `net/http/pprof`, you need to import a new package and it will expose new endpoints `/debug/pprof/delta_heap`, `/debug/pprof/delta_block`, `/debug/pprof/delta_mutex`
```go
_ "github.com/pyroscope-io/godeltaprof/http/pprof"
```
In the scrape config you need to enable new delta endpoints with `use-delta-endpoints: true`, for example:
```yaml
scrape-configs:
  - job-name: pyroscope1
    enabled-profiles: [cpu, mem, block, mutex]
    use-delta-profiles: true
```

## Examples

### Static targets

You can find an example of how Pyroscope scrapes static targets in
[the pyroscope repository](https://github.com/github/pyroscope/blob/main/examples/golang-pull/static).

### Kubernetes service discovery

You can see how Pyroscope discovers remote targets in Kubernetes using
[the example setup](https://github.com/github/pyroscope/tree/main/examples/golang-pull/kubernetes) in the pyroscope repository.

Here is an example of how to configure Pyroscope to scrape targets from Kubernetes pods:
```yaml
---
pyroscopeConfigs:
  log-level: debug
  scrape-configs:
  - job-name: 'kubernetes-pods'
    enabled-profiles: [ cpu, mem ]
    kubernetes-sd-configs:
      - role: pod
    relabel-configs:
      - source-labels: [__meta_kubernetes_pod_annotation_pyroscope_io_scrape]
        action: keep
        regex: true
      - source-labels:
          [__meta_kubernetes_pod_annotation_pyroscope_io_application_name]
        action: replace
        target-label: __name__
      - source-labels: [__meta_kubernetes_pod_annotation_pyroscope_io_scheme]
        action: replace
        regex: (https?)
        target-label: __scheme__
      - source-labels:
          [__address__, __meta_kubernetes_pod_annotation_pyroscope_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target-label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source-labels: [__meta_kubernetes_namespace]
        action: replace
        target-label: kubernetes_namespace
      - source-labels: [__meta_kubernetes_pod_name]
        action: replace
        target-label: kubernetes_pod_name
      - source-labels: [__meta_kubernetes_pod_phase]
        regex: Pending|Succeeded|Failed|Completed
        action: drop
      - action: labelmap
        regex: __meta_kubernetes_pod_annotation_pyroscope_io_profile_(.+)
        replacement: __profile_$1
```

Kubernetes Service Discovery requires RBAC set up. Please refer to
[Pyroscope helm chart](https://github.com/pyroscope-io/helm-chart) for details.

### File service discovery

You can configure Pyroscope server to scrape targets defined in files that can be updated dynamically without
restarting the server. You can find an example configuration in [the pyroscope repository](https://github.com/github/pyroscope/blob/main/examples/golang-pull/file).

### Consul service discovery

Consul service discovery allow retrieving scrape targets from Consul's Catalog API.
You can find an example configuration in [the pyroscope repository](https://github.com/github/pyroscope/blob/main/examples/golang-pull/consul).
