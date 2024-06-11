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
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevel_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "DEBUG", LevelDebug.String())
	assert.Equal(t, "INFO", LevelInfo.String())
	assert.Equal(t, "WARN", LevelWarn.String())
	assert.Equal(t, "ERROR", LevelError.String())
}

func TestLevel_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expected    Level
		expectedErr string
	}{
		{
			name:     "debug level",
			input:    "debug",
			expected: LevelDebug,
		},
		{
			name:     "info level all caps",
			input:    "INFO",
			expected: LevelInfo,
		},
		{
			name:     "warn level mixed caps",
			input:    "WarN",
			expected: LevelWarn,
		},
		{
			name:     "error level",
			input:    "error",
			expected: LevelError,
		},
		{
			name:        "invalid input",
			input:       "not valid",
			expectedErr: "unknown name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var l Level
			err := l.Set(tc.input)

			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, l)
			}

		})
	}
}

func TestLevel_Unmarshal(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("level", "debug")

	c := struct{ Level Level }{}
	v.Unmarshal(&c, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	require.Equal(t, LevelDebug, c.Level)
}
