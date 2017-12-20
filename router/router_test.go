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

package router

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var (
	getExecs = []*TestExecutor{{"GET", 200}}
	delExecs = []*TestExecutor{{"DELETE", 201}}
	putExecs = []*TestExecutor{{"PUT", 202}}
)

func errorCompare(t *testing.T, got, wanted error) {
	if wanted != nil {
		if got == nil {
			t.Fatalf("No expected error: %s", wanted.Error())
		}
		if got.Error() != wanted.Error() {
			t.Fatalf("Untracked error: %s want %s", got.Error(), wanted.Error())
		}
	} else {
		if got != nil {
			t.Fatalf("Unexpected error: %s", got.Error())
		}
	}
}

func TestReorganize(t *testing.T) {
	var tab = []struct {
		path   []string
		result []*Segment
		err    error
	}{
		{
			[]string{"/segments/segment/resources/resource"},
			[]*Segment{{"/segments/segment/resources/resource", nil, String}},
			nil,
		},
		{
			[]string{"/segments/", "{segment}", "/resources/", "{resource}"},
			[]*Segment{
				{"/segments/", nil, String},
				{"(?P<segment>.*)", []string{"segment"}, Regexp},
				{"/resources/", nil, String},
				{"(?P<resource>.*)", []string{"resource"}, Regexp},
			},
			nil,
		},
		{
			[]string{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp}", "sss/paths/", "{path:*}"},
			[]*Segment{
				{"/segments/", nil, String},
				{`(?P<segment>[a-z]{1,2})\.log(?P<temp>.*)sss`, []string{"segment", "temp"}, Regexp},
				{"/paths/", nil, String},
				{"", []string{"path"}, Path},
			},
			nil,
		},
		{
			[]string{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp", "sss/paths/", "{path:*}"},
			nil,
			fmt.Errorf("exp does not have normative format: %s", "{temp"),
		},
		{
			[]string{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp}", "{path:*}", "sss/paths/"},
			nil,
			fmt.Errorf("key %s should be last element in the path", "path"),
		},
		{
			[]string{"/segments/", "{segment:[a-z]{1,2}}", ".log", "{temp}", "{path:*}"},
			[]*Segment{
				{"/segments/", nil, String},
				{`(?P<segment>[a-z]{1,2})\.log(?P<temp>.*)`, []string{"segment", "temp"}, Regexp},
				{"", []string{"path"}, Path},
			},
			nil,
		},
	}

	for _, p := range tab {
		result, err := Reorganize(p.path)
		errorCompare(t, err, p.err)
		if len(result) != len(p.result) {
			t.Fatalf("Length is not equal for path: %v", p)
		}
		for j := 0; j < len(result); j++ {
			t.Logf("%+v", result[j])
			t.Logf("%+v", p.result[j])
			if !reflect.DeepEqual(result[j], p.result[j]) {
				t.Fatalf("The split result is incorrect: %v %v", result[j], p.result[j])
			}
		}
	}
}

func TestParse(t *testing.T) {
	type tab struct {
		routerZeroValue   interface{}
		target            string
		lenRegexpChildren int
		lenStringChildren int
		hasPathChildren   bool
		isLeaf            bool
	}
	var router Router
	var caseTabs = []struct {
		path string
		tab  []tab
		err  error
	}{
		{
			path: "/segments/{cmd}/{segment:[a-z]{1,2}}.log{temp}sss/paths/{path:*}",
			tab: []tab{
				{&StringNode{}, "/segments/", 1, 0, false, false},
				{&FullMatchRegexpNode{}, (&ExpSegment{FullMatchTarget, "cmd"}).Target(), 0, 1, false, false},
				{&StringNode{}, "/", 1, 0, false, false},
				{&RegexpNode{}, `(?P<segment>[a-z]{1,2})\.log(?P<temp>.*)sss`, 0, 1, false, false},
				{&StringNode{}, "/paths/", 0, 0, true, false},
				{&PathNode{}, "", 0, 0, false, true},
			},
			err: nil,
		},
		{
			path: "/segments/{{cmd}/{segment:[a-z]{1,2}}.log{temp}sss/paths/{path:*}",
			tab:  nil,
			err:  errors.New("unmatched braces"),
		},
		{
			path: "",
			tab:  nil,
			err:  errors.New("invalid path"),
		},
		{
			path: "/segments/{cmd}/{segment:[a-z]{1,2}}.log{temp}sss/paths/{path:*}/why",
			tab:  nil,
			err:  fmt.Errorf("key %s should be last element in the path", "path"),
		},
		{
			path: "/segments/{cmd}/{segment:[a-z]{1,2}}.log{temp{why}}sss/paths/{path:*}",
			tab:  nil,
			err:  errors.New("error parsing regexp: invalid named capture: `(?P<temp{why}>`"),
		},
	}

	for _, ct := range caseTabs {
		root, leaf, err := Parse(ct.path)
		errorCompare(t, err, ct.err)
		router = root
		for _, tab := range ct.tab {
			if router.Target() != tab.target {
				t.Fatal("Invalid router target")
			}
			t.Log(router.Target())
			switch r := router.(type) {
			case *PathNode:
				if reflect.TypeOf(r) != reflect.TypeOf(tab.routerZeroValue) {
					t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
				}
				if tab.isLeaf {
					if r != leaf {
						t.Fatalf("Invalid router: %s", reflect.TypeOf(router).String())
					}
				}
			case *StringNode:
				if reflect.TypeOf(r) != reflect.TypeOf(tab.routerZeroValue) {
					t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
				}
				if len(r.RegexpRouters) != tab.lenRegexpChildren || len(r.StringRouters) != tab.lenStringChildren {
					t.Fatal("Invalid children router")
				}
				if len(r.RegexpRouters) != 0 {
					router = r.RegexpRouters[0]
				}
				if len(r.StringRouters) != 0 {
					router = r.StringRouters[0].Router
				}
				if tab.hasPathChildren {
					if r.PathRouter == nil {
						t.Fatal("Invalid children router")
					}
					router = r.PathRouter
				}
				if tab.isLeaf {
					if r != leaf {
						t.Fatalf("Invalid router: %s", reflect.TypeOf(router).String())
					}
				}
			case *FullMatchRegexpNode:
				if reflect.TypeOf(r) != reflect.TypeOf(tab.routerZeroValue) {
					t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
				}
				if len(r.RegexpRouters) != tab.lenRegexpChildren || len(r.StringRouters) != tab.lenStringChildren {
					t.Fatal("Invalid children router")
				}
				if len(r.RegexpRouters) != 0 {
					router = r.RegexpRouters[0]
				}
				if len(r.StringRouters) != 0 {
					router = r.StringRouters[0].Router
				}
				if tab.hasPathChildren {
					if r.PathRouter == nil {
						t.Fatal("Invalid children router")
					}
					router = r.PathRouter
				}
				if tab.isLeaf {
					if r != leaf {
						t.Fatalf("Invalid router: %s", reflect.TypeOf(router).String())
					}
				}
			case *RegexpNode:
				if reflect.TypeOf(r) != reflect.TypeOf(tab.routerZeroValue) {
					t.Fatalf("Invalid node type: %s", reflect.TypeOf(router).String())
				}
				if len(r.RegexpRouters) != tab.lenRegexpChildren || len(r.StringRouters) != tab.lenStringChildren {
					t.Fatal("Invalid children router")
				}
				if len(r.RegexpRouters) != 0 {
					router = r.RegexpRouters[0]
				}
				if len(r.StringRouters) != 0 {
					router = r.StringRouters[0].Router
				}
				if tab.hasPathChildren {
					if r.PathRouter == nil {
						t.Fatal("Invalid children router")
					}
					router = r.PathRouter
				}
				if tab.isLeaf {
					if r != leaf {
						t.Fatalf("Invalid router: %s", reflect.TypeOf(router).String())
					}
				}
			}
		}
	}
}

func TestPathNodeMerge(t *testing.T) {
	rds := []TestRouterData{
		{"/api/{object:*}", []*TestExecutor{{"DELETE", 4}}, nil},
		{"/api/{abc:*}", []*TestExecutor{{"PUT", 400}}, nil},
	}

	defer func() {
		if x := recover(); x != nil {
			if fmt.Sprint(x) != "failed to merge path router : unmatched path key: object abc" {
				t.Fatal(x)
			}
		} else {
			t.Fatal("should panic")
		}
	}()

	var root Router
	for _, d := range rds {
		router, leaf, err := Parse(d.Path)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		for _, exec := range d.Executor {
			leaf.AddExecutor(exec)
		}
		if root == nil {
			root = router
		} else {
			root.Merge(router)
		}
	}
}

func TestCommonPrefixMergeError(t *testing.T) {
	rds := []TestRouterData{
		{"api", []*TestExecutor{{"DELETE", 4}}, nil},
		{"/api", []*TestExecutor{{"PUT", 400}}, nil},
	}

	var root Router
	for _, d := range rds {
		router, leaf, err := Parse(d.Path)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		for _, exec := range d.Executor {
			leaf.AddExecutor(exec)
		}
		if root == nil {
			root = router
		} else {
			_, err = root.Merge(router)
			errorCompare(t, err, errors.New("there is no common prefix for the two routers"))
		}
	}
}

func TestRegexpMergeError(t *testing.T) {
	rds := []TestRouterData{
		{"{api1}/v1", []*TestExecutor{{"DELETE", 4}}, nil},
		{"{api}/v1", []*TestExecutor{{"PUT", 400}}, nil},
	}

	var root Router
	for _, d := range rds {
		router, leaf, err := Parse(d.Path)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		for _, exec := range d.Executor {
			leaf.AddExecutor(exec)
		}
		if root == nil {
			root = router
		} else {
			_, err = root.Merge(router)
			errorCompare(t, err, errors.New("unmatched full match key: api1 api"))
		}
	}
}

func makeRouter(t *testing.T, rds []TestRouterData) Router {
	var root Router
	for _, d := range rds {
		router, leaf, err := Parse(d.Path)
		if err != nil {
			t.Fatalf("Untracked error: %s %s", d.Path, err.Error())
		}
		for _, exec := range d.Executor {
			leaf.AddExecutor(exec)
		}
		leaf.AddMiddleware(d.Middleware...)
		if root == nil {
			root = router
		} else {
			_, err = root.Merge(router)
			errorCompare(t, err, nil)
		}
	}
	return root
}

type contextKey string

func (c contextKey) String() string {
	return "mypackage context key " + string(c)
}

func testMatch(t *testing.T, router Router, right []TestData, wrong []TestData) {
	for _, d := range right {
		values := NewTestValueContainer()
		ctx := context.WithValue(context.Background(), contextKey("Type"), d.Type)
		e := router.Match(ctx, values, d.Path)
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
		ctx = context.WithValue(ctx, contextKey("result"), &result)
		err := e.Execute(ctx)
		if err != nil {
			t.Fatalf("Untracked error: %s", err.Error())
		}
		if result != d.Result {
			t.Fatalf("Executor returns an invalid value: %d, Expect: %d path %s", result, d.Result, d.Path)
		}
	}
	for _, d := range wrong {
		values := NewTestValueContainer()
		ctx := context.WithValue(context.Background(), contextKey("Type"), d.Type)
		e := router.Match(ctx, values, d.Path)
		if e != nil {
			t.Logf("%+v", e)
			t.Fatalf("Matched by mistake: %s", d.Path)
		}
	}
}

func TestMiddleWare(t *testing.T) {
	rds := []TestRouterData{
		{"/api/v2", []*TestExecutor{{"GET", 202}, {"DELETE", 203}}, []Middleware{func(ctx context.Context, c RoutingChain) error {
			resultptr := ctx.Value(contextKey("Result")).(*int)
			*resultptr++
			c.Continue(context.WithValue(ctx, contextKey("Result"), resultptr))
			return nil
		}}},
	}
	right := []TestData{
		{"/api/v2", "GET", 203, map[string]string{}},
		{"/api/v2", "DELETE", 204, map[string]string{}},
	}
	testMatch(t, makeRouter(t, rds), right, nil)
}

func TestRouter(t *testing.T) {
	rds := []TestRouterData{
		{"/api/v1/namespaces", []*TestExecutor{{"GET", 1}, {"DELETE", 1}}, nil},
		{"/api/v1/namespaces/{namespace}", []*TestExecutor{{"POST", 2}}, nil},
		{"/api/v1/namespaces/{namespace2:one[3456]?}/subjects", []*TestExecutor{{"CONNECT", 200}}, nil},
		{"/api/v1/namespaces/{namespace}/objects", []*TestExecutor{{"PUT", 3}}, nil},
		{"/api/v1/namespaces/{namespace}/objects/{object:*}", []*TestExecutor{{"DELETE", 4}}, nil},
		{"/api/v1/namespaces", []*TestExecutor{{"POST", 100}, {"HEAD", 100}}, nil},
		{"/api/v1/namespaces/{namespace}/objects/{object:*}", []*TestExecutor{{"PUT", 400}}, nil},
		{"/api/v1/namespaces/{namespace2:one[3456]?}/subjects", []*TestExecutor{{"GET", 201}}, nil},
		{"/api/v2/namespaces/{namespace2:one[3456]?}/subjects", []*TestExecutor{{"GET", 201}}, nil},
		{"/api/v3", []*TestExecutor{{"GET", 203}}, nil},
		{"/api/v4", []*TestExecutor{{"GET", 204}}, nil},
		{"/api/v6", []*TestExecutor{{"GET", 204}}, nil},
		{"/hello", []*TestExecutor{{"GET", 204}}, nil},
		{"/404", []*TestExecutor{{"GET", 404}}, nil},
	}
	right := []TestData{
		{"/api/v1/namespaces", "GET", 1, map[string]string{}},
		{"/api/v1/namespaces", "POST", 100, map[string]string{}},
		{"/api/v1/namespaces", "HEAD", 100, map[string]string{}},
		{"/api/v1/namespaces/one", "POST", 2, map[string]string{"namespace": "one"}},
		{"/api/v1/namespaces/one/subjects", "CONNECT", 200, map[string]string{"namespace2": "one"}},
		{"/api/v1/namespaces/one3/subjects", "CONNECT", 200, map[string]string{"namespace2": "one3"}},
		{"/api/v1/namespaces/one5/subjects", "CONNECT", 200, map[string]string{"namespace2": "one5"}},
		{"/api/v1/namespaces/one4/subjects", "GET", 201, map[string]string{"namespace2": "one4"}},
		{"/api/v1/namespaces/one/objects", "PUT", 3, map[string]string{"namespace": "one"}},
		{"/api/v1/namespaces/one/objects/two", "DELETE", 4, map[string]string{"namespace": "one", "object": "two"}},
		{"/api/v1/namespaces/one/objects/two2", "PUT", 400, map[string]string{"namespace": "one", "object": "two2"}},
		{"/api/v2/namespaces/one4/subjects", "GET", 201, map[string]string{"namespace2": "one4"}},
		{"/api/v3", "GET", 203, map[string]string{}},
		{"/api/v4", "GET", 204, map[string]string{}},
		{"/hello", "GET", 204, map[string]string{}},
	}
	wrong := []TestData{
		{"/api/v1/namespaces/", "GET", 0, nil},
		{"/api/v1/namespaces/one3/subjects", "POST", 0, nil},
		{"/api/v1/namespaces/one1/subjects", "CONNECT", 0, nil},
		{"/api/v1/namespaces/one12/subjects", "CONNECT", 0, nil},
		{"/api/v1/namespaces/", "GET", 0, nil},
		{"/api/v1/namespaces", "PUT", 0, nil},
		{"", "GET", 404, nil},
		{"api/v3", "GET", 203, nil},
		{"api/v4", "GET", 203, nil},
		{"/api/v9", "GET", 203, nil},
		{"/api/v5", "GET", 203, nil},
		{"/pages", "GET", 200, map[string]string{}},
		{"/pages/", "GET", 200, map[string]string{}},
	}
	testMatch(t, makeRouter(t, rds), right, wrong)
}

func TestFromChiTree(t *testing.T) {
	rds := []TestRouterData{
		{"/", getExecs, nil},
		{"/favicon.ico", getExecs, nil},
		{"/pages/{*:*}", getExecs, nil},
		{"/article", getExecs, nil},
		{"/article/", getExecs, nil},
		{"/article/{iffd}/edit", getExecs, nil},
		{"/article/{id}", getExecs, nil},
		{"/article/near", getExecs, nil},
		{"/article/@{user}", getExecs, nil},
		{"/article/{sup}/{opts}", getExecs, nil},
		{"/article/{id}/{opts}", getExecs, nil},
		{"/article/{id}//related", getExecs, nil},
		{"/article/slug/{month}/-/{day}/{year}", getExecs, nil},
		{"/admin/user", getExecs, nil},
		{"/admin/user/", getExecs, nil},
		{"/admin/user//{id}", getExecs, nil},
		{"/admin/user/{id}", getExecs, nil},
		{"/admin/apps/{id}", getExecs, nil},
		{"/admin/apps/{id}/{*:*}", getExecs, nil},
		{"/admin/*ff", getExecs, nil},
		{"/admin/{*:*}", getExecs, nil},
		{"/users/{userID}/profile", getExecs, nil},
		{"/users/super/{*:*}", getExecs, nil},
		{"/users/{*:*}", getExecs, nil},
		{"/hubs/{hubID}/view", getExecs, nil},
		{"/hubs/{hubID}/view/{*:*}", getExecs, nil},
		{"/users", getExecs, nil},
		{"/hubs/{hubID}/{*:*}", getExecs, nil},
		{"/hubs/{hubID}/users", getExecs, nil},
	}
	right := []TestData{
		{"/", "GET", 200, nil},
		{"/favicon.ico", "GET", 200, nil},
		{"/pages/yes", "GET", 200, map[string]string{"*": "yes"}},
		{"/article", "GET", 200, nil},
		{"/article/", "GET", 200, nil},
		{"/article/near", "GET", 200, nil},
		{"/article/neard", "GET", 200, map[string]string{"id": "neard"}},
		{"/article/123", "GET", 200, map[string]string{"id": "123"}},
		{"/article/123/456", "GET", 200, map[string]string{"id": "123", "opts": "456"}},
		{"/article/@peter", "GET", 200, map[string]string{"user": "peter"}},
		{"/article/22//related", "GET", 200, map[string]string{"id": "22"}},
		{"/article/111/edit", "GET", 200, map[string]string{"iffd": "111"}},
		{"/article/slug/sept/-/4/2015", "GET", 200, map[string]string{"month": "sept", "day": "4", "year": "2015"}},
		{"/article/:id", "GET", 200, map[string]string{"id": ":id"}},
		{"/admin/user", "GET", 200, nil},
		{"/admin/user/", "GET", 200, nil},
		{"/admin/user/1", "GET", 200, map[string]string{"id": "1"}},
		{"/admin/user//1", "GET", 200, map[string]string{"id": "1"}},
		{"/admin/hi", "GET", 200, map[string]string{"*": "hi"}},
		{"/admin/lots/of/:fun", "GET", 200, map[string]string{"*": "lots/of/:fun"}},
		{"/admin/apps/333", "GET", 200, map[string]string{"id": "333"}},
		{"/admin/apps/333/woot", "GET", 200, map[string]string{"id": "333", "*": "woot"}},
		{"/hubs/123/view", "GET", 200, map[string]string{"hubID": "123"}},
		{"/hubs/123/view/index.html", "GET", 200, map[string]string{"hubID": "123", "*": "index.html"}},
		{"/hubs/123/users", "GET", 200, map[string]string{"hubID": "123"}},
		{"/users/123/profile", "GET", 200, map[string]string{"userID": "123"}},
		{"/users/super/123/okay/yes", "GET", 200, map[string]string{"*": "123/okay/yes"}},
		{"/users/123/okay/yes", "GET", 200, map[string]string{"*": "123/okay/yes"}},
	}
	testMatch(t, makeRouter(t, rds), right, nil)
}

func TestFromChiTreeMoar(t *testing.T) {
	rds := []TestRouterData{
		{"/articlefun", getExecs, nil},
		{"/articles/{id}:delete", getExecs, nil},
		{"/articles/{iidd}!sup", getExecs, nil},
		{"/articles/{id}:{op}", getExecs, nil},
		{"/articles/{id}:{op}", getExecs, nil},
		{"/articles/{id}.json", getExecs, nil},
		{"/articles/{id}", getExecs, nil},
		{"/articles/{slug}", delExecs, nil},
		{"/articles/search", getExecs, nil},
		{"/articles/{slug:^[a-z]+}/posts", getExecs, nil},
		{"/articles/{id}/posts/{pid}", getExecs, nil},
		{"/articles/{id}/posts/{month}/{day}/{year}/{slug}", getExecs, nil},
		{"/articles/{id}/data.json", getExecs, nil},
		{"/articles/files/{file}.{ext}", getExecs, nil}, // TODO: Should we handle this case?
		{"/articles/me", putExecs, nil},
		{"/pages/*ff", getExecs, nil},
		{"/pages/{*:*}", getExecs, nil},
		{"/users/{id}", getExecs, nil},
		{"/users/{id}/settings/{key}", getExecs, nil},
		{"/users/{id}/settings/{*:*}", getExecs, nil},
	}
	right := []TestData{
		{"/articles/search", "GET", 200, nil},
		{"/articlefun", "GET", 200, nil},
		{"/articles/123", "GET", 200, map[string]string{"id": "123"}},
		{"/articles/123mm", "DELETE", 201, map[string]string{"slug": "123mm"}},
		{"/articles/789:delete", "GET", 200, map[string]string{"id": "789"}},
		{"/articles/789!sup", "GET", 200, map[string]string{"iidd": "789"}},
		{"/articles/123:sync", "GET", 200, map[string]string{"id": "123", "op": "sync"}},
		{"/articles/456/posts/1", "GET", 200, map[string]string{"id": "456", "pid": "1"}},
		{"/articles/456/posts/09/04/1984/juice", "GET", 200, map[string]string{"id": "456", "month": "09", "day": "04", "year": "1984", "slug": "juice"}},
		{"/articles/456.json", "GET", 200, map[string]string{"id": "456"}},
		{"/articles/456/data.json", "GET", 200, map[string]string{"id": "456"}},
		{"/articles/files/file.zip", "GET", 200, map[string]string{"file": "file", "ext": "zip"}},
		// {"/articles/files/photos.tar.gz", "GET", 200, map[string]string{"file": "photos", "ext": "tar.gz"}},
		// {"/articles/files/photos.tar.gz", "GET", 200, map[string]string{"file": "photos", "ext": "tar.gz"}},
		{"/articles/me", "PUT", 202, nil},
		{"/articles/me", "GET", 200, map[string]string{"id": "me"}},
		{"/pages/yes", "GET", 200, map[string]string{"*": "yes"}},
		{"/users/1", "GET", 200, map[string]string{"id": "1"}},
		{"/users/2/settings/password", "GET", 200, map[string]string{"id": "2", "key": "password"}},
	}
	testMatch(t, makeRouter(t, rds), right, nil)
}

func TestFromChiTreeRegexp(t *testing.T) {
	rds := []TestRouterData{
		{"/articles/{rid:^[0-9]{5,6}}", getExecs, nil},
		{"/articles/{zid:^0[0-9]+}", getExecs, nil},
		{"/articles/{name:^@[a-z]+}/posts", getExecs, nil},
		{"/articles/{op:^[0-9]+}/run", getExecs, nil},
		{"/articles/{id:^[0-9]+}", getExecs, nil},
		{"/articles/{id:^[1-9]+}-{aux}", getExecs, nil},
		{"/articles/{slug}", getExecs, nil},
	}
	right := []TestData{
		{"/articles/12345", "GET", 200, map[string]string{"rid": "12345"}},
		{"/articles/123", "GET", 200, map[string]string{"id": "123"}},
		{"/articles/how-to-build-a-router", "GET", 200, map[string]string{"slug": "how-to-build-a-router"}},
		{"/articles/0456", "GET", 200, map[string]string{"zid": "0456"}},
		{"/articles/@pk/posts", "GET", 200, map[string]string{"name": "@pk"}},
		{"/articles/1/run", "GET", 200, map[string]string{"op": "1"}},
		{"/articles/1122", "GET", 200, map[string]string{"id": "1122"}},
		{"/articles/1122-yes", "GET", 200, map[string]string{"id": "1122", "aux": "yes"}},
	}
	testMatch(t, makeRouter(t, rds), right, nil)
}

func TestFromChiTreeRegexMatchWholeParam(t *testing.T) {
	rds := []TestRouterData{
		{"/{id:[0-9]+}", getExecs, nil},
	}
	right := []TestData{
		{"/13", "GET", 200, map[string]string{"id": "13"}},
	}
	wrong := []TestData{
		{"/a13", "GET", 200, map[string]string{"id": "a13"}},
		{"/13.jpg", "GET", 200, map[string]string{"id": "13.jpg"}},
		{"/a13.jpg", "GET", 200, map[string]string{"id": "a13.jpg"}},
	}
	testMatch(t, makeRouter(t, rds), right, wrong)
}

type TestData struct {
	Path   string
	Type   string
	Result int
	Values map[string]string
}

type TestRouterData struct {
	Path       string
	Executor   []*TestExecutor
	Middleware []Middleware
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
	ins := c.Value(contextKey("Type"))
	if typ, ok := ins.(string); ok && typ == te.Type {
		return te, true
	}
	return nil, false
}

func (te *TestExecutor) Execute(c context.Context) error {
	result := c.Value(contextKey("Result"))
	if pointer, ok := result.(*int); ok {
		*pointer += te.Result
	} else {
		panic("can't find result from context")
	}
	return nil
}
