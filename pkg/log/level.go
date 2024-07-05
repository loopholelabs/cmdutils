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

import "log/slog"

// Level represents a log level threshold.
type Level slog.Level

const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// String returns the string representation of a log level.
//
// Implements the fmt.Stringer and pflag.Value interfaces.
func (l Level) String() string {
	return slog.Level(l).String()
}

// Type returns the string type of a log level.
//
// Implements the pflag.Value interface.
func (l Level) Type() string {
	return "string"
}

// Level returns the slog.Level value of a Level.
//
// Implements the slog.Leveler interface.
func (l Level) Level() slog.Level {
	return slog.Level(l)
}

// Set updates the value of a Level variable.
//
// Implements the pflag.Value interface.
func (l *Level) Set(s string) error {
	return (*slog.Level)(l).UnmarshalText([]byte(s))
}

// UnmarshalText decodes the text representation of a log level.
//
// Implements the encoding.TextUnmarshaler interface used by mapstructure and
// viper.
func (l *Level) UnmarshalText(text []byte) error {
	return l.Set(string(text))
}
