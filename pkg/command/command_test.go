// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"testing"

	"github.com/loopholelabs/cmdutils"
	"github.com/loopholelabs/logging/loggers/zerolog"
	"github.com/loopholelabs/logging/types"
	"github.com/stretchr/testify/require"
)

func TestLogLevel(t *testing.T) {
	t.Setenv("TEST_DISABLE_DEV_WARNING", "true")

	testCases := []struct {
		name          string
		args          []string
		expectedLevel types.Level
		expectError   bool
	}{
		{
			name:          "default is info",
			args:          []string{"run"},
			expectedLevel: types.InfoLevel,
		},
		{
			name:          "case insensitive",
			args:          []string{"run", "--log-level", "WarN"},
			expectedLevel: types.WarnLevel,
		},
		{
			name:        "missing",
			args:        []string{"run", "--log-level", ""},
			expectError: true,
		},
		{
			name:        "invalid",
			args:        []string{"run", "--log-level", "not-valid"},
			expectError: true,
		},
		{
			name:          "trace",
			args:          []string{"run", "--log-level", "trace"},
			expectedLevel: types.TraceLevel,
		},
		{
			name:          "debug",
			args:          []string{"run", "--log-level", "debug"},
			expectedLevel: types.DebugLevel,
		},
		{
			name:          "info",
			args:          []string{"run", "--log-level", "info"},
			expectedLevel: types.InfoLevel,
		},
		{
			name:          "warn",
			args:          []string{"run", "--log-level", "warn"},
			expectedLevel: types.WarnLevel,
		},
		{
			name:          "error",
			args:          []string{"run", "--log-level", "error"},
			expectedLevel: types.ErrorLevel,
		},
		{
			name:          "fatal",
			args:          []string{"run", "--log-level", "fatal"},
			expectedLevel: types.FatalLevel,
		},
	}

	fn := func(ch *cmdutils.Helper[*TestConfig]) error {
		ch.Logger.Trace().Msg("TRACE")
		ch.Logger.Debug().Msg("DEBUG")
		ch.Logger.Info().Msg("INFO")
		ch.Logger.Warn().Msg("WARN")
		ch.Logger.Error().Msg("ERROR")

		// Skip FATAL level with zerolog because it calls os.Exit().
		if _, ok := ch.Logger.(*zerolog.Logger); !ok {
			ch.Logger.Fatal().Msg("FATAL")
		}
		return nil
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/default", tc.name), func(t *testing.T) {
			h := NewTestCommandHarness(t, fn)

			rc := h.Execute(context.Background(), tc.args)
			if tc.expectError {
				require.NotZero(t, rc)
				require.Contains(t, h.Stderr(), "--log-level")
				return
			}

			require.Zero(t, rc, "expected no error, got:\n%s", h.Stderr())
			require.Equal(t, tc.expectedLevel, h.cmd.logLevel)

			for l := types.FatalLevel; l <= types.TraceLevel; l++ {
				msg := fmt.Sprintf("msg=%s", l)
				if tc.expectedLevel >= l {
					require.Contains(t, h.Stderr(), msg)
				} else {
					require.NotContains(t, h.Stderr(), msg)
				}
			}
		})

		t.Run(fmt.Sprintf("%s/human", tc.name), func(t *testing.T) {
			h := NewTestCommandHarness(t, fn)

			args := make([]string, len(tc.args))
			copy(args, tc.args)
			args = append(args, "--format=human")

			rc := h.Execute(context.Background(), args)
			if tc.expectError {
				require.NotZero(t, rc)
				require.Contains(t, h.Stderr(), "--log-level")
				return
			}

			require.Zero(t, rc, "expected no error, got:\n%s", h.Stderr())
			require.Equal(t, tc.expectedLevel, h.cmd.logLevel)

			for l := types.FatalLevel; l <= types.TraceLevel; l++ {
				msg := fmt.Sprintf("msg=%s", l)
				if tc.expectedLevel >= l {
					require.Contains(t, h.Stderr(), msg)
				} else {
					require.NotContains(t, h.Stderr(), msg)
				}
			}
		})

		t.Run(fmt.Sprintf("%s/json", tc.name), func(t *testing.T) {
			h := NewTestCommandHarness(t, fn)

			args := make([]string, len(tc.args))
			copy(args, tc.args)
			args = append(args, "--format=json")

			rc := h.Execute(context.Background(), args)
			if tc.expectError {
				require.NotZero(t, rc)
				require.Contains(t, h.Stderr(), "--log-level")
				return
			}

			require.Zero(t, rc, "expected no error, got:\n%s", h.Stderr())
			require.Equal(t, tc.expectedLevel, h.cmd.logLevel)

			for l := types.ErrorLevel; l <= types.TraceLevel; l++ {
				msg := fmt.Sprintf(`"message":"%s"`, l)
				if tc.expectedLevel >= l {
					require.Contains(t, h.Stderr(), msg)
				} else {
					require.NotContains(t, h.Stderr(), msg)
				}
			}
		})
	}
}
