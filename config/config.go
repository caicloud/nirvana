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

// Option must be a pointer to struct.
type Option interface{}

// Plugin is for plugins to collect configurations
type Plugin interface {
	// Name returns plugin name.
	Name() string
	// Configure configures nirvana config via current options.
	Configure(cfg *nirvana.Config) error
}

// NirvanaStarter is a nirvana starter.
type NirvanaStarter interface {
	// EnablePlugin enables plugins.
	EnablePlugin(plugins ...Plugin) NirvanaStarter
	// AddOption will fill up options from config/ENV/flags after executing.
	AddOption(prefix string, options ...Option) NirvanaStarter
	// Add adds a field by key
	Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaStarter
	// Execute runs nirvana server.
	Execute(descriptors ...definition.Descriptor) error
	// ExecuteWithConfig runs nirvana server from a custom config.
	ExecuteWithConfig(cfg *nirvana.Config) error
	// Command returns a command for starter.
	Command(cfg *nirvana.Config) *cobra.Command
}

// NewNirvanaStarter creates a nirvana starter.
func NewNirvanaStarter() NirvanaStarter {
	return NewNamedNirvanaStarter("")
}

// NewNamedNirvanaStarter creates a nirvana starter with an unique name.
func NewNamedNirvanaStarter(name string) NirvanaStarter {
	return &starter{
		name:    name,
		plugins: []Plugin{},
		fields:  map[string]*configField{},
	}
}

type configField struct {
	pointer     interface{}
	key         string
	env         string
	shortFlag   string
	longFlag    string
	description string
}

type starter struct {
	name    string
	plugins []Plugin
	fields  map[string]*configField
}

// EnablePlugin enables plugins.
func (s *starter) EnablePlugin(plugins ...Plugin) NirvanaStarter {
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
func (s *starter) AddOption(prefix string, options ...Option) NirvanaStarter {
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

// Add adds a field by key
func (s *starter) Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaStarter {
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
func (s *starter) Execute(descriptors ...definition.Descriptor) error {
	cfg := nirvana.NewDefaultConfig("", 80)
	cfg.Configure(nirvana.Descriptor(descriptors...))
	return s.Command(cfg).Execute()
}

// ExecuteWithConfig runs nirvana server from a custom config.
func (s *starter) ExecuteWithConfig(cfg *nirvana.Config) error {
	return s.Command(cfg).Execute()
}

// Command returns a command for starter.
func (s *starter) Command(cfg *nirvana.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: s.name,
		Run: func(cmd *cobra.Command, args []string) {
			cfg.Logger.Info(banner)
			// Restore configs.
			for _, f := range s.fields {
				val := reflect.ValueOf(f.pointer).Elem().Interface()
				Set(f.key, val)
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
			fs.Uint8VarP(v, f.longFlag, f.shortFlag, val.(uint8), f.description+desc)
		case *uint16:
			if IsSet(f.key) {
				value = GetUint16(f.key)
			}
			if env != "" {
				envValue = cast.ToUint16(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Uint16VarP(v, f.longFlag, f.shortFlag, val.(uint16), f.description+desc)
		case *uint32:
			if IsSet(f.key) {
				value = GetUint32(f.key)
			}
			if env != "" {
				envValue = cast.ToUint32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Uint32VarP(v, f.longFlag, f.shortFlag, val.(uint32), f.description+desc)
		case *uint64:
			if IsSet(f.key) {
				value = GetUint64(f.key)
			}
			if env != "" {
				envValue = cast.ToUint64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Uint64VarP(v, f.longFlag, f.shortFlag, val.(uint64), f.description+desc)
		case *uint:
			if IsSet(f.key) {
				value = GetUint(f.key)
			}
			if env != "" {
				envValue = cast.ToUint(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.UintVarP(v, f.longFlag, f.shortFlag, val.(uint), f.description+desc)

		case *int8:
			if IsSet(f.key) {
				value = GetInt8(f.key)
			}
			if env != "" {
				envValue = cast.ToInt8(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Int8VarP(v, f.longFlag, f.shortFlag, val.(int8), f.description+desc)
		case *int16:
			if IsSet(f.key) {
				value = GetInt16(f.key)
			}
			if env != "" {
				envValue = cast.ToInt16(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Int16VarP(v, f.longFlag, f.shortFlag, val.(int16), f.description+desc)
		case *int32:
			if IsSet(f.key) {
				value = GetInt32(f.key)
			}
			if env != "" {
				envValue = cast.ToInt32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Int32VarP(v, f.longFlag, f.shortFlag, val.(int32), f.description+desc)
		case *int64:
			if IsSet(f.key) {
				value = GetInt64(f.key)
			}
			if env != "" {
				envValue = cast.ToInt64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Int64VarP(v, f.longFlag, f.shortFlag, val.(int64), f.description+desc)
		case *int:
			if IsSet(f.key) {
				value = GetInt(f.key)
			}
			if env != "" {
				envValue = cast.ToInt(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.IntVarP(v, f.longFlag, f.shortFlag, val.(int), f.description+desc)

		case *float32:
			if IsSet(f.key) {
				value = GetFloat32(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat32(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Float32VarP(v, f.longFlag, f.shortFlag, val.(float32), f.description+desc)
		case *float64:
			if IsSet(f.key) {
				value = GetFloat64(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat64(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.Float64VarP(v, f.longFlag, f.shortFlag, val.(float64), f.description+desc)

		case *string:
			if IsSet(f.key) {
				value = GetString(f.key)
			}
			if env != "" {
				envValue = cast.ToString(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.StringVarP(v, f.longFlag, f.shortFlag, val.(string), f.description+desc)
		case *[]string:
			if IsSet(f.key) {
				value = GetStringSlice(f.key)
			}
			if env != "" {
				envValue = cast.ToStringSlice(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.StringSliceVarP(v, f.longFlag, f.shortFlag, val.([]string), f.description+desc)

		case *bool:
			if IsSet(f.key) {
				value = GetBool(f.key)
			}
			if env != "" {
				envValue = cast.ToBool(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.BoolVarP(v, f.longFlag, f.shortFlag, val.(bool), f.description+desc)

		case *time.Duration:
			if IsSet(f.key) {
				value = GetDuration(f.key)
			}
			if env != "" {
				envValue = cast.ToDuration(env)
			}
			desc, val := chooseValue(f.key, value, f.env, envValue, *v)
			fs.DurationVarP(v, f.longFlag, f.shortFlag, val.(time.Duration), f.description+desc)

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
		desc += fmt.Sprintf(" (%s=%v)", key, value)
	} else {
		desc += fmt.Sprintf(" (%s)", key)
	}
	if envValue != nil {
		val = envValue
		desc = fmt.Sprintf(" (%s=%v)", env, envValue) + desc
	} else {
		desc = fmt.Sprintf(" (%s)", env) + desc
	}
	return desc, val
}
