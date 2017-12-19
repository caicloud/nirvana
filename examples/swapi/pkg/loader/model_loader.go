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

package loader

import (
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"path"
	"io/ioutil"
	"encoding/json"
)

type ModelLoader interface {
	LoadPeople() []model.Person
	LoadFilms() []model.Film
	LoadPlanet() []model.Planet
	LoadSpecies() []model.Specie
	LoadStarships() []model.Starship
	LoadTransport() []model.Transport
}

type serializedForm struct {
	Pk     int                    `json:"pk"`
	Model  string                 `json:"model"`
	Fields map[string]interface{} `json:"fields"`
}

type modelLoader struct {
	dataPath string
}

func New(dataPath string) (*modelLoader, error) {
	m := &modelLoader{dataPath: dataPath}
	return m, nil
}

func (ml *modelLoader) loadJsonData(fileName string, v interface{}) error {
	filename := path.Join(ml.dataPath, fileName)
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	serializedData := make([]serializedForm, 0)
	if err = json.Unmarshal(raw, &serializedData); err != nil {
		return err
	}

	data := make([]map[string]interface{}, len(serializedData))
	for i, item := range serializedData {
		data[i] = item.Fields
		data[i]["id"] = item.Pk
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(encoded, v)
}

func (ml *modelLoader) LoadPeople() []model.Person {
	result := make([]model.Person, 0)
	if err := ml.loadJsonData("people.json", &result); err != nil {
		panic(err)
	}
	return result
}

func (ml *modelLoader) LoadFilms() []model.Film {
	result := make([]model.Film, 0)
	if err := ml.loadJsonData("films.json", &result); err != nil {
		panic(err)
	}
	return result
}

func (ml *modelLoader) LoadPlanet() []model.Planet {
	result := make([]model.Planet, 0)
	if err := ml.loadJsonData("planets.json", &result); err != nil {
		panic(err)
	}
	return result
}

func (ml *modelLoader) LoadSpecies() []model.Specie {
	result := make([]model.Specie, 0)
	if err := ml.loadJsonData("species.json", &result); err != nil {
		panic(err)
	}
	return result
}

func (ml *modelLoader) LoadStarships() []model.Starship {
	result := make([]model.Starship, 0)
	if err := ml.loadJsonData("starships.json", &result); err != nil {
		panic(err)
	}
	return result
}

func (ml *modelLoader) LoadTransport() []model.Transport {
	result := make([]model.Transport, 0)
	if err := ml.loadJsonData("transport.json", &result); err != nil {
		panic(err)
	}
	return result
}
