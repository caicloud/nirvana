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
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
)

func TestInstall(t *testing.T) {
	const path = "/api/v1/messages"
	resetAll()
	Install(&Options{
		NamespaceLabel: "Module!",
		NamespaceValue: "In-sight!",
		SubsystemLabel: "COMPONENT",
		SubsystemValue: "Controller_Manager",
	})
	defer func() {
		// we don't usually allow re-registering the metrics, but we must make an exception
		// for the unit tests here
		prometheus.Unregister(requestCount)
		prometheus.Unregister(requestDuration)
		prometheus.Unregister(responseSize)
		once = sync.Once{}
	}()
	RecordRestfulRequest(path, http.MethodGet, http.StatusOK, 50, time.Millisecond)
	RecordRestfulRequest(path, http.MethodPost, http.StatusCreated, 100, time.Millisecond)
	var testCases metricsTestCases = map[string]metricsTestCase{
		"AbnormalLabelKeyAndValue": {
			Target: requestCount,
			Want: `
				# HELP insight_controller_manager_request_total Counter of server requests.
				# TYPE insight_controller_manager_request_total counter
				insight_controller_manager_request_total{action="",code="200",component="controller_manager",module="insight",path="/api/v1/messages",verb="GET",version=""} 1
				insight_controller_manager_request_total{action="",code="201",component="controller_manager",module="insight",path="/api/v1/messages",verb="POST",version=""} 1
`,
		},
	}
	testCases.Test(t)
}

func TestRecordRestfulRequest(t *testing.T) {
	const path = "/api/v1/messages"
	resetAll()
	Install(nil)
	for i, duration := 0, time.Millisecond; i < 16; i, duration = i+1, duration*2 {
		RecordRestfulRequest(path, http.MethodGet, http.StatusOK, 50, duration)
	}
	for i, duration := 0, time.Millisecond; i < 8; i, duration = i+1, duration*2 {
		RecordRestfulRequest(path, http.MethodPost, http.StatusCreated, 100, duration)
	}
	var testCases metricsTestCases = map[string]metricsTestCase{
		"RequestCounter": {
			Target: requestCount,
			Want: `
				# HELP nirvana_request_total Counter of server requests.
				# TYPE nirvana_request_total counter
				nirvana_request_total{action="",code="200",path="/api/v1/messages",verb="GET",version=""} 16
				nirvana_request_total{action="",code="201",path="/api/v1/messages",verb="POST",version=""} 8
`,
		},
		"ResponseSize": {
			Target: responseSize,
			Want: `
				# HELP nirvana_response_sizes Response content length distribution in bytes.
				# TYPE nirvana_response_sizes histogram
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="10000"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="100000"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="1e+06"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="1e+07"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="1e+08"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="1e+09"} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="+Inf"} 16
				nirvana_response_sizes_sum{action="",path="/api/v1/messages",verb="GET",version=""} 800
				nirvana_response_sizes_count{action="",path="/api/v1/messages",verb="GET",version=""} 16
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="10000"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="100000"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="1e+06"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="1e+07"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="1e+08"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="1e+09"} 8
				nirvana_response_sizes_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="+Inf"} 8
				nirvana_response_sizes_sum{action="",path="/api/v1/messages",verb="POST",version=""} 800
				nirvana_response_sizes_count{action="",path="/api/v1/messages",verb="POST",version=""} 8
`,
		},
		"RequestDuration": {
			Target: requestDuration,
			Want: `
				# HELP nirvana_request_duration_seconds Request duration distribution in seconds.
				# TYPE nirvana_request_duration_seconds histogram
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.005"} 3
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.01"} 4
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.025"} 5
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.05"} 6
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.1"} 7
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.25"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="0.5"} 9
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="1"} 10
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="2.5"} 12
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="5"} 13
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="10"} 14
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="GET",version="",le="+Inf"} 16
				nirvana_request_duration_seconds_sum{action="",path="/api/v1/messages",verb="GET",version=""} 65.535
				nirvana_request_duration_seconds_count{action="",path="/api/v1/messages",verb="GET",version=""} 16
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.005"} 3
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.01"} 4
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.025"} 5
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.05"} 6
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.1"} 7
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.25"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="0.5"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="1"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="2.5"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="5"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="10"} 8
				nirvana_request_duration_seconds_bucket{action="",path="/api/v1/messages",verb="POST",version="",le="+Inf"} 8
				nirvana_request_duration_seconds_sum{action="",path="/api/v1/messages",verb="POST",version=""} 0.255
				nirvana_request_duration_seconds_count{action="",path="/api/v1/messages",verb="POST",version=""} 8
`,
		},
	}
	testCases.Test(t)
}

