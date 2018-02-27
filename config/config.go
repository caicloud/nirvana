/*
Copyright 2018 Caicloud Authors

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

package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var banner = `
                                                '-:::::-
                                            .:/oo+++oo+:
                                         .:+++//+sys.'
                             .:.       :+++//+sysss+
                            /o++     :o++//osysos+.     ''.......'
                           oo/o-   -o+///ossoos+.'--/+oooooooooooooo/-.
                          +o//y  '+o///+ssooos::+oooooooossooooooooooooo'
                         's//o+ .s+///osoooohoooooosooo+:-...''''''''''.'
                     '// /+//y./s+//+ssooooosoooso/-.'  '-
                     s+s/s///ysh+///ysoooooooos+-'      yh'  ::
                    'h//yy///sd+///ssoooooooso:..'   '+-dys'-h+
                    'y//ys///+s///+hoooooooso'h++o+'.-hydhyyshy'
                     y//++////////ssooooooos 'y+/sso+++syyyyyhyo  ''
                     o+///////+s//yoooooooh:+oo+:/:::::::+++ooshss++o
                     's///////yh//ysoooooodo/:::::::::::::::::::/sss+
                   //:/s//////hyo/+hoooosh+:::::::::::::::::::::::/y.
                   s/+os////++hoh//ysoosh/::::::::::::+o///+o/:::::/h'
                   y////////hyhosy//yssh::::::::::::/o-     ':s/::::so.
                   +o//////+yshyosy+/sd/::::::::::::y      '+hyy::::o+:o.
                   'y//////+hoodyooy+/+ssoo+/+oooo::s      .mNsy::::s::dy
///.                /+//////ssoosyooss+////+++//+y::s:        ++:/ssh .ss
++/oo/.    o+/.      s////+yosyooooooossso++++//y/:::+o/-''-/o+s++syy-//
 +o///o+/. oo/+o/- -soh+///yysyhoooooooooooosyoso:::::::/++/::/s....o:
  .+o////+o/os+//+o/y++so///+ssssooooooooossh/-::/:::::::::::::o/oo//
    .+o+////+shy+///oy+/oso////++ossssssso+os+++/::::::::::::::+oy+/:
      '-+o+////+o+///////ossssoo+++++oooooo/:::::::::o/::::::::/h/::/
       s++sys+//yysssso+oysssm///+++////::::::::::::/s:::/+:::/y'''
       :+o+++oo/+ssssoosyyhyod:::::::::::::::::::::+o:::+h++:/s'
        '-+++/////+++shyyooooys::::::::::::::::::::s++++sooh+s'
           './oysoo///osssssssho::::::::::::::::::::/:::o+/so'
              ss+++//////////+oys/:::::::::::::::::::::::+o:
              '.:/+++++++++//-.'.soo/o::::::::::::::::/oo:'
                   ''''''        h/-./ss+++///////++oo/-'
                                 oo++/.''o+/y////::-.
                                         ://-
        ____  _____   _
       |_   \|_   _| (_)
         |   \ | |   __   _ .--.  _   __  ,--.   _ .--.   ,--.
         | |\ \| |  [  | [ '/''\][ \ [  ]''_\ : [ '.-. | ''_\ :
        _| |_\   |_  | |  | |     \ \/ / // | |, | | | | // | |,
       |_____|\____|[___][___]     \__/  \'-;__/[___||__]\'-;__/

`

// CustomOption must be a pointer to struct.
//
// Here is an example:
//   type Option struct {
//       FirstName string `desc:"Desc for First Name"`
//       Age       uint16 `desc:"Desc for Age"`
//   }
// The struct has two fields (with prefix example):
//   Field       Flag                   ENV                  Key (In config file)
//   FirstName   --example-first-name   EXAMPLE_FIRST_NAME   example.firstName
//   Age         --example-age          EXAMPLE_AGE          example.age
// When you execute command with `--help`, you can see the help doc of flags and
// descriptions (From field tag `desc`).
//
// The priority is:
//   Flag > ENV > Key > The value you set in option
type CustomOption interface{}

// Plugin is for plugins to collect configurations
type Plugin interface {
	// Name returns plugin name.
	Name() string
	// Configure configures nirvana config via current options.
	Configure(cfg *nirvana.Config) error
}

// Option contains basic configurations of nirvana.
type Option struct {
	// IP is the IP to listen.
	IP string `desc:"Nirvana server listening IP"`
	// Port is the port to listen.
	Port uint16 `desc:"Nirvana server listening Port"`
}

// NewDefaultOption creates a default option.
func NewDefaultOption() *Option {
	return &Option{
		IP:   "",
		Port: 8080,
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return "nirvana"
}

// Configure configures nirvana config via current option.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		nirvana.IP(p.IP),
		nirvana.Port(p.Port),
	)
	return nil
}

// NirvanaCommand is a nirvana command.
type NirvanaCommand interface {
	// EnablePlugin enables plugins.
	EnablePlugin(plugins ...Plugin) NirvanaCommand
	// AddOption will fill up options from flags/ENV/config after executing.
	// A non-empty prefix is recommended. It's used to divide option namespaces.
	AddOption(prefix string, options ...CustomOption) NirvanaCommand
	// Add adds a field by key.
	// If you don't have any struct to describe an option, you can use the method to
	// add a single field into nirvana command.
	// `pointer` must be a pointer to golang basic data type (e.g. *int, *string).
	// `key` must a config key. It's like 'nirvana.ip' and 'myconfig.name.firstName'.
	// The key will be converted to flag and env (e.g. --nirvana-ip and NIRVANA_IP).
	// If you want a short flag for the field, you can only set a one-char string.
	// `desc` describes the field.
	Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaCommand
	// Execute runs nirvana server.
	Execute(descriptors ...definition.Descriptor) error
	// ExecuteWithConfig runs nirvana server from a custom config.
	ExecuteWithConfig(cfg *nirvana.Config) error
	// Command returns a command for command.
	Command(cfg *nirvana.Config) *cobra.Command
}

// NewDefaultNirvanaCommand creates a nirvana command with default option.
func NewDefaultNirvanaCommand() NirvanaCommand {
	return NewNirvanaCommand(NewDefaultOption())
}

// NewNirvanaCommand creates a nirvana command. Nil option means default option.
func NewNirvanaCommand(option *Option) NirvanaCommand {
	return NewNamedNirvanaCommand("", option)
}

// NewNamedNirvanaCommand creates a nirvana command with an unique name.
func NewNamedNirvanaCommand(name string, option *Option) NirvanaCommand {
	if option == nil {
		option = NewDefaultOption()
	}
	cmd := &command{
		name:    name,
		option:  option,
		plugins: []Plugin{},
		fields:  map[string]*configField{},
	}
	cmd.EnablePlugin(cmd.option)
	return cmd
}

type configField struct {
	pointer     interface{}
	desired     interface{}
	key         string
	env         string
	shortFlag   string
	longFlag    string
	description string
}

type command struct {
	name    string
	option  *Option
	plugins []Plugin
	fields  map[string]*configField
}

// EnablePlugin enables plugins.
func (s *command) EnablePlugin(plugins ...Plugin) NirvanaCommand {
	s.plugins = append(s.plugins, plugins...)
	for _, plugin := range plugins {
		s.AddOption(plugin.Name(), plugin)
	}
	return s
}

func walkthrough(index []int, typ reflect.Type, f func(index []int, field reflect.StructField)) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			walkthrough(append(index, i), field.Type, f)
		} else {
			f(append(index, i), field)
		}
	}
}

// AddOption will fill up options from config/ENV/flags after executing.
func (s *command) AddOption(prefix string, options ...CustomOption) NirvanaCommand {
	if prefix != "" {
		prefix += "."
	}
	for _, opt := range options {
		val := reflect.ValueOf(opt)
		typ := val.Type()
		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			panic(errors.InternalServerError.Error("${type} is not a pointer to struct", typ.String()))
		}
		if val.IsNil() {
			panic(errors.InternalServerError.Error("${type} should not be nil", typ.String()))
		}
		val = val.Elem()
		typ = val.Type()
		walkthrough([]int{}, typ, func(index []int, field reflect.StructField) {
			ptr := val.FieldByIndex(index).Addr().Interface()
			s.Add(ptr, prefix+field.Name, "", field.Tag.Get("desc"))
		})
	}
	return s
}

// Add adds a field by key.
func (s *command) Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaCommand {
	if pointer == nil || reflect.ValueOf(pointer).IsNil() {
		panic(errors.InternalServerError.Error("pointer of ${key} should not be nil", key))
	}
	if s.name != "" {
		key = s.name + "." + key
	}
	fieldParts := strings.Split(key, ".")
	nameParts := make([][]string, len(fieldParts))
	for index, field := range fieldParts {
		parts := []string{}
		lastIsCapital := true
		lastIndex := 0
		for i, char := range field {
			if char >= '0' && char <= '9' {
				// Numbers inherit last char.
				continue
			}
			currentIsCapital := char >= 'A' && char <= 'Z'
			if i > 0 && lastIsCapital != currentIsCapital {
				end := 0
				if currentIsCapital {
					end = i
				} else {
					end = i - 1
				}
				if end > lastIndex {
					parts = append(parts, field[lastIndex:end])
					lastIndex = end
				}
			}
			lastIsCapital = currentIsCapital
		}
		if lastIndex < len(field) {
			parts = append(parts, field[lastIndex:])
		}
		nameParts[index] = parts
	}
	cf := &configField{
		pointer:     pointer,
		shortFlag:   shortFlag,
		description: desc,
	}
	for i, parts := range nameParts {
		if i > 0 {
			cf.key += "."
		}
		for j, part := range parts {
			if j == 0 {
				part = strings.ToLower(part)
			}
			cf.key += part
		}
	}
	for i, parts := range nameParts {
		if i > 0 {
			cf.longFlag += "-"
		}
		cf.longFlag += strings.Join(parts, "-")
	}
	cf.longFlag = strings.ToLower(cf.longFlag)
	for i, parts := range nameParts {
		if i > 0 {
			cf.env += "_"
		}
		cf.env += strings.Join(parts, "_")
	}
	cf.env = strings.ToUpper(cf.env)
	if _, ok := s.fields[cf.key]; ok {
		panic(errors.InternalServerError.Error("${key} has been registered", cf.key))
	}
	s.fields[cf.key] = cf
	return s
}

// Execute runs nirvana server.
func (s *command) Execute(descriptors ...definition.Descriptor) error {
	cfg := nirvana.NewDefaultConfig()
	cfg.Configure(nirvana.Descriptor(descriptors...))
	return s.Command(cfg).Execute()
}

// ExecuteWithConfig runs nirvana server from a custom config.
func (s *command) ExecuteWithConfig(cfg *nirvana.Config) error {
	return s.Command(cfg).Execute()
}

// Command returns a command for nirvana.
func (s *command) Command(cfg *nirvana.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: s.name,
		Run: func(cmd *cobra.Command, args []string) {
			fs := cmd.Flags()
			cfg.Logger.Info(banner)
			// Restore configs.
			for _, f := range s.fields {
				val := reflect.ValueOf(f.pointer).Elem()
				if f.desired != nil && !fs.Lookup(f.longFlag).Changed {
					val.Set(reflect.ValueOf(f.desired))
				}
				Set(f.key, val.Interface())
			}
			for _, plugin := range s.plugins {
				if err := plugin.Configure(cfg); err != nil {
					cfg.Logger.Fatalf("Failed to install plugin %s: %s", plugin.Name(), err.Error())
				}
			}
			cfg.Logger.Infof("Listening on %s:%d", cfg.IP, cfg.Port)
			if err := nirvana.NewServer(cfg).Serve(); err != nil {
				cfg.Logger.Fatal(err)
			}
		},
	}
	fs := cmd.Flags()
	for _, f := range s.fields {
		var value, envValue interface{} = nil, nil
		env := os.Getenv(f.env)
		switch v := f.pointer.(type) {
		case *uint8:
			if IsSet(f.key) {
				value = GetUint8(f.key)
			}
			if env != "" {
				envValue = cast.ToUint8(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Uint8VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *uint16:
			if IsSet(f.key) {
				value = GetUint16(f.key)
			}
			if env != "" {
				envValue = cast.ToUint16(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Uint16VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *uint32:
			if IsSet(f.key) {
				value = GetUint32(f.key)
			}
			if env != "" {
				envValue = cast.ToUint32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Uint32VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *uint64:
			if IsSet(f.key) {
				value = GetUint64(f.key)
			}
			if env != "" {
				envValue = cast.ToUint64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Uint64VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *uint:
			if IsSet(f.key) {
				value = GetUint(f.key)
			}
			if env != "" {
				envValue = cast.ToUint(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.UintVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		case *int8:
			if IsSet(f.key) {
				value = GetInt8(f.key)
			}
			if env != "" {
				envValue = cast.ToInt8(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Int8VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *int16:
			if IsSet(f.key) {
				value = GetInt16(f.key)
			}
			if env != "" {
				envValue = cast.ToInt16(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Int16VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *int32:
			if IsSet(f.key) {
				value = GetInt32(f.key)
			}
			if env != "" {
				envValue = cast.ToInt32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Int32VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *int64:
			if IsSet(f.key) {
				value = GetInt64(f.key)
			}
			if env != "" {
				envValue = cast.ToInt64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Int64VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *int:
			if IsSet(f.key) {
				value = GetInt(f.key)
			}
			if env != "" {
				envValue = cast.ToInt(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.IntVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		case *float32:
			if IsSet(f.key) {
				value = GetFloat32(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Float32VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *float64:
			if IsSet(f.key) {
				value = GetFloat64(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.Float64VarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		case *string:
			if IsSet(f.key) {
				value = GetString(f.key)
			}
			if env != "" {
				envValue = cast.ToString(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.StringVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)
		case *[]string:
			if IsSet(f.key) {
				value = GetStringSlice(f.key)
			}
			if env != "" {
				envValue = cast.ToStringSlice(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.StringSliceVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		case *bool:
			if IsSet(f.key) {
				value = GetBool(f.key)
			}
			if env != "" {
				envValue = cast.ToBool(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.BoolVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		case *time.Duration:
			if IsSet(f.key) {
				value = GetDuration(f.key)
			}
			if env != "" {
				envValue = cast.ToDuration(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			f.desired = val
			fs.DurationVarP(v, f.longFlag, f.shortFlag, *v, f.description+desc)

		default:
			panic(errors.InternalServerError.Error("unrecognized type ${type} for ${key}", reflect.TypeOf(f.pointer).String(), f.key))
		}
	}
	return cmd
}

func chooseValue(key string, value interface{}, env string, envValue interface{}, defaultValue interface{}) (string, interface{}) {
	val := defaultValue
	desc := ""
	if value != nil {
		val = value
		desc += fmt.Sprintf(" (cfg %s=%v)", key, value)
	} else {
		desc += fmt.Sprintf(" (cfg %s)", key)
	}
	if envValue != nil {
		val = envValue
		desc = fmt.Sprintf(" (env %s=%v)", env, envValue) + desc
	} else {
		desc = fmt.Sprintf(" (env %s)", env) + desc
	}
	return desc, val
}
