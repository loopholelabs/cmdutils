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

func Test_Format_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "text", FormatText.String())
	assert.Equal(t, "json", FormatJSON.String())
}

func Test_Format_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expected    Format
		expectedErr string
	}{
		{
			name:     "text",
			input:    "text",
			expected: FormatText,
		},
		{
			name:     "text all caps",
			input:    "TEXT",
			expected: FormatText,
		},
		{
			name:     "text mixed caps",
			input:    "TeXt",
			expected: FormatText,
		},
		{
			name:     "json",
			input:    "json",
			expected: FormatJSON,
		},
		{
			name:     "json all caps",
			input:    "JSON",
			expected: FormatJSON,
		},
		{
			name:     "json mixed caps",
			input:    "jSoN",
			expected: FormatJSON,
		},
		{
			name:        "invalid input",
			input:       "not valid",
			expectedErr: "unknown log format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var f Format
			err := f.Set(tc.input)

			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, f)
			}

		})
	}
}

func Test_Format_Unmarshal(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("format", "json")

	c := struct{ Format Format }{}
	v.Unmarshal(&c, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	require.Equal(t, FormatJSON, c.Format)
}
