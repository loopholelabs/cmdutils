// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type testCommandFn func(*cmdutils.Helper[*TestConfig]) error

type TestCommandHarness struct {
	t                 *testing.T
	cmd               *Command[*TestConfig]
	stdout            *bytes.Buffer
	stderr            *bytes.Buffer
	config            *TestConfig
	defaultConfigFile string
	defaultLogFile    string
	fn                testCommandFn
}

func NewTestCommandHarness(t *testing.T, fn testCommandFn) *TestCommandHarness {
	h := &TestCommandHarness{
		t:                 t,
		fn:                fn,
		stdout:            new(bytes.Buffer),
		stderr:            new(bytes.Buffer),
		defaultConfigFile: filepath.Join(t.TempDir(), "test-config.yml"),
		defaultLogFile:    filepath.Join(t.TempDir(), "test-log.log"),
	}

	h.cmd = New("test", "A CLI test", "CLI test", true,
		version.New[*TestConfig]("", "", "", "", ""),
		NewTestConfigFn(h.defaultConfigFile, h.defaultLogFile),
		[]SetupCommand[*TestConfig]{
			h.setupCommandRun,
		},
	)
	h.cmd.stdout = h.stdout
	h.cmd.stderr = h.stderr

	return h
}

func (h *TestCommandHarness) setupCommandRun(rootCmd *cobra.Command, ch *cmdutils.Helper[*TestConfig]) {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run test function",
		RunE: func(cmd *cobra.Command, args []string) error {
			h.config = ch.Config
			if h.fn != nil {
				return h.fn(ch)
			}
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}

func (h *TestCommandHarness) Execute(ctx context.Context, args []string) int {
	h.cmd.command.SetArgs(args)
	return h.cmd.Execute(ctx, Noninteractive)
}

func (h *TestCommandHarness) Stdout() string {
	return h.stdout.String()
}

func (h *TestCommandHarness) Stderr() string {
	return h.stderr.String()
}

type TestConfig struct {
	cfgFile string
	logFile string

	// Common configuration values.
	Format  string
	Debug   bool
	NoColor bool `mapstructure:"no-color"`
}

func NewTestConfigFn(cfgFile string, logFile string) func() *TestConfig {
	return func() *TestConfig {
		return &TestConfig{
			cfgFile: cfgFile,
			logFile: logFile,
		}
	}
}

func (_ *TestConfig) RootPersistentFlags(flags *pflag.FlagSet)     { return }
func (_ *TestConfig) GlobalRequiredFlags(cmd *cobra.Command) error { return nil }
func (_ *TestConfig) Validate() error                              { return nil }
func (c *TestConfig) DefaultConfigDir() (string, error)            { return filepath.Dir(c.cfgFile), nil }
func (c *TestConfig) DefaultConfigFile() string                    { return c.cfgFile }
func (c *TestConfig) DefaultLogDir() (string, error)               { return filepath.Dir(c.logFile), nil }
func (c *TestConfig) DefaultLogFile() string                       { return c.logFile }
func (c *TestConfig) SetConfigFile(cfg string)                     { c.cfgFile = cfg }
func (c *TestConfig) GetConfigFile() string                        { return c.cfgFile }
func (c *TestConfig) SetLogFile(l string)                          { c.logFile = l }
func (c *TestConfig) GetLogFile() string                           { return c.logFile }
