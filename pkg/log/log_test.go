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

package log

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		level        Level
		format       Format
		expectedLogs []string
	}{
		{
			name:   "text log warn",
			level:  LevelWarn,
			format: FormatText,
			expectedLogs: []string{
				"msg=warn",
				"msg=error",
			},
		},
		{
			name:   "text log debug",
			level:  LevelDebug,
			format: FormatText,
			expectedLogs: []string{
				"msg=debug",
				"msg=info",
				"msg=warn",
				"msg=error",
			},
		},
		{
			name:   "json log warn",
			level:  LevelWarn,
			format: FormatJSON,
			expectedLogs: []string{
				`"msg":"warn"`,
				`"msg":"error"`,
			},
		},
		{
			name:   "json log debug",
			level:  LevelDebug,
			format: FormatJSON,
			expectedLogs: []string{
				`"msg":"debug"`,
				`"msg":"info"`,
				`"msg":"warn"`,
				`"msg":"error"`,
			},
		},
	}

	logDir := t.TempDir()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize logger.
			logFile := path.Join(logDir, t.Name())
			logger := New(logFile, tc.level, tc.format)

			// Verify logger configuration.
			handler := logger.Handler()
			require.True(t, handler.Enabled(nil, slog.Level(tc.level)))

			switch tc.format {
			case FormatText:
				require.IsType(t, &slog.TextHandler{}, handler)
			case FormatJSON:
				require.IsType(t, &slog.JSONHandler{}, handler)
			}

			// Emit test log messages.
			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			// Verify log file contents.
			buf, err := os.ReadFile(logFile)
			require.NoError(t, err)

			logs := strings.Split(string(buf), "\n")
			require.Len(t, logs[:len(logs)-1], len(tc.expectedLogs))

			for _, s := range tc.expectedLogs {
				assert.Contains(t, string(buf), s)
			}
		})
	}
}

// TestLogger_StdoutStderr modifies the global os.Stdout and os.Stderr values
// so it SHOULD NOT run with t.Parallel().
func TestLogger_StdoutStderr(t *testing.T) {
	// Hijack stdout and stderr.
	stdoutOrig := os.Stdout
	stderrOrig := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	t.Cleanup(func() {
		os.Stdout = stdoutOrig
		os.Stderr = stderrOrig
	})

	// Log to stdout.
	logger := New("stdout", LevelInfo, FormatText)
	logger.Info("test1")

	// Log to stderr.
	logger = New("", LevelInfo, FormatText)
	logger.Info("test2")

	logger = New("stderr", LevelInfo, FormatText)
	logger.Info("test3")

	// Read stdout and stderr.
	stdoutW.Close()
	stderrW.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// Verify log output.
	assert.Contains(t, stdout, "msg=test1")
	assert.Contains(t, stderr, "msg=test2")
	assert.Contains(t, stderr, "msg=test3")
}
