// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
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