func TestRecordRPCRequest(t *testing.T) {
	const (
		action  = "echo"
		version = "2020-01-01"
	)
	resetAll()
	Install(nil)
	for i, duration := 0, time.Millisecond; i < 16; i, duration = i+1, duration*2 {
		RecordRPCRequest(action, version, http.StatusOK, 50, duration)
	}
	for i, duration := 0, time.Millisecond; i < 8; i, duration = i+1, duration*2 {
		RecordRPCRequest(action, version, http.StatusCreated, 100, duration)
	}
	var testCases metricsTestCases = map[string]metricsTestCase{
		"RequestCounter": {
			Target: requestCount,
			Want: `
				# HELP nirvana_request_total Counter of server requests.
				# TYPE nirvana_request_total counter
				nirvana_request_total{action="echo",code="200",path="",verb="",version="2020-01-01"} 16
				nirvana_request_total{action="echo",code="201",path="",verb="",version="2020-01-01"} 8
`,
		},
		"ResponseSize": {
			Target: responseSize,
			Want: `
				# HELP nirvana_response_sizes Response content length distribution in bytes.
				# TYPE nirvana_response_sizes histogram
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="10000"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="100000"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="1e+06"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="1e+07"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="1e+08"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="1e+09"} 24
				nirvana_response_sizes_bucket{action="echo",path="",verb="",version="2020-01-01",le="+Inf"} 24
				nirvana_response_sizes_sum{action="echo",path="",verb="",version="2020-01-01"} 1600
				nirvana_response_sizes_count{action="echo",path="",verb="",version="2020-01-01"} 24
`,
		},
		"RequestDuration": {
			Target: requestDuration,
			Want: `
				# HELP nirvana_request_duration_seconds Request duration distribution in seconds.
				# TYPE nirvana_request_duration_seconds histogram
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.005"} 6
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.01"} 8
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.025"} 10
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.05"} 12
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.1"} 14
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.25"} 16
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="0.5"} 17
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="1"} 18
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="2.5"} 20
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="5"} 21
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="10"} 22
				nirvana_request_duration_seconds_bucket{action="echo",path="",verb="",version="2020-01-01",le="+Inf"} 24
				nirvana_request_duration_seconds_sum{action="echo",path="",verb="",version="2020-01-01"} 65.78999999999999
				nirvana_request_duration_seconds_count{action="echo",path="",verb="",version="2020-01-01"} 24
`,
		},
	}
	testCases.Test(t)
}

// metricsTestCase can be used to unit test a Prometheus Collector implementation. It takes
// a initialized Prometheus Collector and check if the metric it defines is standard and
// produces the expected output.
type metricsTestCase struct {
	// Target is the Prometheus Collector implementation to test
	Target prometheus.Collector
	// Want is the expected output in the metric API response as the result of Target Collector
	Want string
	// Metrics is the names of the metrics to test; leave empty to test all metrics.
	Metrics []string
}

// Test runs the tests. It returns an error if the Collector does not work properly. It returns
// a list of link errors and no errors if the Collector works but has non-standard definition.
func (mtc metricsTestCase) Test() ([]promlint.Problem, error) {
	if problems, err := testutil.CollectAndLint(mtc.Target, mtc.Metrics...); err != nil {
		return nil, errors.Wrap(err, "collection failed")
	} else if len(problems) > 0 {
		return problems, nil
	}
	want := bytes.NewBufferString(mtc.Want)
	if err := testutil.CollectAndCompare(mtc.Target, want, mtc.Metrics...); err != nil {
		return nil, errors.Wrap(err, "output verification failed")
	}
	return nil, nil
}

// metricsTestCases is a alias for a set of metricsTestCase. It provide a convenient and standard
// way to unit test multiple metricsTestCase. The keys of the map are the names of the test cases.
type metricsTestCases map[string]metricsTestCase

// Test runs all given test cases under the given testing.T object.
func (mtc metricsTestCases) Test(t *testing.T) {
	for name, tc := range mtc {
		t.Run(name, func(tt *testing.T) {
			problems, err := tc.Test()
			if err != nil {
				tt.Error(err)
			}
			for _, problem := range problems {
				tt.Errorf("non-standard metric '%s': %s", problem.Metric, problem.Text)
			}
		})
	}
}

func resetAll() {
	if requestCount != nil {
		requestCount.Reset()
	}
	if requestDuration != nil {
		requestDuration.Reset()
	}
	if responseSize != nil {
		responseSize.Reset()
	}
}
