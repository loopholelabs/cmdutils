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

package version

import (
	"fmt"
	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
	"github.com/spf13/cobra"
)

type Version[T config.Config] struct {
	gitCommit string
	goVersion string
	platform  string
	version   string
	buildDate string
}

func New[T config.Config](gitCommit string, goVersion string, platform string, version string, buildDate string) *Version[T] {
	return &Version[T]{
		gitCommit: gitCommit,
		goVersion: goVersion,
		platform:  platform,
		version:   version,
		buildDate: buildDate,
	}
}

func (v *Version[T]) GitCommit() string {
	return v.gitCommit
}

func (v *Version[T]) GoVersion() string {
	return v.goVersion
}

func (v *Version[T]) Platform() string {
	return v.platform
}

func (v *Version[T]) Version() string {
	return v.version
}

func (v *Version[T]) BuildDate() string {
	return v.buildDate
}

func (v *Version[T]) Format(cli string) string {
	if v.GitCommit() == "" && v.GoVersion() == "" && v.Platform() == "" || v.Version() == "" || v.BuildDate() == "" {
		return fmt.Sprintf("%s version (built from source)\n", cli)
	}

	return fmt.Sprintf("%s version %s (build date: %s git commit: %s go version: %s build platform: %s)\n", cli, v.Version(), v.BuildDate(), v.GitCommit(), v.GoVersion(), v.Platform())
}

func (v *Version[T]) Cmd(ch *cmdutils.Helper[T], cli string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ch.Printer.Format() == printer.Human {
				ch.Printer.Println(v.Format(cli))
				return nil
			}
			v := map[string]string{
				"version":    v.Version(),
				"commit":     v.GitCommit(),
				"build_date": v.BuildDate(),
				"go_version": v.GoVersion(),
				"platform":   v.Platform(),
			}
			return ch.Printer.PrintResource(v)
		},
	}

	return cmd
}
