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

package people

import "github.com/caicloud/nirvana/examples/swapi/pkg/model"

type ById []model.Person

func (r ById) Len() int {
	return len(r)
}

func (r ById) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ById) Less(i, j int) bool {
	return r[i].Id < r[j].Id
}
