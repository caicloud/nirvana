package router

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	paths := []string{
		"/segments/segment/resources/resource",
		"/segments/{segment}/resources/{resource}",
		"/segments/{segment:[a-z]{1,2}}.log{temp}sss/paths/{path:*}",
	}
	results := [][]string{
		{"/segments/segment/resources/resource"},
		{"/segments/", "{segment}", "/resources/", "{resource}"},
		{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp}", "sss/paths/", "{path:*}"},
	}
	for i, p := range paths {
		result, err := Split(p)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		want := results[i]
		if len(result) != len(want) {
			t.Fatalf("Length is not equal for url: %s", p)
		}
		t.Log(result)
		t.Log(want)
		for j := 0; j < len(result); j++ {
			if result[j] != want[j] {
				t.Fatalf("The split result is incorrect: %v,%v", result, want)
			}
		}
	}
}

func TestReorganize(t *testing.T) {
	paths := [][]string{
		{"/segments/segment/resources/resource"},
		{"/segments/", "{segment}", "/resources/", "{resource}"},
		{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp}", "sss/paths/", "{path:*}"},
	}
	results := [][]*Segment{
		{{"/segments/segment/resources/resource", nil, String}},
		{
			{"/segments/", nil, String},
			{"(?P<segment>.*)", []string{"segment"}, Regexp},
			{"/resources/", nil, String},
			{"(?P<resource>.*)", []string{"resource"}, Regexp},
		},
		{
			{"/segments/", nil, String},
			{`(?P<segment>[a-z]{1,2})\.log(?P<temp>.*)sss`, []string{"segment", "temp"}, Regexp},
			{"/paths/", nil, String},
			{"", []string{"path"}, Path},
		},
	}
	for i, p := range paths {
		result, err := Reorganize(p)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		want := results[i]
		if len(result) != len(want) {
			t.Fatalf("Length is not equal for path: %v", p)
		}
		for j := 0; j < len(result); j++ {
			t.Logf("%+v", result[j])
			t.Logf("%+v", want[j])
			if !reflect.DeepEqual(result[j], want[j]) {
				t.Fatalf("The split result is incorrect: %v", p)
			}
		}
	}

}

func TestParse(t *testing.T) {
	path := "/segments/{cmd}/{segment:[a-z]{1,2}}.log{temp}sss/paths/{path:*}"
	root, leaf, err := Parse(path)
	if err != nil {
		t.Fatalf("Untracked error: %s", err.Error())
	}
	router := root

	if r, ok := router.(*StringNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != "/segments/" {
			t.Fatalf("Invalid router target")
		}
		if len(r.RegexpRouters) != 1 {
			t.Fatalf("Invalid children router")
		}
		router = r.RegexpRouters[0]
		t.Log(r.Target())
	}

	if r, ok := router.(*FullMatchRegexpNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != (&ExpSegment{FullMatchTarget, r.Key}).Target() {
			t.Fatalf("Invalid router target")
		}
		if len(r.StringRouters) != 1 {
			t.Fatalf("Invalid children router")
		}
		router = r.StringRouters[0].Router
		t.Log(r.Target())
	}

	if r, ok := router.(*StringNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != "/" {
			t.Fatalf("Invalid router target")
		}
		if len(r.RegexpRouters) != 1 {
			t.Fatalf("Invalid children router")
		}
		router = r.RegexpRouters[0]
		t.Log(r.Target())
	}

	if r, ok := router.(*RegexpNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != `(?P<segment>[a-z]{1,2})\.log(?P<temp>.*)sss` {
			t.Fatalf("Invalid router target")
		}
		if len(r.StringRouters) != 1 {
			t.Fatalf("Invalid children router")
		}
		router = r.StringRouters[0].Router
		t.Log(r.Target())
	}

	if r, ok := router.(*StringNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != "/paths/" {
			t.Fatalf("Invalid router target")
		}
		if r.PathRouter == nil {
			t.Fatalf("Invalid children router")
		}
		router = r.PathRouter
		t.Log(r.Target())
	}

	if r, ok := router.(*PathNode); !ok {
		t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
	} else {
		if r.Target() != "" {
			t.Fatalf("Invalid router target")
		}
		if r != leaf {
			t.Fatalf("Invalid router: %s", reflect.TypeOf(router).String())
		}
		t.Log(r.Target())
	}
}

