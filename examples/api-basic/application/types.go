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

package application

type ApplicationV1 struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	Replica   int    `json:"replica"`
	Phase     string `json:"phase"`
	Message   string `json:"message"`
}

type ApplicationV2 struct {
	Metadata Metadata            `json:"metadata"`
	Spec     ApplicationV2Spec   `json:"spec"`
	Status   ApplicationV2Status `json:"status"`
}

type Metadata struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
}

type ApplicationV2Spec struct {
	Replica     int    `json:"replica"`
	OtherFields string `json:"other"`
}

type ApplicationV2Status struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
}
