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
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrNotFound      = errors.New("command not found")
	ErrNotExecutable = errors.New("command not executable")
	ErrInvalidArgs   = errors.New("invalid arguments in command")
)

type Status struct {
	ScriptName string
	ScriptBody string
	Output     string
	PID        int
	Done       bool
	CPUTime    time.Duration
	ExitCode   int
	Error      error
}

func newStatus(cmd *exec.Cmd, script string, body string, output string, err error) *Status {
	s := &Status{
		ScriptName: script,
		ScriptBody: body,
		Output:     output,
		Error:      StatusError(err),
	}
	if cmd.ProcessState != nil {
		s.PID = cmd.ProcessState.Pid()
		s.Done = cmd.ProcessState.Exited()
		s.CPUTime = cmd.ProcessState.SystemTime()
		s.ExitCode = cmd.ProcessState.ExitCode()
	}
	return s
}

func (s *Status) Sprintf(msg string) string {
	return fmt.Sprintf("%s: error: %s, script: %q, body: %q, output: %q, exit code: %d", msg, s.Error, s.ScriptName, s.ScriptBody, s.Output, s.ExitCode)
}

func StatusError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case strings.Contains(err.Error(), "exit status 126"):
		return ErrNotExecutable
	case strings.Contains(err.Error(), "exit status 127"):
		return ErrNotFound
	case strings.Contains(err.Error(), "exit status 128"):
		return ErrInvalidArgs
	}

	return err
}
