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
