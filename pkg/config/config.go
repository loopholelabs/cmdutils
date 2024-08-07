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