func TestRouter(t *testing.T) {
	rds := []TestRouterData{
		{"/api/v1/namespaces", &TestExecutor{"GET", 1}},
		{"/api/v1/namespaces/{namespace}", &TestExecutor{"POST", 2}},
		{"/api/v1/namespaces/{namespace2:one[3456]?}/subjects", &TestExecutor{"CONNECT", 200}},
		{"/api/v1/namespaces/{namespace}/objects", &TestExecutor{"PUT", 3}},
		{"/api/v1/namespaces/{namespace}/objects/{object:*}", &TestExecutor{"DELETE", 4}},
		{"/api/v1/namespaces", &TestExecutor{"POST", 100}},
		{"/api/v1/namespaces/{namespace}/objects/{object:*}", &TestExecutor{"PUT", 400}},
		{"/api/v1/namespaces/{namespace2:one[3456]?}/subjects", &TestExecutor{"GET", 201}},
	}
	right := []TestData{
		{"/api/v1/namespaces", "GET", 1, map[string]string{}},
		{"/api/v1/namespaces", "POST", 100, map[string]string{}},
		{"/api/v1/namespaces/one", "POST", 2, map[string]string{"namespace": "one"}},
		{"/api/v1/namespaces/one/subjects", "CONNECT", 200, map[string]string{"namespace2": "one"}},
		{"/api/v1/namespaces/one3/subjects", "CONNECT", 200, map[string]string{"namespace2": "one3"}},
		{"/api/v1/namespaces/one5/subjects", "CONNECT", 200, map[string]string{"namespace2": "one5"}},
		{"/api/v1/namespaces/one4/subjects", "GET", 201, map[string]string{"namespace2": "one4"}},
		{"/api/v1/namespaces/one/objects", "PUT", 3, map[string]string{"namespace": "one"}},
		{"/api/v1/namespaces/one/objects/two", "DELETE", 4, map[string]string{"namespace": "one", "object": "two"}},
		{"/api/v1/namespaces/one/objects/two2", "PUT", 400, map[string]string{"namespace": "one", "object": "two2"}},
	}
	wrong := []TestData{
		{"/api/v1/namespaces/", "GET", 0, nil},
		{"/api/v1/namespaces/one3/subjects", "POST", 0, nil},
		{"/api/v1/namespaces/one1/subjects", "CONNECT", 0, nil},
		{"/api/v1/namespaces/one12/subjects", "CONNECT", 0, nil},
		{"/api/v1/namespaces/", "GET", 0, nil},
		{"/api/v1/namespaces", "PUT", 0, nil},
	}

	var root Router
	for _, d := range rds {
		router, leaf, err := Parse(d.Path)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		leaf.AddExecutor(d.Executor)
		if root == nil {
			root = router
		} else {
			root.Merge(router)
		}
	}
	rj, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("Can't Marshal router: %s", err.Error())
	}
	t.Logf("%s", string(rj))

	for _, d := range right {
		values := NewTestValueContainer()
		ctx := context.WithValue(context.Background(), "Type", d.Type)
		e := root.Match(ctx, values, d.Path)
		if e == nil {
			t.Fatalf("Can't match path: %s", d.Path)
		}
		for k, v := range d.Values {
			pv, ok := values.Get(k)
			if !ok || pv != v {
				t.Fatalf("Can't match path with values: %s(%+v, %+v)", d.Path, d.Values, values)
			}
		}
		result := 0
		ctx = context.WithValue(ctx, "Result", &result)
		err := e.Execute(ctx)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		if result != d.Result {
			t.Fatalf("Executor returns an invalid value: %d, Expect: %d", result, d.Result)
		}
	}

	for _, d := range wrong {
		values := NewTestValueContainer()
		ctx := context.WithValue(context.Background(), "Type", d.Type)
		e := root.Match(ctx, values, d.Path)
		if e != nil {
			t.Logf("%+v", e)
			t.Fatalf("Matched by mistake: %s", d.Path)
		}
	}
}

type TestData struct {
	Path   string
	Type   string
	Result int
	Values map[string]string
}

type TestRouterData struct {
	Path     string
	Executor *TestExecutor
}

type TestValueContainer struct {
	Data map[string]string
}

func NewTestValueContainer() *TestValueContainer {
	return &TestValueContainer{
		make(map[string]string),
	}
}
func (tvc *TestValueContainer) Set(key, value string) {
	tvc.Data[key] = value
}

func (tvc *TestValueContainer) Get(key string) (string, bool) {
	v, ok := tvc.Data[key]
	return v, ok
}

type TestExecutor struct {
	Type   string
	Result int
}

func (te *TestExecutor) Inspect(c context.Context) (Executor, bool) {
	ins := c.Value("Type")
	if typ, ok := ins.(string); ok && typ == te.Type {
		return te, true
	}
	return nil, false
}

func (te *TestExecutor) Execute(c context.Context) error {
	result := c.Value("Result")
	if pointer, ok := result.(*int); ok {
		*pointer = te.Result
	} else {
		panic("can't find result from context")
	}
	return nil
}
