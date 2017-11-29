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

package log

type Verboser interface {
	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})
}

type Logger interface {
	V(int) Verboser
	Verboser
	Warning(...interface{})
	Warningf(string, ...interface{})
	Warningln(...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})
}
