// Copyright 2024, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

type DebugContext interface {
	// StartDebugging asks the host to start a debug session for the given configuration.
	StartDebugging(info DebuggingInfo) error

	// AttachDebugger returns true if debugging is enabled.
	AttachDebugger() bool
}

type DebuggingInfo struct {
	// Config is the debug configuration (language-specific, see Debug Adapter Protocol)
	Config map[string]interface{}
}
