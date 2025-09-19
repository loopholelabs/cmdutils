# CMD Utils

A CMD utility designed for use within Loophole Labs projects

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-brightgreen.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Discord](https://dcbadge.limes.pink/api/server/JYmFhtdPeu?style=flat)](https://loopholelabs.io/discord)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.22-61CFDD.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/loopholelabs/cmdutils.svg)](https://pkg.go.dev/github.com/loopholelabs/cmdutils)

## Overview

`cmdutils` is a comprehensive framework for building CLI applications in Go. It provides standardized configuration management, logging, command structure, and output formatting capabilities. Built on top of popular libraries like Cobra and Viper, it offers a consistent development experience across Loophole Labs projects.

## Features

- **Generic Configuration Management**: Type-safe configuration using Go generics
- **Multiple Output Formats**: Human-readable and JSON output support
- **Structured Logging**: File and console logging with multiple log levels
- **Interactive Mode**: Progress spinners and confirmation prompts
- **Version Management**: Built-in version command with detailed build information
- **Environment Variable Support**: Automatic binding with configurable prefix
- **Color Output**: Automatic TTY detection with disable option
- **Debug Mode**: Toggleable debug output for troubleshooting

## Installation

```bash
go get github.com/loopholelabs/cmdutils
```

## Quick Start

Here's a minimal example to get you started:

```go
package main

import (
   "context"
   "fmt"
   "os"
	
   "github.com/adrg/xdg"
   "github.com/spf13/cobra"
   "github.com/spf13/pflag"
   "github.com/spf13/viper"

   "github.com/loopholelabs/cmdutils"
   "github.com/loopholelabs/cmdutils/pkg/command"
   "github.com/loopholelabs/cmdutils/pkg/config"
   "github.com/loopholelabs/cmdutils/pkg/version"
)

var _ config.Config = (*Config)(nil)

const (
	defaultConfigName = "config.yaml"
	defaultLogName = "log.log"
)

var (
   configFile string
   logFile    string
)

type Config struct {
	APIKey string
	Port int
}

func New() *Config {
   return new(Config)
}

func (c *Config) RootPersistentFlags(flags *pflag.FlagSet) {
   flags.StringVar(&c.APIKey, "api-key", "", "API key used for authentication")
   flags.IntVarP(&c.Port, "port", "p", 8080, "Port to listen on")
}

func (c *Config) GlobalRequiredFlags(cmd *cobra.Command) error {
   // Mark required flags
   return cmd.MarkPersistentFlagRequired("api-key")
}

func (c *Config) Validate() error {
   if err := viper.Unmarshal(c); err != nil {
      return err
   }

   if c.Port < 1 || c.Port > 65535 {
      return fmt.Errorf("invalid port number: %d", c.Port)
   }

   return nil
}

func (c *Config) DefaultConfigDir() (string, error) {
   return xdg.ConfigHome, nil
}

func (c *Config) DefaultConfigFile() string {
   return defaultConfigName
}

func (c *Config) DefaultLogDir() (string, error) {
   return xdg.StateHome, nil
}

func (c *Config) DefaultLogFile() string {
   return defaultLogName
}

func (c *Config) SetConfigFile(file string) {
   configFile = file
}

func (c *Config) GetConfigFile() string {
   return configFile
}

func (c *Config) SetLogFile(file string) {
   logFile = file
}

func (c *Config) GetLogFile() string {
   return logFile
}

func main() {
    // Create version info
    ver := version.New[*Config](
        "git-commit-hash",
        "go1.21",
        "linux/amd64",
        "v1.0.0",
        "2024-01-01",
    )

    // Create command
    cmd := command.New[*Config](
        "myapp",
        "My CLI Application",
        "A longer description of my CLI application",
        false, // Allow arguments
        ver,
        func() *Config { return &Config{} },
        []command.SetupCommand[*Config]{setupSubcommands},
    )

    // Execute
    exitCode := cmd.Execute(context.Background(), command.Interactive)
    os.Exit(exitCode)
}

func setupSubcommands(root *cobra.Command, ch *cmdutils.Helper[*MyConfig]) {
    // Add your subcommands here
    root.AddCommand(&cobra.Command{
        Use:   "serve",
        Short: "Start the server",
        RunE: func(cmd *cobra.Command, args []string) error {
            ch.Printer.Printf("Starting server on port %d", ch.Config.Port)
            return nil
        },
    })
}
```

## Core Components

### 1. Configuration Interface

All configurations must implement the `config.Config` interface:

```go
type Config interface {
    RootPersistentFlags(flags *pflag.FlagSet)
    GlobalRequiredFlags(cmd *cobra.Command) error
    Validate() error
    DefaultConfigDir() (string, error)
    DefaultConfigFile() string
    DefaultLogDir() (string, error)
    DefaultLogFile() string
    SetConfigFile(configFile string)
    GetConfigFile() string
    SetLogFile(logFile string)
    GetLogFile() string
}
```

### 2. Helper Struct

The `Helper[T]` struct is passed to all commands and provides:

- `Config`: Your configuration instance
- `Printer`: Output formatting utilities
- `Logger`: Structured logging instance
- `Debug`: Debug mode flag

### 3. Printer

The printer handles output formatting:

```go
// Human-readable output
ch.Printer.Printf("Processing %d items...", count)

// JSON output (when --json flag is used)
ch.Printer.PrintJSON(myData)

// Resource output (automatically formats as table or YAML based on data structure)
ch.Printer.PrintResource(myData)

// Progress spinner
stop := ch.Printer.PrintProgress("Loading...")
// ... do work
stop()

// Interactive confirmation (requires specific format)
err := ch.Printer.ConfirmCommand("app-name", "delete", "deletion")
```

### 4. Logging

Structured logging with multiple levels:

```go
ch.Logger.Debug().Msg("Debug information")
ch.Logger.Info().Str("key", "value").Msg("Info message")
ch.Logger.Error().Err(err).Msg("Error occurred")
```

## Advanced Usage

### Environment Variables

Environment variables are automatically bound with a configurable prefix:

```go
// If your app is called "myapp", these environment variables are supported:
// MYAPP_API_KEY -> --api-key
// MYAPP_PORT -> --port
// MYAPP_FORMAT -> --format
// MYAPP_DEBUG -> --debug
```

### Configuration Files

Configuration files are automatically loaded from:
1. Path specified by `--config` flag
2. Current directory
3. Default config directory (OS-specific)

Example `myapp.yaml`:
```yaml
api-key: "your-api-key"
port: 9090
debug: true
```

### Custom Error Handling

Use `cmdutils.Error` for custom exit codes:

```go
return &cmdutils.Error{
    Msg: fmt.Sprintf("operation failed: %v", err),
    ExitCode: 2,
}
```

## Important Notes and Caveats

1. **Development Warning**: Self-compiled binaries show a development warning unless `MYAPP_DISABLE_DEV_WARNING=true` is set

2. **TTY Detection**: Output formatting automatically adjusts for TTY vs non-TTY environments

3. **Color Output**: Colors are automatically disabled in non-TTY environments or when `--no-color` is used

4. **Fatal Logging**: Using `ch.Logger.Fatal()` will call `os.Exit(1)`

5. **Spinner Cleanup**: Always call the returned stop function from `PrintProgress()` to ensure proper cleanup

6. **Config Search Order**: 
   - Command-line flags (highest priority)
   - Environment variables
   - Config file
   - Default values (lowest priority)

7. **JSON Output**: When `--format=json` is used, logging switches to structured JSON format

8. **Version Command**: The built-in `version` command is hidden by default but can be accessed with `myapp version`

9. **Exit Codes**: The library uses specific exit codes:
   - `0`: Success
   - `1`: Action requested exit (ActionRequestedExitCode)
   - `2`: Fatal error exit (FatalErrExitCode)
   - Custom exit codes via `cmdutils.Error`

## Contributing

Bug reports and pull requests are welcome on GitHub at [https://github.com/loopholelabs/cmdutils][gitrepo]. For more contribution information check out [the contribution guide](https://github.com/loopholelabs/cmdutils/blob/master/CONTRIBUTING.md).

## License

The CMD Utils project is available as open source under the terms of the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

## Code of Conduct

Everyone interacting in the CMD Utils project's codebases, issue trackers, chat rooms and mailing lists is expected to follow the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Project Managed By:

[![https://loopholelabs.io][loopholelabs]](https://loopholelabs.io)

[gitrepo]: https://github.com/loopholelabs/cmdutils
[loopholelabs]: https://cf-cdn.loopholelabs.io/loopholelabs-logo-light.svg
[loophomepage]: https://loopholelabs.io