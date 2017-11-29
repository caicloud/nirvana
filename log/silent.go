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

type SilentLogger struct{}
type SilentVerboser struct{}

func (*SilentVerboser) Info(...interface{})               {}
func (*SilentVerboser) Infof(string, ...interface{})      {}
func (*SilentVerboser) Infoln(...interface{})             {}
func (*SilentLogger) V(v int) Verboser                    { return &SilentVerboser{} }
func (*SilentLogger) Info(...interface{})                 {}
func (*SilentLogger) Infof(string, ...interface{})        {}
func (*SilentLogger) Infoln(...interface{})               {}
func (*SilentLogger) Warning(v ...interface{})            {}
func (*SilentLogger) Warningf(f string, v ...interface{}) {}
func (*SilentLogger) Warningln(v ...interface{})          {}
func (*SilentLogger) Error(v ...interface{})              {}
func (*SilentLogger) Errorf(f string, v ...interface{})   {}
func (*SilentLogger) Errorln(v ...interface{})            {}
func (*SilentLogger) Fatal(v ...interface{})              {}
func (*SilentLogger) Fatalf(f string, v ...interface{})   {}
func (*SilentLogger) Fatalln(v ...interface{})            {}
