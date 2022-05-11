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

//go:build linux

package svcrunner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
)

// Ensure Run implements the correct public interface.
var _ runFn = Run

func Run(service S) error {
	if err := service.Init(); err != nil {
		return fmt.Errorf("svchost/init: %w", err)
	}

	daemon.SdNotify(false, daemon.SdNotifyReady)

	if err := service.Start(); err != nil {
		return fmt.Errorf("svchost/start: %w", err)
	}

	var ctx context.Context
	ctx, cancelFn = context.WithCancel(context.Background())
	defer cancelFn()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination request.
	select {
	case <-ctx.Done():
	case <-sigCh:
	}

	daemon.SdNotify(false, daemon.SdNotifyStopping)

	return service.Stop()
}
