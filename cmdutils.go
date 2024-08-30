// SPDX-License-Identifier: Apache-2.0

package cmdutils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/loopholelabs/logging/types"

	"github.com/loopholelabs/cmdutils/pkg/config"
	"github.com/loopholelabs/cmdutils/pkg/printer"
)

// Helper is passed to every single command and is used by individual
// subcommands.
type Helper[T config.Config] struct {
	// Config contains globally sourced configuration
	Config T

	// Printer is used to print output of a command to stdout.
	Printer *printer.Printer

	// Logger is used to provide structured logging for a command.
	Logger types.RootLogger

	// debug defines the debug mode
	debug *bool
}

func (h *Helper[T]) SetDebug(debug *bool) {
	h.debug = debug
}

func (h *Helper[T]) Debug() bool { return *h.debug }

// RequiredArgs - required arguments are not available.
func RequiredArgs(reqArgs ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		n := len(reqArgs)
		if len(args) >= n {
			return nil
		}

		missing := reqArgs[len(args):]

		a := fmt.Sprintf("arguments <%s>", strings.Join(missing, ", "))
		if len(missing) == 1 {
			a = fmt.Sprintf("argument <%s>", missing[0])
		}

		return fmt.Errorf("missing %s \n\n%s", a, cmd.UsageString())
	}
}
