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
	"github.com/spf13/cobra"
)

// Command contains a cobra.Command
type Command struct {
	*cobra.Command
}

// NewCommand returns a a Command
func NewCommand(cmd *cobra.Command) *Command {

	if cmd == nil {
		cmd = &cobra.Command{}
	}

	return &Command{
		Command: cmd,
	}
}

// AddFlag adds one or more commands to this command
func (c *Command) AddFlag(fs ...Flag) error {
	for _, f := range fs {
		flagset := c.Flags()
		if f.IsPersistent() {
			flagset = c.PersistentFlags()
		}
		err := f.ApplyTo(flagset)
		if err != nil {
			return err
		}
	}
	return nil
}

// Commands returns a sorted slice of child commands.
func (c *Command) Commands() []*Command {
	cmds := []*Command{}
	cobraCommands := c.Command.Commands()
	for _, cmd := range cobraCommands {
		cmds = append(cmds, NewCommand(cmd))
	}
	return cmds
}

// CobraCommands returns a sorted slice of child cobra commands.
func (c *Command) CobraCommands() []*cobra.Command {
	return c.Command.Commands()
}

// AddCommand adds one or more commands to this parent command.
func (c *Command) AddCommand(cmds ...*Command) {
	for _, x := range cmds {
		c.Command.AddCommand(x.Command)
	}
}

// AddCobraCommand adds one or more cobra commands to this parent command.
func (c *Command) AddCobraCommand(cmds ...*cobra.Command) {
	c.Command.AddCommand(cmds...)
}

// RemoveCommand removes one or more commands from a parent command.
func (c *Command) RemoveCommand(cmds ...*Command) {
	// remove commands from cobra
	cobraCommands := []*cobra.Command{}
	for _, cmd := range cmds {
		cobraCommands = append(cobraCommands, cmd.Command)
	}
	c.Command.RemoveCommand(cobraCommands...)
}

// RemoveCobraCommand removes one or more cobra commands from a parent command.
func (c *Command) RemoveCobraCommand(cmds ...*cobra.Command) {
	c.Command.RemoveCommand(cmds...)
}
