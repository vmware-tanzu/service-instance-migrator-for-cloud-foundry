/* 
 *  Copyright 2022 VMware, Inc.
 *  
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *  http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	sio "github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

type ShellScriptExecutor struct {
	debug   bool
	dryRun  bool
	timeout time.Duration
	result  Result
}

type Result struct {
	Output string
	Status *Status
	DryRun bool
}

type Option func(*ShellScriptExecutor)

func NewExecutor(options ...Option) *ShellScriptExecutor {
	e := &ShellScriptExecutor{}

	for _, o := range options {
		o(e)
	}

	return e
}

func WithDryRun(dryRun bool) Option {
	return func(e *ShellScriptExecutor) {
		e.dryRun = dryRun
	}
}

func WithDebug(debug bool) Option {
	return func(e *ShellScriptExecutor) {
		e.debug = debug
	}
}

func WithTimeout(t time.Duration) Option {
	if t < 0 {
		log.Panic("timeout cannot be less 0")
	}

	return func(e *ShellScriptExecutor) {
		e.timeout = t
	}
}

func (e *ShellScriptExecutor) Execute(ctx context.Context, src io.Reader) (Result, error) {
	script, err := e.copyToTempFile(src)
	if err != nil {
		return Result{}, err
	}
	defer func() {
		err := os.Remove(script.Name())
		if err != nil {
			// just log err
			log.Errorf("error %s, failed to remove file: %q after executing", err, script.Name())
		}
	}()

	input, err := e.saveInput(script)
	if err != nil {
		return Result{}, err
	}

	if e.dryRun {
		err := e.printInput(os.Stdout, input)
		return Result{DryRun: true}, err
	}

	if e.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "/bin/bash", script.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Warnln("command context was canceled")
		}
		return Result{
			Output: out.String(),
			Status: newStatus(cmd, script.Name(), input, out.String(), err),
		}, StatusError(err)
	}

	e.saveOutput(cmd, script.Name(), input, out.String(), err)

	return e.result, nil
}

func (e *ShellScriptExecutor) LastResult() Result {
	return e.result
}

func (e *ShellScriptExecutor) printInput(writer io.Writer, input string) error {
	_, err := fmt.Fprintln(writer, input)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to print %q", input))
	}
	return nil
}

func (e *ShellScriptExecutor) copyToTempFile(r io.Reader) (*os.File, error) {
	script, err := sio.CopyToTempFile(r)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to save script to temp file, %v", r))
	}

	return script, err
}

func (e *ShellScriptExecutor) saveInput(script *os.File) (string, error) {
	data, err := os.ReadFile(script.Name())
	if string(data) == "" {
		return "", errors.Wrap(err, fmt.Sprintf("no input data to save, script file %q is empty", script.Name()))
	}
	var b bytes.Buffer
	b.WriteString(string(data))

	if e.debug {
		log.Debugf("Command input: %s", b.String())
	}

	return b.String(), nil
}

func (e *ShellScriptExecutor) saveOutput(cmd *exec.Cmd, script string, input string, output string, err error) {
	if e.debug {
		log.Debugf("Command output: %s", output)
	}

	e.result = Result{
		Output: output,
		Status: newStatus(cmd, script, input, output, err),
	}
}
