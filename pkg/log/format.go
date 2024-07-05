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
	"fmt"
	"strings"
)

// Format represents a log output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

// String returns the string representation of a log format.
//
// Implements the fmt.Stringer and pflag.Value interfaces.
func (f Format) String() string {
	switch f {
	case FormatText:
		return "text"
	case FormatJSON:
		return "json"
	default:
		return "<undefined>"
	}
}

// Type returns the string type of a log format.
//
// Implements the pflag.Value interface.
func (f Format) Type() string {
	return "string"
}

// Set updates the value of a Format variable.
//
// Implements the pflag.Value interface.
func (f *Format) Set(s string) error {
	switch strings.ToLower(s) {
	case "text":
		*f = FormatText
	case "json":
		*f = FormatJSON
	default:
		return fmt.Errorf("unknown log format %q", s)
	}
	return nil
}

// UnmarshalText decodes the text representation of a log format.
//
// Implements the encoding.TextUnmarshaler interface used by mapstructure and
// viper.
func (f *Format) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}
