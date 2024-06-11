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
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// New returns a logger with the given configuration.
func New(logFile string, level Level, format Format) *slog.Logger {
	var writer io.Writer

	switch logFile {
	case "stdout":
		writer = os.Stdout
	case "", "stderr":
		writer = os.Stderr
	default:
		writer = &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    128,
			MaxAge:     7,
			MaxBackups: 4,
		}
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch format {
	case FormatText:
		handler = slog.NewTextHandler(writer, opts)
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, opts)
	}

	return slog.New(handler)
}
