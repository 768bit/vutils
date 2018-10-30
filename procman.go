package vutils

import (
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type ProcessManager struct {
	processMap    map[int]*ProcessManagerProcess
	LogStdOut     bool
	CaptureSignal bool
	ExitOnSignal  bool
	OnExit        func()
	exitChan      chan os.Signal
	cleaningUp    bool
}

func (pm *ProcessManager) init() {

	if pm.CaptureSignal {

		//if we are forcing a clean exit we need to signal all child processes and do the cleanup

		pm.exitChan = make(chan os.Signal)
		signal.Notify(pm.exitChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
		go func() {
			<-pm.exitChan
			if !pm.CaptureSignal {
				return
			}
			log.Print("Killing all processes")
			err := pm.signalTermAllProcesses()
			if err != nil {
				log.Print("Error killing all processes")
				log.Print(err)
			}

			if pm.OnExit != nil {

				pm.OnExit()

			}

			if pm.ExitOnSignal {
				os.Exit(0)
			}
		}()

	}

}

func (pm *ProcessManager) signalTermAllProcesses() error {

	g, _ := errgroup.WithContext(context.Background())

	for pid, proc := range pm.processMap {

		g.Go(func() error {

			if err := proc.Signal(syscall.SIGINT); err != nil {
				//return nil as we want all items to complete killing
				log.Printf("Error sending signal to process %d", pid)
			} else if err := proc.Wait(); err != nil {
				log.Printf("Error waiting for process %d to exit on cleanup.", pid)
			} else {
				log.Printf("Successfully killed %d", pid)
			}

			return nil

		})

	}

	return g.Wait()

}

func (pm *ProcessManager) addProcessToMap(proc *ProcessManagerProcess) error {

	if _, ok := pm.processMap[proc.execProc.Proc.Process.Pid]; ok {

		return errors.New("Unable to add process to process manager as it already exists")

	}

	pm.processMap[proc.execProc.Proc.Process.Pid] = proc

	go func() {
		err := proc.Wait()
		if err != nil {
			log.Print("Error Waiting for Process to finish")
			log.Print(err)
		}
		if !pm.cleaningUp {
			pm.removeProcessFromMap(proc)
		}
	}()

	return nil

}

func (pm *ProcessManager) removeProcessFromMap(proc *ProcessManagerProcess) error {

	if _, ok := pm.processMap[proc.execProc.Proc.Process.Pid]; !ok {

		return errors.New("Unable to remove process from process manager as it doesnt exist")

	}

	delete(pm.processMap, proc.execProc.Proc.Process.Pid)

	return nil

}

func (pm *ProcessManager) SignIntAll() error {

	return pm.signalTermAllProcesses()

}

func (pm *ProcessManager) Shell(cmd string, options *ProcessManagerProcessOptions) (*ProcessManagerProcess, error) {

	if proc, err := pm.RunAsync("/bin/bash", options, "-c", cmd); err != nil {
		return nil, err
	} else {
		return proc, nil
	}

}
func (pm *ProcessManager) ShellWait(cmd string, options *ProcessManagerProcessOptions) error {

	return pm.RunAsyncWait("/bin/bash", options, "-c", cmd)

}

func (pm *ProcessManager) ExecScript(scriptPath string, options *ProcessManagerProcessOptions, cmdArgs ...string) (*ProcessManagerProcess, error) {

	nargs := make([]string, 1+len(cmdArgs))

	nargs = append(nargs, cmdArgs...)

	nargs[0] = scriptPath

	if proc, err := pm.RunAsync("/bin/bash", options, nargs...); err != nil {
		return nil, err
	} else {
		return proc, nil
	}

}
func (pm *ProcessManager) ExecScriptWait(scriptPath string, options *ProcessManagerProcessOptions, cmdArgs ...string) error {

	nargs := make([]string, 1+len(cmdArgs))

	nargs[0] = scriptPath

	nargs = append(nargs, cmdArgs...)

	return pm.RunAsyncWait("/bin/bash", options, nargs...)

}

func (pm *ProcessManager) RunAsync(binary string, options *ProcessManagerProcessOptions, cmdArgs ...string) (*ProcessManagerProcess, error) {

	if options == nil {

		nopts, err := Processes.DefaultProcessOptions()

		if err != nil {
			return nil, err
		}

		options = nopts

	}

	if pm.LogStdOut {
		options.OutputStdOut = true
	}

	proc := newProcManProcess(pm, binary, options, cmdArgs...)

	if err := proc.Start(); err != nil {
		return nil, err
	} else if err := pm.addProcessToMap(proc); err == nil {
		return proc, nil
	} else {
		return nil, err
	}

}
func (pm *ProcessManager) RunAsyncWait(binary string, options *ProcessManagerProcessOptions, cmdArgs ...string) error {

	if proc, err := pm.RunAsync(binary, options, cmdArgs...); err == nil {
		return proc.Wait()
	} else {
		return err
	}

}
func (pm *ProcessManager) RunExec(pr *ExecAsyncCommand) (*ProcessManagerProcess, error) {

	options, err := Processes.DefaultProcessOptions()

	if err != nil {
		return nil, err
	}

	if !pr.errOnly {
		options.OutputStdOut = true
	}

	proc := newProcManProcessFromExec(pm, options, pr)

	if err := proc.Start(); err != nil {
		return nil, err
	} else if err := pm.addProcessToMap(proc); err == nil {
		return proc, nil
	} else {
		return nil, err
	}

}
func (pm *ProcessManager) RunExecWait(pr *ExecAsyncCommand) error {

	if proc, err := pm.RunExec(pr); err == nil {
		return proc.Wait()
	} else {
		return err
	}

}
