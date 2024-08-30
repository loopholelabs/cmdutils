// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/loopholelabs/logging"
	"github.com/loopholelabs/logging/types"

	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
	"github.com/loopholelabs/cmdutils/pkg/version"
)

type Type int

const (
	Interactive Type = iota
	Noninteractive
)

type SetupCommand[T config.Config] func(cmd *cobra.Command, ch *cmdutils.Helper[T])

type Command[T config.Config] struct {
	cli           string
	command       *cobra.Command
	version       *version.Version[T]
	newConfig     config.New[T]
	config        T
	setupCommands []SetupCommand[T]
}

var (
	cfgFile        string
	logFile        string
	logOutput      io.Writer
	logClosersLock sync.Mutex
	logClosers     []func() error
	replacer       = strings.NewReplacer("-", "_", ".", "_")
)

func New[T config.Config](cli string, short string, long string, noargs bool, version *version.Version[T], newConfig config.New[T], setupCommands []SetupCommand[T]) *Command[T] {
	c := &cobra.Command{
		Use:              cli,
		Short:            short,
		Long:             long,
		TraverseChildren: true,
	}
	if noargs {
		c.Args = cobra.NoArgs
	}
	return &Command[T]{
		cli:           cli,
		command:       c,
		version:       version,
		newConfig:     newConfig,
		setupCommands: setupCommands,
	}
}

func (c *Command[T]) Execute(ctx context.Context, commandType Type) int {
	var format printer.Format
	var debug bool

	devEnv := fmt.Sprintf("%s_DISABLE_DEV_WARNING", strings.ToUpper(replacer.Replace(c.cli)))
	devWarning := fmt.Sprintf("!! WARNING: You are using a self-compiled binary which is not officially supported.\n!! To dismiss this warning, set %s=true\n\n", devEnv)

	if _, ok := os.LookupEnv(devEnv); !ok {
		if c.version.GitCommit() == "" || c.version.GoVersion() == "" || c.version.BuildDate() == "" || c.version.Version() == "" || c.version.Platform() == "" {
			_, _ = fmt.Fprintf(os.Stderr, devWarning)
		}
	}

	err := c.runCmd(ctx, &format, &debug, commandType)
	if err == nil {
		return 0
	}

	// print any user specific messages first
	switch format {
	case printer.JSON:
		_, _ = fmt.Fprintf(os.Stderr, `{"error": "%s"}`, err)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	logClosersLock.Lock()
	defer logClosersLock.Unlock()
	for _, closeLog := range logClosers {
		_ = closeLog()
	}

	// check if a sub command wants to return a specific exit code
	var cmdErr *cmdutils.Error
	if errors.As(err, &cmdErr) {
		return cmdErr.ExitCode
	}

	return cmdutils.FatalErrExitCode
}

// runCmd adds all child commands to the root command, sets flags
// appropriately, and runs the root command.
func (c *Command[T]) runCmd(ctx context.Context, format *printer.Format, debug *bool, commandType Type) error {
	c.config = c.newConfig()

	configDir, err := c.config.DefaultConfigDir()
	if err != nil {
		return err
	}

	logDir, err := c.config.DefaultLogDir()
	if err != nil {
		return err
	}

	configPath := path.Join(configDir, c.config.DefaultConfigFile())
	logPath := ""
	if commandType == Interactive {
		logPath = path.Join(logDir, c.config.DefaultLogFile())
	}

	c.command.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`Config file (default "%s")`, configPath))
	c.command.PersistentFlags().StringVar(&logFile, "log", logPath, "Log file")

	ch := &cmdutils.Helper[T]{
		Config: c.config,
	}

	cobra.OnInitialize(func() {
		err := c.initConfig()
		if err != nil {
			switch *format {
			case printer.JSON:
				_, _ = fmt.Fprintf(os.Stderr, `{"error": "%s"}`, err)
			default:
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}

			os.Exit(cmdutils.FatalErrExitCode)
		}

		ch.SetDebug(debug)

		ch.Printer = printer.NewPrinter(format)

		if strings.TrimSpace(logFile) == "" {
			logOutput = os.Stderr
		} else {
			if err := os.MkdirAll(filepath.Dir(logFile), 0700); err != nil {
				switch *format {
				case printer.JSON:
					_, _ = fmt.Fprintf(os.Stderr, `{"error": "%s"}`, err)
				default:
					_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				}

				os.Exit(cmdutils.FatalErrExitCode)
			}

			fileLogOutput, err := os.OpenFile(logFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
			if err != nil {
				switch *format {
				case printer.JSON:
					_, _ = fmt.Fprintf(os.Stderr, `{"error": "%s"}`, err)
				default:
					_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				}

				os.Exit(cmdutils.FatalErrExitCode)
			}

			logClosersLock.Lock()
			defer logClosersLock.Unlock()
			logClosers = append(logClosers, fileLogOutput.Close)

			if ch.Debug() {
				logOutput = io.MultiWriter(fileLogOutput, os.Stderr)
			} else {
				logOutput = fileLogOutput
			}
		}

		switch *format {
		case printer.JSON:
			ch.Logger = logging.New(logging.Zerolog, strings.ToLower(c.cli), logOutput)
		default:
			ch.Logger = logging.New(logging.Slog, strings.ToLower(c.cli), logOutput)
		}

		if ch.Debug() {
			ch.Logger.SetLevel(types.TraceLevel)
		} else {
			ch.Logger.SetLevel(types.InfoLevel)
		}
	})

	c.command.SilenceUsage = true
	c.command.SilenceErrors = true

	v := c.version.Format(c.cli)
	c.command.SetVersionTemplate(v)
	c.command.Version = v
	c.command.Flags().Bool("version", false, fmt.Sprintf("Show %s version", c.cli))

	c.config.RootPersistentFlags(c.command.PersistentFlags())

	c.command.PersistentFlags().VarP(printer.NewFormatValue(printer.Human, format), "format", "f", "Show output in a specific format. Possible values: [human, json]")
	if err = viper.BindPFlag("format", c.command.PersistentFlags().Lookup("format")); err != nil {
		return err
	}
	_ = c.command.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"human", "json"}, cobra.ShellCompDirectiveDefault
	})

	c.command.PersistentFlags().BoolVar(debug, "debug", false, "Enable debug mode")
	if err = viper.BindPFlag("debug", c.command.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	c.command.PersistentFlags().BoolVar(&color.NoColor, "no-color", false, "Disable color output")
	if err = viper.BindPFlag("no-color", c.command.PersistentFlags().Lookup("no-color")); err != nil {
		return err
	}

	c.command.AddCommand(c.version.Cmd(ch, c.cli))

	for _, setup := range c.setupCommands {
		setup(c.command, ch)
	}

	return c.command.ExecuteContext(ctx)
}

