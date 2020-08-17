# Contextualized Stats Tracker for Go

This library provides context-driven stats tracker.

[![Build Status](https://github.com/bool64/stats/workflows/test/badge.svg)](https://github.com/bool64/stats/actions?query=branch%3Amaster+workflow%3Atest)
[![Coverage Status](https://codecov.io/gh/bool64/stats/branch/master/graph/badge.svg)](https://codecov.io/gh/bool64/stats)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/bool64/stats)
![Code lines](https://sloc.xyz/github/bool64/stats/?category=code)
![Comments](https://sloc.xyz/github/bool64/stats/?category=comments)

## Features

* Loosely coupled with underlying implementation.
* Context-driven labels control.
* Zero allocation implementation for Prometheus client.
* Simple interface with variadic number of key-value pairs for labels.
* Easily mockable interface free from 3rd party dependencies.

## Example

```go
// Bring your own Prometheus registry.
registry := prometheus.NewRegistry()
tr := prom.Tracker{
    Registry: registry,
}

// Add custom Prometheus configuration where necessary.
tr.DeclareHistogram("my_latency_seconds", prometheus.HistogramOpts{
    Buckets: []float64{1e-4, 1e-3, 1e-2, 1e-1, 1, 10, 100},
})

ctx := context.Background()

// Add labels to context.
ctx = stats.AddKeysAndValues(ctx, "ctx-label", "ctx-value0")

// Override label values.
ctx = stats.AddKeysAndValues(ctx, "ctx-label", "ctx-value1")

// Collect stats with last mile labels.
tr.Add(ctx, "my_count", 1,
    "some-label", "some-value",
)

tr.Add(ctx, "my_latency_seconds", 1.23)

tr.Set(ctx, "temperature", 33.3)
```

## Performance

Sample benchmark result with Dell XPS 7590 i9-9980HK on Ubuntu 19 and go1.14.2.

```
name              time/op
Tracker_Add-16    814ns ± 6%
RawPrometheus-16  801ns ± 1%

name              alloc/op
Tracker_Add-16    0.00B     
RawPrometheus-16   336B ± 0%

name              allocs/op
Tracker_Add-16     0.00     
RawPrometheus-16   2.00 ± 0%
```

## Caveats

Prometheus client does not support metrics with same name and different label sets. 
If you add a label to context, make sure you have it in all cases, at least with empty value `""`.
