# Cli - Mamba

The CLI pkg also named Mamba.

## About the project

Mamba provide a higher level and more friendly interfaces to build modern command line interfaces and manage configurations for Go applications.

It bases on

-   [spf13/pflag](https://github.com/spf13/pflag): to provide POSIX/GNU-style flags
-   [spf13/cobra](https://github.com/spf13/cobra): to provide a simple interface to create powerful modern CLI interfaces
-   [spf13/viper](https://github.com/spf13/viper): to provide a complete configuration sulution for [12-Factor](https://12factor.net/zh_cn/) apps

## Status

**Working in process**

## Concepts

A cli is composed of commands, argumetns & flags.

**Commands** represent actions, **Args** are things and **Flags** are modifiers for those actions.

The pattern to follow is `appname commands args --flag`

The flag is compatible with the [GNU extensions to the POSIX recommendations for command-line options](http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html).

## Getting Started

### Commands

You can migrate your cobra commands to mamba easily.

the `cli.NewCommand` receive a `*cobra.Command` to build the whole CLI.

```go
import (
	mamba "github.com/caicloud/nirvana/cli"
	"github.com/spf13/cobra"
)

func main() {
	cmd := mamba.NewCommand(&cobra.Command{
		Use:  "example",
		Long: "this is an cli example",
		Run: func(cmd *cobra.Command, args []string) {
			// you code
		},
	})
}
```

[More about cobra command](https://godoc.org/github.com/spf13/cobra#Command)

### Flags

Flag, different from `pflag.Flag`, is an interface implementing the following functions in Mamba.

And it is declarative in Mamba not imperative like Cobra. That makes the flags **more readable**.

Mamba will bind `viper` and `flags` automatically. That means you can get all your flag value from a configuration center.

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

#### Bind flag with ENV

Mamba supports the ability to bind your flags with env.

Mamba use the following precedence order. Each item takes precedence over the item bellow it.

-   explicit call to set
-   flag
-   env
-   default

```go
f := cli.StringFlag{
	Name: "log",
	EnvKey: "LOG",
  	DefValue: "test",
}
```

#### Automatic ENV and Prefix

Mamba supports the ability to bind you flags with ENV and add prefix automatically

The following functions exist to aid working with ENV

-   `AutomaticEnv()`
-   `SetEnvPrefix(string)`
-   `SetEnvKeyReplacer(*strings.Replacer)`

*When working with ENV variables, it’s important to recognize that Mamba treats ENV variables as case insensitive. All ENV variables are treated as UPPER case*

By using `AutomaticEnv` , you tell mamba to bind all flags with ENV automatically. 

`SetEnvPrefix` is **always** working with `AutomaticEnv` . It makes mamba add a prefix while reading from env variables.

By using  `SetEnvKeyReplacer` , you make mamba change the ENV by key replacer. For example, you can set a `UnderlineReplacer` to replace all `-` with `_` .

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

#### Hidden,Deprecated, ShorthandDeprecated 

-   `Hidden` and `Deprecated` hide the whole flag in help information.
-   `ShorthandDeprecated` hides the shorhand flag.

You can set the flag as usual.

-   `Hidden` just hide the flag
-   `Deprecated` and `ShorthandDeprecated` will print a deprecated message to you if you use the flag

### [WIP] Viper 

You can get a flag value from viper easily.

```go
cli.Viper.GetString("log")
```

>   Need further integration



## Thanks

spf13 creates awesome tools to make it easier to build a beautiful modern CLI apps.