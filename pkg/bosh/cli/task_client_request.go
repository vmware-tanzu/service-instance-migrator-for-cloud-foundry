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

package cli

import (
	"fmt"
	"github.com/cloudfoundry/bosh-cli/director"
	"net/http"
	"time"
)

type TaskClientRequest struct {
	clientRequest         ClientRequest
	taskReporter          director.TaskReporter
	taskCheckStepDuration time.Duration
}

func NewTaskClientRequest(
	clientRequest ClientRequest,
	taskReporter director.TaskReporter,
	taskCheckStepDuration time.Duration,
) TaskClientRequest {
	return TaskClientRequest{
		clientRequest:         clientRequest,
		taskReporter:          taskReporter,
		taskCheckStepDuration: taskCheckStepDuration,
	}
}

type taskShortResp struct {
	ID    int    // 165
	State string // e.g. "queued", "processing", "done", "error", "cancelled"
}

func (r taskShortResp) IsRunning() bool {
	return r.State == "queued" || r.State == "processing" || r.State == "cancelling"
}

func (r taskShortResp) IsSuccessfullyDone() bool {
	return r.State == "done"
}

func (r TaskClientRequest) GetResult(path string) (int, []byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Get(path, &taskResp)
	if err != nil {
		return 0, nil, err
	}

	respBody, err := r.waitForResult(taskResp)

	return taskResp.ID, respBody, err
}

func (r TaskClientRequest) PostResult(path string, payload []byte, f func(*http.Request)) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Post(path, payload, f, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) PutResult(path string, payload []byte, f func(*http.Request)) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Put(path, payload, f, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) DeleteResult(path string) ([]byte, error) {
	var taskResp taskShortResp

	err := r.clientRequest.Delete(path, &taskResp)
	if err != nil {
		return nil, err
	}

	return r.waitForResult(taskResp)
}

func (r TaskClientRequest) WaitForCompletion(id int, type_ string, taskReporter director.TaskReporter) error {
	taskReporter.TaskStarted(id)

	var taskResp taskShortResp
	var outputOffset int

	defer func() {
		taskReporter.TaskFinished(id, taskResp.State)
	}()

	taskPath := fmt.Sprintf("/tasks/%d", id)

	for {
		err := r.clientRequest.Get(taskPath, &taskResp)
		if err != nil {
			return fmt.Errorf("error getting task state: %w", err)
		}

		// retrieve output *after* getting state to make sure
		// it's complete in case of task being finished
		outputOffset, err = r.reportOutputChunk(taskResp.ID, outputOffset, type_, taskReporter)
		if err != nil {
			return fmt.Errorf("error getting task output: %w", err)
		}

		if taskResp.IsRunning() {
			time.Sleep(r.taskCheckStepDuration)
			continue
		}

		if taskResp.IsSuccessfullyDone() {
			return nil
		}

		msgFmt := "error: expected task '%d' to succeed but state is '%s'"

		return fmt.Errorf(msgFmt, taskResp.ID, taskResp.State)
	}
}

func (r TaskClientRequest) waitForResult(taskResp taskShortResp) ([]byte, error) {
	err := r.WaitForCompletion(taskResp.ID, "event", r.taskReporter)
	if err != nil {
		return nil, err
	}

	resultPath := fmt.Sprintf("/tasks/%d/output?type=result", taskResp.ID)

	respBody, _, err := r.clientRequest.RawGet(resultPath, nil, nil)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

type taskReporterWriter struct {
	id           int
	totalLen     int
	taskReporter director.TaskReporter
}

var _ director.ShouldTrackDownload = &taskReporterWriter{}

func (w *taskReporterWriter) Write(buf []byte) (int, error) {
	bufLen := len(buf)
	if bufLen > 0 {
		w.taskReporter.TaskOutputChunk(w.id, buf)
	}
	w.totalLen += bufLen
	return bufLen, nil
}

func (w taskReporterWriter) ShouldTrackDownload() bool { return false }

func (r TaskClientRequest) reportOutputChunk(id, offset int, type_ string, taskReporter director.TaskReporter) (int, error) {
	outputPath := fmt.Sprintf("/tasks/%d/output?type=%s", id, type_)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	writer := &taskReporterWriter{id, 0, taskReporter}

	_, resp, err := r.clientRequest.RawGet(outputPath, writer, setHeaders)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return offset, nil
		}

		return 0, err
	}

	return offset + writer.totalLen, nil
}
