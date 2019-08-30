// +build !js

package vutils

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type ProcessManagerProcess struct {
	execProc *ExecAsyncCommand
	pm       *ProcessManager
	options  *ProcessManagerProcessOptions
	waitChan chan bool
	waitErr  error
}

type ProcessManagerProcessOptions struct {
	CWD             string
	ENV             map[string]string
	OutputStdOut    bool
	ParseStdOutLine func(string)
	ParseStdErrLine func(string)
	OutputStdErr    bool
	OnError         func(err error)
	OnExit          func()
}

func newProcManProcess(procMan *ProcessManager, binary string, options *ProcessManagerProcessOptions, cmdArgs ...string) *ProcessManagerProcess {

	eproc := Exec.CreateAsyncCommand(binary, !options.OutputStdOut, cmdArgs...)

	if options.CWD != "" {

		eproc.Proc.Dir = options.CWD

	}

	if options.ENV != nil && len(options.ENV) > 0 {

		env := os.Environ()

		for key, val := range options.ENV {

			env = append(env, fmt.Sprintf("%s=\"%s\"", key, val))

		}

		eproc.Proc.Env = env

	}

	if options.OutputStdErr {
		errScanner := bufio.NewScanner(eproc.error)
		go func() {
			for errScanner.Scan() {
				txt := errScanner.Text()
				if options.ParseStdErrLine != nil {
					options.ParseStdErrLine(txt)
				}
				log.Print(txt)
			}
		}()
	}

	if options.OutputStdOut {
		scanner := bufio.NewScanner(eproc.reader)
		go func() {
			for scanner.Scan() {
				txt := scanner.Text()
				if options.ParseStdOutLine != nil {
					options.ParseStdOutLine(txt)
				}
				log.Print(txt)
			}
		}()
	}

	pmp := ProcessManagerProcess{
		pm:       procMan,
		execProc: eproc,
		options:  options,
		waitChan: make(chan bool),
	}

	return &pmp

}

func newProcManProcessFromExec(procMan *ProcessManager, options *ProcessManagerProcessOptions, eproc *ExecAsyncCommand) *ProcessManagerProcess {

	if options.CWD != "" {

		eproc.SetWorkingDir(options.CWD)

	}

	if options.ENV != nil && len(options.ENV) > 0 {

		env := os.Environ()

		for key, val := range options.ENV {

			env = append(env, fmt.Sprintf("%s=\"%s\"", key, val))

		}

		eproc.SetEnv(env)

	}

	if options.OutputStdErr {
		errScanner := bufio.NewScanner(eproc.error)
		go func() {
			for errScanner.Scan() {
				txt := errScanner.Text()
				if options.ParseStdErrLine != nil {
					options.ParseStdErrLine(txt)
				}
				fmt.Println(txt)
			}
		}()
	}

	if options.OutputStdOut {
		scanner := bufio.NewScanner(eproc.reader)
		go func() {
			for scanner.Scan() {
				txt := scanner.Text()
				if options.ParseStdOutLine != nil {
					options.ParseStdOutLine(txt)
				}
				fmt.Println(txt)
			}
		}()
	}

	pmp := ProcessManagerProcess{
		pm:       procMan,
		execProc: eproc,
		options:  options,
		waitChan: make(chan bool),
	}

	return &pmp

}

func (pmp *ProcessManagerProcess) Signal(signal os.Signal) error {

	log.Printf("Signalling process with %d", signal)

	pmp.waitChan <- true

	return pmp.execProc.Proc.Process.Signal(signal)

}

func (pmp *ProcessManagerProcess) Wait() error {

	waitRes := <-pmp.waitChan
	if !waitRes {
		return pmp.waitErr
	}

	return nil

}

func (pmp *ProcessManagerProcess) Start() error {

	err := pmp.execProc.Proc.Start()

	if err != nil {
		return err
	}

	go func() {

		err := pmp.execProc.Proc.Wait()
		fmt.Println("Ending Run Proc")
		defer close(pmp.waitChan)
		if err != nil {
			pmp.waitErr = err
			pmp.waitChan <- false
			return
		}
		pmp.waitChan <- true
		pmp.pm.removeProcessFromMap(pmp)
		if pmp.options.OnExit != nil {
			pmp.options.OnExit()
		}

	}()

	return nil

}
