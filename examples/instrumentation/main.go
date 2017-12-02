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

package main

import (
	"log"
	"net/http"

	"github.com/caicloud/nirvana"
)

func main() {
	s, err := nirvana.New(&nirvana.Config{
		EnableMetrics:   true,
		EnableProfiling: true,
	})
	if err != nil {
		panic(err)
	}
	log.Printf("Listening on 8080")
	http.ListenAndServe(":8080", s)
}
