// +build !js

package vutils

import (
	"os"
)

type processUtils struct {
}

func (pu *processUtils) NewProcessManager(logStdOut bool, captureSignal bool, exitOnSignal bool) *ProcessManager {

	pm := &ProcessManager{
		LogStdOut:     logStdOut,
		CaptureSignal: captureSignal,
		ExitOnSignal:  exitOnSignal,
		processMap:    map[int]*ProcessManagerProcess{},
	}

	pm.init()

	return pm

}

func (pu *processUtils) DefaultProcessOptions() (*ProcessManagerProcessOptions, error) {

	cwd, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	return &ProcessManagerProcessOptions{
		OutputStdErr: true,
		OutputStdOut: false,
		CWD:          cwd,
	}, nil

}

func (pu *processUtils) NewProcessOptions() *ProcessManagerProcessOptions {

	return &ProcessManagerProcessOptions{}

}

var Processes = &processUtils{}
