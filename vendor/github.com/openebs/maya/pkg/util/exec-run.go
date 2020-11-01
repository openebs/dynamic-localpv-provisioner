// Copyright © 2018-2019 The OpenEBS Authors
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

package util

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"context"

	"k8s.io/klog"
)

// Runner interface implements various methods of running binaries which can be
// modified for unit testing.
type Runner interface {
	RunCombinedOutput(string, ...string) ([]byte, error)
	RunStdoutPipe(string, ...string) ([]byte, error)
	RunCommandWithTimeoutContext(time.Duration, string, ...string) ([]byte, error)
	RunCommandWithLog(string, ...string) ([]byte, error)
}

// RealRunner is the real runner for the program that actually execs the command.
type RealRunner struct{}

// RunCombinedOutput runs the command and returns its combined standard output
// and standard error.
func (r RealRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	//execute pool creation command.
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

// RunStdoutPipe returns a pipe that will be connected to the command's standard output
// when the command starts.
func (r RealRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	data, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		klog.Errorf(err.Error())
		return []byte{}, err
	}
	return data, nil
}

// RunCommandWithLog triggers command passed as arguments and it also does the
// following things before command completion
// 1. Logs the stdout of command to stdout(standard output)
// 2. Logs stderr of the command to standard error
func (r RealRunner) RunCommandWithLog(command string, args ...string) ([]byte, error) {
	// #nosec
	cmd := exec.Command(command, args...)
	// Redirect the command output to stdout
	cmd.Stdout = os.Stdout
	// Redirect the command output to stderr
	cmd.Stderr = os.Stderr
	// Start the command
	if err := cmd.Start(); err != nil {
		return []byte{}, err
	}
	// below will return error when command exit with return code 1
	if err := cmd.Wait(); err != nil {
		return []byte{}, err
	}
	return []byte{}, nil
}

// RunCommandWithTimeoutContext executes command provides and returns stdout
// error. If command does not returns within given timout interval command will
// be killed and return "Context time exceeded"
func (r RealRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, errors.Wrapf(ctx.Err(), "Failed to run command: %v %v", command, args)
		default:
			return nil, err
		}
	}
	return out, nil
}

//TestRunner is used as a dummy Runner
type TestRunner struct{}

// RunCombinedOutput is to mock Real runner exec.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunStdoutPipe is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunCommandWithTimeoutContext is to mock Real runner exec.
func (r TestRunner) RunCommandWithTimeoutContext(timeout time.Duration, command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunCommandWithLog is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunCommandWithLog(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}
