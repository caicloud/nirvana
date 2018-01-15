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

package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(nil)
	assert.NotNil(t, cmd)
}

func TestCommand_AddFlag(t *testing.T) {
	tests := []struct {
		name string
		fs   []Flag
	}{
		{"", []Flag{StringFlag{Name: "1"}, BoolFlag{Name: "2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommand(nil)
			c.AddFlag(tt.fs...)
		})
	}
}

func TestCommand_Commands(t *testing.T) {
	cs := []*Command{
		NewCommand(&cobra.Command{}),
		NewCommand(&cobra.Command{}),
	}
	c := &cobra.Command{}
	cmd := NewCommand(c)
	cmd.AddCommand(cs...)

	cmds := cmd.Commands()
	for i, cc := range cmds {
		assert.Equal(t, cs[i].Command, cc.Command)
	}
}

func TestCommand_CobraCommands(t *testing.T) {
	cs := []*cobra.Command{
		{},
		{},
	}
	c := &cobra.Command{}
	cmd := NewCommand(c)
	cmd.AddCobraCommand(cs...)

	cmds := cmd.CobraCommands()
	for i, cc := range cmds {
		assert.Equal(t, cs[i], cc)
	}
}

func TestCommand_RemoveCommand(t *testing.T) {
	cs := []*Command{
		NewCommand(&cobra.Command{}),
		NewCommand(&cobra.Command{}),
	}
	c := &cobra.Command{}
	cmd := NewCommand(c)
	cmd.AddCommand(cs...)

	cmd.RemoveCommand(cs...)

	cmds := cmd.Commands()
	assert.Len(t, cmds, 0)
}

func TestCommand_RemoveCobraCommand(t *testing.T) {
	cs := []*cobra.Command{
		{},
		{},
	}
	c := &cobra.Command{}
	cmd := NewCommand(c)
	cmd.AddCobraCommand(cs...)
	cmd.RemoveCobraCommand(cs...)

	cmds := cmd.CobraCommands()
	assert.Len(t, cmds, 0)
}
