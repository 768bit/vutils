// +build !js

package vutils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type execUtils struct {
}

type ExecAsyncCommand struct {
	path           string
	args           []string
	env            []string
	Proc           *exec.Cmd
	errOnly        bool
	dir            string
	reader         io.ReadCloser
	error          io.ReadCloser
	writer         io.WriteCloser
	stdoutBuffer   bytes.Buffer
	stdoutWriter   *bufio.Writer
	stderrWriter   *bufio.Writer
	stderrBuffer   bytes.Buffer
	intChan        chan os.Signal
	intBound       bool
	stdioBound     bool
	stdioCapture   bool
	combineCapture bool
	useSudo        bool
}

func (ec *ExecAsyncCommand) init() *ExecAsyncCommand {
	if ec.env == nil {
		ec.env = []string{}
	}
	if ec.useSudo {
		cargs := append([]string{"-n", ec.path}, ec.args...)
		ec.Proc = exec.Command("sudo", cargs...)
	} else {
		ec.Proc = exec.Command(ec.path, ec.args...)
	}
	if !ec.errOnly {
		ec.reader, _ = ec.Proc.StdoutPipe()
	}

	ec.error, _ = ec.Proc.StderrPipe()
	ec.writer, _ = ec.Proc.StdinPipe()

	return ec
}

func (ec *ExecAsyncCommand) BindToStdoutAndStdErr() *ExecAsyncCommand {
	if ec.stdioBound {
		return ec
	} else if ec.stdioCapture {
		log.Println("Unable to Bind STDIO as STDIO is already being captured.")
		return ec
	}
	if !ec.errOnly {
		go func() {
			io.Copy(os.Stdout, ec.reader)
		}()
	}

	go func() {
		io.Copy(os.Stderr, ec.error)
	}()
	ec.stdioBound = true
	return ec
}

func (ec *ExecAsyncCommand) CaptureStdoutAndStdErr(combine bool, outputToStdIO bool) *ExecAsyncCommand {
	if ec.stdioCapture {
		return ec
	} else if ec.stdioBound {
		log.Println("Unable to Capture STDIO as STDIO is already bound.")
		return ec
	}
	if !ec.errOnly {
		go func() {
			if outputToStdIO {
				//outScanner := bufio.NewScanner(ec.reader)
				ec.stdoutWriter = bufio.NewWriter(&ec.stdoutBuffer)

				tee := io.TeeReader(ec.reader, os.Stdout)
				//_ = io.TeeReader(tee, os.Stdout)
				io.Copy(ec.stdoutWriter, tee)

				//if combine {
				//	ec.stdoutWriter = stdoutWriter
				//}
				//for outScanner.Scan() {
				//	txt := outScanner.Text()
				//	println(txt)
				//	ec.stdoutWriter.WriteString(txt + "\n")
				//}
			} else {
				ec.stdoutWriter = bufio.NewWriter(&ec.stdoutBuffer)
				io.Copy(ec.stdoutWriter, ec.reader)
			}
		}()
	}

	go func() {
		if outputToStdIO || (combine && !ec.errOnly) {
			//outScanner := bufio.NewScanner(ec.error)
			ec.stderrWriter = bufio.NewWriter(&ec.stderrBuffer)

			tee := io.TeeReader(ec.error, os.Stderr)
			//_ = io.TeeReader(tee, os.Stderr)
			if combine && !ec.errOnly {
				outTee := io.TeeReader(tee, ec.stdoutWriter)
				io.Copy(ec.stderrWriter, outTee)
			} else {
				io.Copy(ec.stderrWriter, tee)
			}

			//for outScanner.Scan() {
			//	txt := outScanner.Text()
			//	println(txt)
			//	ec.stderrWriter.WriteString(txt + "\n")
			//	if combine && !ec.errOnly {
			//		ec.stdoutWriter.WriteString(txt + "\n")
			//	}
			//}
		} else {
			ec.stderrWriter = bufio.NewWriter(&ec.stderrBuffer)
			io.Copy(ec.stderrWriter, ec.error)
		}
	}()
	ec.stdioCapture = true
	return ec
}

