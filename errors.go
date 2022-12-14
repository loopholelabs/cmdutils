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

package cmdutils

const ActionRequestedExitCode = 1
const FatalErrExitCode = 2

// Error can be used by a command to change the exit status of the CLI.
type Error struct {
	Msg string
	// Status
	ExitCode int
}

func (e *Error) Error() string { return e.Msg }
