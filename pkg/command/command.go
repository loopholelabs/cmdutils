/*
	Copyright 2023 Loophole Labs

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

package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
	"github.com/loopholelabs/cmdutils/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

type SetupCommand[T config.Config] func(cmd *cobra.Command, ch *cmdutils.Helper[T])

type Command[T config.Config] struct {
	cli           string
	command       *cobra.Command
	version       *version.Version[T]
	new           config.New[T]
	config        T
	setupCommands []SetupCommand[T]
}

var (
	cfgFile  string
	replacer = strings.NewReplacer("-", "_", ".", "_")
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
		new:           newConfig,
		setupCommands: setupCommands,
	}
}

func (c *Command[T]) Execute(ctx context.Context) int {
	var format printer.Format
	var debug bool

	devEnv := fmt.Sprintf("%s_DISABLE_DEV_WARNING", replacer.Replace(c.cli))
	devWarning := fmt.Sprintf("!! WARNING: You are using a self-compiled binary which is not officially supported.\n!! To dismiss this warning, set %s=true\n\n", devEnv)

	if _, ok := os.LookupEnv(devEnv); !ok {
		if c.version.GitCommit() == "" || c.version.GoVersion() == "" || c.version.BuildDate() == "" || c.version.Version() == "" || c.version.Platform() == "" {
			_, _ = fmt.Fprintf(os.Stderr, devWarning)
		}
	}

	err := c.runCmd(ctx, &format, &debug)
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

	// check if a sub command wants to return a specific exit code
	var cmdErr *cmdutils.Error
	if errors.As(err, &cmdErr) {
		return cmdErr.ExitCode
	}

	return cmdutils.FatalErrExitCode
}

// runCmd adds all child commands to the root command, sets flags
// appropriately, and runs the root command.
func (c *Command[T]) runCmd(ctx context.Context, format *printer.Format, debug *bool) error {
	cobra.OnInitialize(c.initConfig)

	c.config = c.new()

	configPath, err := c.config.DefaultConfigPath()
	if err != nil {
		return err
	}

	c.command.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file (default is %s)", configPath))
	c.command.SilenceUsage = true
	c.command.SilenceErrors = true

	v := c.version.Format(c.cli)
	c.command.SetVersionTemplate(v)
	c.command.Version = v
	c.command.Flags().Bool("version", false, fmt.Sprintf("Show %s version", c.cli))

	c.config.RootFlags(c.command.PersistentFlags())

	c.command.PersistentFlags().VarP(printer.NewFormatValue(printer.Human, format), "format", "f", "Show output in a specific format. Possible values: [human, json, csv]")
	if err = viper.BindPFlag("format", c.command.PersistentFlags().Lookup("format")); err != nil {
		return err
	}
	_ = c.command.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"human", "json", "csv"}, cobra.ShellCompDirectiveDefault
	})

	c.command.PersistentFlags().BoolVar(debug, "debug", false, "Enable debug mode")
	if err := viper.BindPFlag("debug", c.command.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	ch := &cmdutils.Helper[T]{
		Printer: printer.NewPrinter(format),
		Config:  c.config,
	}
	ch.SetDebug(debug)

	c.command.PersistentFlags().BoolVar(&color.NoColor, "no-color", false, "Disable color output")
	if err := viper.BindPFlag("no-color", c.command.PersistentFlags().Lookup("no-color")); err != nil {
		return err
	}

	c.command.AddCommand(c.version.Cmd(ch, c.cli))

	for _, setup := range c.setupCommands {
		setup(c.command, ch)
	}

	return c.command.ExecuteContext(ctx)
}

// RequiredFlags is meant to be used by a pre-run function to mark flags as required
func (c *Command[T]) RequiredFlags() error {
	return c.config.CmdRequired(c.command)
}

// initConfig reads in config file and ENV variables if set.
func (c *Command[T]) initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := c.config.DefaultConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(cmdutils.FatalErrExitCode)
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
			fmt.Println(err)
			os.Exit(cmdutils.FatalErrExitCode)
		}
	}

	postInitCommands(c.command.Commands())
}

// Hacky fix for getting Cobra required flags and Viper playing well together.
// See: https://github.com/spf13/viper/issues/397
func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
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
}