// initConfig reads in config file and ENV variables if set.
func (c *Command[T]) initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := c.config.DefaultConfigDir()
		if err != nil {
			return fmt.Errorf("failed to read default configuration directory: %w", err)
		}

		viper.AddConfigPath(configDir)

		configFile := c.config.DefaultConfigFile()
		configFileSplit := strings.Split(configFile, ".")
		viper.SetConfigName(configFileSplit[0])
		if len(configFileSplit) > 1 {
			viper.SetConfigType(configFileSplit[1])
		}
	}

	viper.SetEnvPrefix(strings.ToUpper(c.cli))
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only handle errors when it's something unrelated to the config file not
			// existing.
			return fmt.Errorf("failed to read configuration: %w", err)
		}
	}
	err := viper.Unmarshal(c.config, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	c.postInitCommands(c.command.Commands())

	if c.config.GetConfigFile() != "" {
		err := os.MkdirAll(filepath.Dir(c.config.GetConfigFile()), 0700)
		if err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("failed to create configuration directory: %w", err)
			}
		}
	}

	return nil
}

// Hacky fix for getting Cobra required flags and Viper playing well together.
// See: https://github.com/spf13/viper/issues/397
func (c *Command[T]) postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		c.presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			c.postInitCommands(cmd.Commands())
		}
	}
}

func (c *Command[T]) presetRequiredFlags(cmd *cobra.Command) {
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		log.Fatalf("error binding flags: %v", err)
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			err = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
			if err != nil {
				log.Fatalf("error setting flag %s: %v", f.Name, err)
			}
		}
	})

	c.config.SetConfigFile(viper.ConfigFileUsed())
	c.config.SetLogFile(logFile)
}
