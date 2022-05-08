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

//go:build windows

package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
)

var cancelFn context.CancelFunc

type windowsService struct {
	service   S
	isService bool
	ctx       context.Context
}

func (ws *windowsService) runAsService() error {
	return svc.Run(ws.service.GetName(), ws)
}

func (ws *windowsService) runInteractive() error {
	if err := ws.service.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	select {
	case <-ws.ctx.Done():
	case <-ch:
	}

	return ws.service.Stop()
}

// Execute is invoked by Windows.
func (ws *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const accepts = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	if err := ws.service.Start(); err != nil {
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: accepts}

	for {
		var c svc.ChangeRequest
		select {
		case <-ws.ctx.Done():
			c = svc.ChangeRequest{Cmd: svc.Stop}
		case c = <-r:
		}

		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			if err := ws.service.Stop(); err != nil {
				return true, 2
			}
			return false, 0
		default:
		}
	}
}

// Run runs an implementation of the Service interface.
//
// Run will block until the Windows Service is stopped or Ctrl+C is
// pressed if running interactively from the console.
func Run(service S) error {
	var err error

	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("win/IsWindowsService: %w", err)
	}

	ws := &windowsService{
		service:   service,
		isService: isService,
	}

	ws.ctx, cancelFn = context.WithCancel(context.Background())
	defer cancelFn()

	if isService {
		dir := filepath.Dir(os.Args[0])
		if err := os.Chdir(dir); err != nil {
			return fmt.Errorf("service/chdir: %w", err)
		}
	}

	if err := service.Init(); err != nil {
		return fmt.Errorf("service/init: %w", err)
	}

	if isService {
		return ws.runAsService()
	} else {
		return ws.runInteractive()
	}
}

// Stop will request the service to shut down.
func Stop() {
	cancelFn()
}
