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

type Film struct {
	BaseModel
	Starships    []Identity `json:"starships"`
	Vehicles     []Identity `json:"vehicles"`
	Planets      []Identity `json:"planets"`
	Producer     string     `json:"producer"`
	Title        string     `json:"title"`
	Episode      Identity   `json:"episode"`
	Director     string     `json:"director"`
	OpeningCrawl string     `json:"opening_crawl"`
	Characters   []Identity `json:"characters"`
	Species      []Identity `json:"species"`
}
