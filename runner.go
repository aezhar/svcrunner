// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2022 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// <https://www.apache.org/licenses/LICENSE-2.0>
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package svcrunner

import (
	"context"
)

type S interface {
	// GetName returns a short descriptive name for the service.
	GetName() string

	// Init is called before the service is started.
	Init() error

	// Start is called after Init.
	Start() error

	// Stop is called in response to a request to stop the service.
	Stop() error
}

// runFn is platform specific.
type runFn func(service S) error

var cancelFn context.CancelFunc

// Stop will request the service to shut down.
func Stop() {
	if cancelFn != nil {
		cancelFn()
	}
}
