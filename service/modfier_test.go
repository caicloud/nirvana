/*
Copyright 2020 Caicloud Authors

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

package service

import (
	"testing"

	"github.com/caicloud/nirvana/definition"
)

func TestProduceAllIfProducesIsEmpty(t *testing.T) {
	d := definition.Definition{
		Summary:     "Test",
		Method:      definition.Create,
		Description: "test",
		Results:     definition.DataErrorResults("data"),
		Function: func() (interface{}, error) {
			return nil, nil
		},
	}
	ProduceAllIfProducesIsEmpty()(&d)
	if len(d.Produces) == 0 {
		t.Errorf("Produces should not be empty")
	}
	if len(d.ErrorProduces) == 0 {
		t.Errorf("ErrorProduces should not be empty")
	}
}
