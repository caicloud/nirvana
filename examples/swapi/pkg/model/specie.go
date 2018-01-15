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

package model

type Specie struct {
	BaseModel
	Classification  string     `json:"classification"`
	Designation     string     `json:"designation"`
	EyeColors       string     `json:"eye_colors"`
	People          []Identity `json:"people"`
	SkinColors      string     `json:"skin_colors"`
	Language        string     `json:"language"`
	Homeworld       *Identity  `json:"homeworld"`
	AverageLifespan string     `json:"average_lifespan"`
	AverageHeight   string     `json:"average_height"`
}
