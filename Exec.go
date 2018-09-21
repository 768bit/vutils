package vutils

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type execUtils struct {
}

type ExecAsyncCommand struct {
	Proc    *exec.Cmd
	errOnly bool
	reader  io.ReadCloser
	error   io.ReadCloser
	writer  io.WriteCloser
}

func (ec *ExecAsyncCommand) BindToStdoutAndStdErr() {
	if !ec.errOnly {
		scanner := bufio.NewScanner(ec.reader)
		go func() {
			for scanner.Scan() {
				log.Print(scanner.Text())
			}
		}()
	}

	errScanner := bufio.NewScanner(ec.error)
	go func() {
		for errScanner.Scan() {
			log.Print(errScanner.Text())
		}
	}()
}

func (ex *execUtils) CreateExecCommand(path string, args ...string) *exec.Cmd {

	cmd := exec.Command(path, args...)

	return cmd

}

func (ex *execUtils) ExecCommandShowStdErr(path string, args ...string) (*exec.Cmd, error) {

	cmd := ex.CreateExecCommand(path, args...)

	var out bytes.Buffer
	cmd.Stderr = &out

	err := cmd.Run()

	if err != nil {

		log.Print(out.String())

		return cmd, err

	}

	return cmd, nil

}

func (ex *execUtils) ExecCommandShowStdErrReturnOutput(path string, args ...string) (string, error) {

	cmd := exec.Command(path, args...)

	var out bytes.Buffer
	cmd.Stderr = &out

	stdout, err := cmd.Output()

	if err != nil {

		log.Print(out.String())

		return "", err

	}

	return string(stdout), nil

}

func (ex *execUtils) RunCommandShowStdErr(path string, args ...string) error {

	cmd := exec.Command(path, args...)

	var out bytes.Buffer
	cmd.Stderr = &out

	err := cmd.Run()

	if err != nil {

		log.Print(out.String())

		return err

	}

	return nil

}

func (ex *execUtils) RunCommandAsyncOutput(path string, errOnly bool, args ...string) error {

	pr := ex.CreateAsyncCommand(path, errOnly, args...)

	if !errOnly {
		scanner := bufio.NewScanner(pr.reader)
		go func() {
			for scanner.Scan() {
				log.Print(scanner.Text())
			}
		}()
	}

	errScanner := bufio.NewScanner(pr.error)
	go func() {
		for errScanner.Scan() {
			log.Print(errScanner.Text())
		}
	}()

	if !errOnly {
		defer pr.reader.Close()
	}

	defer pr.writer.Close()
	defer pr.error.Close()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		pr.Proc.Process.Signal(syscall.SIGTERM)
	}()

	defer close(c)

	if err := pr.Proc.Start(); err != nil {
		return err
	} else if err := pr.Proc.Wait(); err != nil {
		return err
	}

	return nil

}

func (ex *execUtils) CreateAsyncCommand(path string, errOnly bool, args ...string) *ExecAsyncCommand {

	pr := ExecAsyncCommand{
		Proc:    exec.Command(path, args...),
		errOnly: errOnly,
	}

	if !errOnly {
		pr.reader, _ = pr.Proc.StdoutPipe()
	}

	pr.error, _ = pr.Proc.StderrPipe()
	pr.writer, _ = pr.Proc.StdinPipe()

	return &pr

}

var Exec = &execUtils{}
