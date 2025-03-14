// Copyright 2023, Pulumi Corporation.
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

package esc

import (
	"context"

	"github.com/pulumi/esc/schema"
)

// A Provider provides environments access to dynamic secrets. These secrets may be generated at runtime, fetched from
// other services, etc.
type Provider interface {
	// Schema returns the provider's input and output schemata.
	Schema() (inputs, outputs *schema.Schema)

	// Open retrieves the provider's secrets.
	Open(ctx context.Context, inputs map[string]Value, executionContext EnvExecContext) (Value, error)
}

// A Rotator enables environments to rotate a secret.
// It is the responsibility of the caller to appropriately persist rotation state (e.g. by writing it back to the environment definition).
type Rotator interface {
	// Schema returns the rotator's input, state, and output schemata.
	Schema() (inputs, state, outputs *schema.Schema)

	// Open retrieves the rotator's secrets, using persisted state.
	Open(ctx context.Context, inputs, state map[string]Value, executionContext EnvExecContext) (Value, error)

	// Rotate rotates the provider's secret, and returns the rotator's new state to be persisted.
	Rotate(ctx context.Context, inputs, state map[string]Value, executionContext EnvExecContext) (Value, error)
}
