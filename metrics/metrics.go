/*
Copyright 2017 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/caicloud/nirvana/service"
)

var (
	once            sync.Once
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	startTime       prometheus.Gauge
)

// Install registers the metrics under the given namespace. This take effect only once; subsequent calls
// has no effect.
func Install(namespace string) {
	once.Do(func() {
		if namespace == "" {
			namespace = "nirvana"
		}
		constLabel := prometheus.Labels{"component": namespace}
		httpLabels := []string{"method", "path"}

		startTime = promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "start_time_seconds",
				Help:        "Start time of the service in unix timestamp",
				ConstLabels: constLabel,
			},
		)
		startTime.Set(float64(time.Now().Unix()))

		requestCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Name:        "request_total",
				Help:        "Counter of server requests for each verb, API resource, and HTTP response code.",
				ConstLabels: constLabel,
			},
			append(httpLabels, "code"),
		)

		requestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Name:        "request_duration_seconds",
				Help:        "Request duration distribution in seconds for each verb and path.",
				ConstLabels: constLabel,
				Buckets:     prometheus.DefBuckets,
			},
			httpLabels,
		)

		responseSize = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Name:        "response_sizes",
				Help:        "Response content length distribution in bytes for each verb and path.",
				ConstLabels: constLabel,
				Buckets:     []float64{1e04, 1e05, 1e06, 1e07, 1e08, 1e09},
			},
			httpLabels,
		)
	})
}

// RecordRequest can be used at the end of each request to record its metric values.
func RecordRequest(start time.Time, ctx service.HTTPContext) {
	req := ctx.Request()
	resp := ctx.ResponseWriter()
	path := ctx.RoutePath()
	requestCounter.WithLabelValues(req.Method, path, strconv.Itoa(resp.StatusCode())).Inc()
	responseSize.WithLabelValues(req.Method, path).Observe(float64(resp.ContentLength()))
	requestDuration.WithLabelValues(req.Method, path).Observe(time.Since(start).Seconds())
}
