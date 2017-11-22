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

package types

import (
	"time"
)

// go:generate flag-gen -i github.com/caicloud/nirvana/cmd/flag-gen/types -o github.com/caicloud/nirvana/cli

// FlagTypes ...
type FlagTypes struct {
	a bool
	// b []bool not support now
	c time.Duration
	d float32
	e float64
	f int
	// g []int not support now
	j int32
	k int64
	l string
	m []string
	n uint
	// o []uint not support now
	r uint32
	s uint64
}
