// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type New[T Config] func() T

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