func (ec *ExecAsyncCommand) GetStdoutBuffer() []byte {
	ec.stdoutWriter.Flush()
	return ec.stdoutBuffer.Bytes()
}

func (ec *ExecAsyncCommand) GetStderrBuffer() []byte {
	ec.stderrWriter.Flush()
	return ec.stderrBuffer.Bytes()
}

func (ec *ExecAsyncCommand) SetWorkingDir(path string) *ExecAsyncCommand {
	ec.dir = path
	return ec
}

func (ec *ExecAsyncCommand) CopyEnv() *ExecAsyncCommand {
	ec.env = os.Environ()
	return ec
}

func (ec *ExecAsyncCommand) AddEnv(key string, value string) *ExecAsyncCommand {
	ec.env = append(ec.env, fmt.Sprintf(`%s=%s`, key, value))
	return ec
}

func (ec *ExecAsyncCommand) SetEnv(env []string) *ExecAsyncCommand {
	ec.env = env
	return ec
}

func (ec *ExecAsyncCommand) Sudo() *ExecAsyncCommand {
	if !ec.useSudo {
		if Exec.CheckSudo() {
			return ec
		}
		ec.useSudo = true
		return ec.init()
	}

	return ec
}

func (ec *ExecAsyncCommand) Start() error {
	//set what needs to be set..

	if ec.env != nil && len(ec.env) > 0 {
		ec.Proc.Env = ec.env
	}
	if ec.dir != "" {
		ec.Proc.Dir = ec.dir
	}

	//if !ec.errOnly {
	//	defer ec.reader.Close()
	//}
	//if ec.intChan != nil && ec.intBound {
	//	defer close(ec.intChan)
	//}
	//defer ec.writer.Close()
	//defer ec.error.Close()
	fmt.Printf("$: %s %s\n", ec.Proc.Path, strings.Join(ec.Proc.Args, ` `))
	if err := ec.Proc.Start(); err != nil {
		return err
	}
	return nil
}

func (ec *ExecAsyncCommand) StartAndWait() error {
	if ec.env != nil && len(ec.env) > 0 {
		ec.Proc.Env = ec.env
	}
	if ec.dir != "" {
		ec.Proc.Dir = ec.dir
	}

	//if ec.stdioCapture && ec.stderrWriter != nil {
	//	if !ec.errOnly && ec.stdoutWriter != nil {
	//		defer ec.stdoutWriter.Flush()
	//	}
	//	defer ec.stderrWriter.Flush()
	//}

	fmt.Printf("$: %s %s - IN: %s\n", ec.Proc.Path, strings.Join(ec.Proc.Args, ` `), ec.Proc.Dir)
	if err := ec.Proc.Start(); err != nil {
		return err
	} else if err := ec.Wait(); err != nil {
		return err
	}
	return nil
}

func (ec *ExecAsyncCommand) Wait() error {

	if !ec.errOnly {
		defer ec.reader.Close()
	}
	if ec.intChan != nil && ec.intBound {
		defer close(ec.intChan)
	}
	defer ec.writer.Close()
	defer ec.error.Close()

	if err := ec.Proc.Wait(); err != nil {
		return err
	}

	return nil
}

func (ec *ExecAsyncCommand) BindSigIntHandler() *ExecAsyncCommand {
	ec.intBound = true
	ec.intChan = make(chan os.Signal)
	signal.Notify(ec.intChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ec.intChan
		ec.Proc.Process.Signal(syscall.SIGTERM)
	}()
	return ec
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
		path:     path,
		args:     args,
		errOnly:  errOnly,
		intBound: false,
	}

	pr.init()

	return &pr

}

func (ex *execUtils) CheckSudo() bool {

	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	// output has trailing \n
	// need to remove the \n
	// otherwise it will cause error for strconv.Atoi
	// log.Println(output[:len(output)-1])

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if i != 0 || err != nil {
		return false
	} else {
		return true
	}

}

var Exec = &execUtils{}
