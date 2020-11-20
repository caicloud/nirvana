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

package rest

import (
	"context"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/executor"
)

type inspector struct {
	path      string
	executors map[string][]executor.Executor
}

func newInspector(path string) *inspector {
	return &inspector{
		path:      path,
		executors: make(map[string][]executor.Executor),
	}
}

func (i *inspector) addDefinition(d definition.Definition) error {
	var method string
	if d.Method == definition.Any {
		method = string(definition.Any)
	} else {
		method = service.HTTPMethodFor(d.Method)
	}
	if method == "" {
		return executor.DefinitionNoMethod.Error(d.Method, i.path)
	}
	c, err := executor.DefinitionToExecutor(i.path, d, 0)
	if err != nil {
		return err
	}
	if err := i.conflictCheck(c, method); err != nil {
		return err
	}
	i.executors[method] = append(i.executors[method], c)
	return nil
}

func (i *inspector) conflictCheck(c executor.Executor, method string) error {
	cs := i.executors[method]
	if len(cs) <= 0 {
		return nil
	}
	ctMap := map[string]bool{}
	for _, extant := range cs {
		result := extant.ContentTypeMap()
		for k, vs := range result {
			for _, v := range vs {
				ctMap[k+":"+v] = true
			}
		}
	}
	cMap := c.ContentTypeMap()
	for k, vs := range cMap {
		for _, v := range vs {
			if ctMap[k+":"+v] {
				return executor.DefinitionConflict.Error(k, v, method, i.path)
			}
		}
	}
	return nil
}

// Inspect finds a valid executor to execute target context.
func (i *inspector) Inspect(ctx context.Context) (executor.MiddlewareExecutor, error) {
	httpCtx := service.HTTPContextFrom(ctx)
	req := httpCtx.Request()
	if req == nil {
		return nil, service.NoContext.Error()
	}
	executors := make([]executor.Executor, 0)
	if cs, ok := i.executors[req.Method]; ok && len(cs) > 0 {
		executors = append(executors, cs...)
	}
	if cs, ok := i.executors[string(definition.Any)]; ok && len(cs) > 0 {
		executors = append(executors, cs...)
	}
	if len(executors) <= 0 {
		return nil, noExecutorForMethod.Error()
	}
	ct, err := service.ContentType(req)
	if err != nil {
		return nil, err
	}
	accepted := 0
	for i, c := range executors {
		if c.Acceptable(ct) {
			if accepted != i {
				executors[accepted] = c
			}
			accepted++
		}
	}
	if accepted <= 0 {
		return nil, noExecutorForContentType.Error()
	}
	ats, err := service.AcceptTypes(req)
	if err != nil {
		return nil, err
	}
	executors = executors[:accepted]
	var target executor.Executor
	for _, c := range executors {
		if c.Producible(ats) {
			target = c
			break
		}
	}
	if target == nil {
		for _, at := range ats {
			if at == definition.MIMEAll {
				target = executors[0]
			}
		}
	}
	if target == nil {
		return nil, noExecutorToProduce.Error()
	}
	httpCtx.SetRoutePath(i.path)
	return target, nil
}
