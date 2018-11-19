/*
Copyright 2018 Caicloud Authors

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

package v2

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/caicloud/nirvana/examples/api-basic/application"

	"github.com/caicloud/nirvana/utils/unittest"
)

func TestCreate(t *testing.T) {
	s, err := unittest.NewTestService(Descriptor())
	if err != nil {
		t.Fatal(err)
	}

	appData := []byte(`{"metadata": {
		"name": "foo",
		"partition": "default"
	},
	"spec": {
		"replica": 1
	}}`)

	req, err := unittest.NewJSONRequest(context.Background(), "POST", "/api/v2/applications", appData)
	if err != nil {
		t.Fatal(err)
	}

	rw := unittest.NewResponseWriter()
	s.ServeHTTP(rw, req)
	if rw.Code() != 201 {
		t.Errorf("expect code 201, got %v", rw.Code())
	}

	t.Logf("%s", rw.Bytes())
	var ret application.Application
	if err := json.Unmarshal(rw.Bytes(), &ret); err != nil {
		t.Fatal(err)
	}
	if ret.Spec.Replica != 1 {
		t.Errorf("expect replica 1, got %v", ret.Spec.Replica)
	}
}
