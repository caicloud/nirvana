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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once            sync.Once
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
)

// Options provide a way to configure the name of the metrics (by setting Namespace and Subsystem) and
// control if and how these value should be used as label values.
// If you are unsure about these options, just use nil or empty Options.
type Options struct {
	NamespaceLabel string `desc:"label name for the metrics namespace"`
	NamespaceValue string `desc:"metrics namespace; also used as the value of the namespace label (if provided)"`
	SubsystemLabel string `desc:"label name for the metrics subsystem"`
	SubsystemValue string `desc:"metrics subsystem; also used as the value of the subsystem value (if provided)"`
}

const defaultNamespace = "nirvana"

// DefaultOptions builds an Options object with default values; it is used automatically when a nil
// Options is provided.
func DefaultOptions() *Options {
	return &Options{
		NamespaceValue: defaultNamespace,
	}
}

// Install registers the metrics under the given namespace and subsystem. This take effect only
// once; subsequent calls has no effect.
func Install(options *Options) {
	once.Do(func() {
		if options == nil {
			options = DefaultOptions()
		}
		if len(options.NamespaceValue) == 0 {
			options.NamespaceValue = defaultNamespace
		}
		namespace := normalizeLabelName(options.NamespaceValue)
		subsystem := normalizeLabelName(options.SubsystemValue)
		constLabel := prometheus.Labels{}
		if labelKey := normalizeLabelName(options.NamespaceLabel); len(labelKey) > 0 {
			constLabel[labelKey] = namespace
		}
		if labelKey := normalizeLabelName(options.SubsystemLabel); len(labelKey) > 0 && len(subsystem) > 0 {
			constLabel[labelKey] = subsystem
		}
		httpLabels := []string{"verb", "path", "action", "version"}

		requestCount = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "request_total",
				Help:        "Counter of server requests.",
				ConstLabels: constLabel,
			},
			append(httpLabels, "code"),
		)

		requestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "request_duration_seconds",
				Help:        "Request duration distribution in seconds.",
				ConstLabels: constLabel,
				Buckets:     prometheus.DefBuckets,
			},
			httpLabels,
		)

		responseSize = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "response_sizes",
				Help:        "Response content length distribution in bytes.",
				ConstLabels: constLabel,
				Buckets:     []float64{1e04, 1e05, 1e06, 1e07, 1e08, 1e09},
			},
			httpLabels,
		)
	})
}

// RecordRestfulRequest gathers the metric values of a Restful HTTP request. It should be called at
// the end of a request session.
func RecordRestfulRequest(path, verb string, code int, contentLength int, duration time.Duration) {
	verb = strings.ToUpper(verb)
	labels := prometheus.Labels{
		"verb":    verb,
		"path":    path,
		"action":  "",
		"version": "",
	}
	labelsWithCode := prometheus.Labels{
		"verb":    verb,
		"path":    path,
		"action":  "",
		"version": "",
		"code":    strconv.Itoa(code),
	}
	requestCount.With(labelsWithCode).Inc()
	responseSize.With(labels).Observe(float64(contentLength))
	requestDuration.With(labels).Observe(duration.Seconds())
}

// RecordRPCRequest gathers the metric values of a RPC HTTP request. It should be called at
// the end of a request session.
func RecordRPCRequest(action, version string, code int, contentLength int, duration time.Duration) {
	labels := prometheus.Labels{
		"verb":    "",
		"path":    "",
		"action":  action,
		"version": version,
	}
	labelsWithCode := prometheus.Labels{
		"verb":    "",
		"path":    "",
		"action":  action,
		"version": version,
		"code":    strconv.Itoa(code),
	}
	requestCount.With(labelsWithCode).Inc()
	responseSize.With(labels).Observe(float64(contentLength))
	requestDuration.With(labels).Observe(duration.Seconds())
}

var labelRegex = regexp.MustCompile("[^a-z0-9_]+")

// normalizeLabelName convert the given string into a valid label name (or any part of one)
func normalizeLabelName(label string) string {
	if len(label) == 0 {
		return ""
	}
	lower := strings.ToLower(label)
	noSpace := strings.ReplaceAll(lower, " ", "_")
	alphaNumeric := labelRegex.ReplaceAllString(noSpace, "")
	trimmed := strings.Trim(alphaNumeric, "_")
	return trimmed
}
