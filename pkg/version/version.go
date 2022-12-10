/*
	Copyright 2022 Loophole Labs

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

package version

import (
	"fmt"
	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
	"github.com/spf13/cobra"
)

type Version interface {
	GitCommit() string
	GoVersion() string
	Platform() string
	Version() string
	BuildDate() string
}

func Format(version Version, cli string) string {
	if version.GitCommit() == "" && version.GoVersion() == "" && version.Platform() == "" || version.Version() == "" || version.BuildDate() == "" {
		return fmt.Sprintf("%s version (built from source)\n", cli)
	}

	return fmt.Sprintf("%s version %s (build date: %s git commit: %s go version: %s build platform: %s)\n", cli, version.Version(), version.BuildDate(), version.GitCommit(), version.GoVersion(), version.Platform())
}

// Cmd encapsulates the commands for showing a version
func Cmd[T config.Config](ch *cmdutils.Helper[T], version Version, cli string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ch.Printer.Format() == printer.Human {
				ch.Printer.Println(Format(version, cli))
				return nil
			}
			v := map[string]string{
				"version":    version.Version(),
				"commit":     version.GitCommit(),
				"build_date": version.BuildDate(),
				"go_version": version.GoVersion(),
				"platform":   version.Platform(),
			}
			return ch.Printer.PrintResource(v)
		},
	}

	return cmd
}
