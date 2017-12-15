# CLI

## About the project

CLI provides a higher level and more friendly interfaces to build modern command line interfaces and manage configurations for Go applications.

It is based on

-   [spf13/pflag](https://github.com/spf13/pflag): to provide POSIX/GNU-style flags
-   [spf13/cobra](https://github.com/spf13/cobra): to provide a simple interface to create powerful modern CLI interfaces
-   [spf13/viper](https://github.com/spf13/viper): to provide a complete configuration sulution for [12-Factor](https://12factor.net) apps

## Status

**Working in process**

## Concepts

A cli is composed of commands, arguments and flags.

**Commands** represent actions, **Args** are things and **Flags** are modifiers for those actions.

The pattern to follow is `appname commands args --flag`

The flag is compatible with the [GNU extensions to the POSIX recommendations for command-line options](http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html).

## Getting Started

### Commands

You can migrate your cobra commands to CLI easily. The `cli.NewCommand` receives a `*cobra.Command` to build the command line interface.

```go
import (
	"github.com/caicloud/nirvana/cli"
	"github.com/spf13/cobra"
)

func main() {
	cmd := cli.NewCommand(&cobra.Command{
		Use:  "example",
		Long: "this is an cli example",
		Run: func(cmd *cobra.Command, args []string) {
			// your code
		},
	})
}
```

[More about cobra command](https://godoc.org/github.com/spf13/cobra#Command)

### Flags

Flag, different from `pflag.Flag`, is an interface containing the following methods in CLI. And it is declarative in CLI, unlike in Cobra where it is imperative, which makes the flags **more readable**.

```go
// Flag describes a flag interface
type Flag interface {
	// IsPersistent specify whether the flag is persistent
	IsPersistent() bool
	// GetName returns the flag's name
	GetName() string
	// ApplyTo adds the flag to a given FlagSet
	ApplyTo(*pflag.FlagSet) error
}
```

A most widely used string flag is defined like

```go
type StringFlag struct {
	// Name as it appears on command line
	Name string
	// one-letter abbreviated flag
	Shorthand string
	// help message
	Usage string
	// specify whether the flag is persistent
	Persistent bool
	// used by cobra.Command bash autocomple code
	Annotations map[string][]string
	// If this flag is deprecated, this string is the new or now thing to use
	Deprecated string
	// If the shorthand of this flag is deprecated, this string is the new or now thing to use
	ShorthandDeprecated string
	// used by cobra.Command to allow flags to be hidden from help/usage text
	Hidden bool
	// bind the flag to env key, you can use AutomaticEnv to bind all flags to env automatically
	// if EnvKey is set, it will override the automatic generated env key
	EnvKey string
	// the default value
	DefValue string
	// points to a variable in which to store the value of the flag
	Destination *string
}
```

#### Binding flag with ENV

CLI supports the ability to bind your flags with ENV variables.

CLI uses the following precedence order. Each item takes precedence over the item below it.

-   explicit calls to set
-   flag
-   env
-   default value

```go
f := cli.StringFlag{
	Name: "log",
	EnvKey: "LOG",
  	DefValue: "test",
}
```

#### Automatic ENV and Prefix

CLI supports the ability to bind you flags with ENV and add prefix automatically

The following methods exist to aid working with ENV:

-   `AutomaticEnv()`: It tells CLI to bind all flags with ENV automatically. 
-   `SetEnvPrefix(string)`: It is **always** working with `AutomaticEnv` . It makes CLI add a prefix while reading from env variables.
-   `SetEnvKeyReplacer(*strings.Replacer)`: It makes CLI change the ENV by key replacer. For example, you can set a `UnderlineReplacer` to replace all `-` with `_` .

*When working with ENV variables, it’s important to recognize that CLI treats ENV variables case insensitively. All ENV variables are treated as UPPER case*

>   **Note: If EnvKey is set, it will override the automatic env and does not automatically add the prefix.**

**Example**

```go
AutomaticEnv()
SetEnvPrefix("nirvana")

log = new(string)

f := cli.StringFlag{
	Name: "log",
	Destination: log,
}

os.Setenv("NIRVANA_LOG", "TEST")

cmd.AddFlag(f)

// *log is TEST
```

#### Hidden, Deprecated, ShorthandDeprecated 

-   `Hidden` and `Deprecated` hide the whole flag in help information.
-   `ShorthandDeprecated` hides the shorhand flag.

You can set the flag as usual.

-   `Hidden` just hide the flag
-   `Deprecated` and `ShorthandDeprecated` will print a deprecated message to you if you use the flag

#### Persistent

Persistent option makes the flag can be inherited by it children‘s commands

### Configuration 

CLI can be also treated as a configuration registry.

*The flags are bound to registry automatically. That means all defined flags' values can be accessed by `Get()` function.*

```go
cmd.AddFlag(cli.StringFlag{Name: "log", DefValue: "test configuration"})
cli.GetString("log") // test configuration
```

#### Reading Config Files

There are two ways for you to let CLI know where to look for config files.

1.  `SetConfigFile(in string)`
2.  `SetConfigPaths(noExtName string, paths …string)`

`SetConfigFile` explicitly defines the path, name and extension of the config file. CLI will use this and not check any of the config paths.

`SetConfigPaths` defines a config file name without the extension and paths where CLI search the config file in. 

Then, use

-   `ReadInConfig`
-   `MergeInConfig`

to load configuration files.

The difference between `ReadInConfig` and `MergeInConfig` is that `ReadInConfig` discards all existing config but `MergeInConfig` merges the configuration with existing one.

#### Watching and reloading config files

After setting config file path, working with `WatchConfig(onChange func(in fsnotify.Event))` to watch the config file changes.

If the watched config file is created/deleted/updated, CLI will read new values from config file automatically. Then the onChange callback function is invoked.

```go
cli.SetConfigFile("/etc/nirvana/config.json")
cli.WatchConfig(func(e fsnotify.Event){
	fmt.Println("Config file changed: ", e.Name)
})
```

#### Reading Config from io.Reader

CLI supports the ability to let you implement your own required configuration source and feed it to CLI.

But the prerequisite is that CLI should known what type it is.

Using `SetConfigType(in string)` to tell CLI the type of configuration. The following type are supported now:

-   json
-   toml
-   yaml or yml
-   hcl
-   properties, props, prop

#### Getting Values From CLI

The following methods exist to aid getting values from CLI.

-   `Get(key string) interface{}`
-   `IsSet(key string) bool`
-   `GetBool(key string) bool`
-   `GetDuration(key string) time.Duration`
-   `GetFloat32(key string) float32`
-   `GetFloat64(key string) float64`
-   `GetInt(key string) int`
-   `GetInt32(key string) int32`
-   `GetInt64(key string) int64`
-   `GetString(key string) string`
-   `GetStringSlice(key string) []string`
-   `GetUint(key string) uint`
-   `GetUint32(key string) uint32`
-   `GetUint64(key string) uint64`

>   Note that each Get function will return a zero value if it is not found. To check if a given key exists, please use `IsSet()`

## Thanks

Thanks spf13 for creating awesome tools to make it easier to build beautiful modern CLI apps.
