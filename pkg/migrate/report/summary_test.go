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

package report

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
)

var errServiceError = errors.New("this is an example error")

func TestAddFailedServiceWithNoName(t *testing.T) {
	s := NewSummary(&bytes.Buffer{})
	s.AddFailedService("", "", "", "", errServiceError)
	if s.ServiceFailureCount() != 0 {
		t.Fatalf("Expected there to be 3 failed services, but instead got %d", s.ServiceFailureCount())
	}
}

func TestAddFailedServices(t *testing.T) {
	failedServiceNames := []string{fmt.Sprintf(keyFormat, "my-org", "my-space", "my-favorite-service", "p.mysql"), fmt.Sprintf(keyFormat, "my-org", "my-space", "playground/my-favorite-service", "ecs"), fmt.Sprintf(keyFormat, "my-org", "my-space", "@#$|^&*(UI", "sqlserver")}
	s := NewSummary(&bytes.Buffer{})

	for _, n := range failedServiceNames {
		svc := strings.Split(n, separator)
		s.AddFailedService(svc[0], svc[1], svc[2], svc[3], errServiceError)
	}

	if s.ServiceFailureCount() != 3 {
		t.Fatalf("Expected there to be 3 failed services, but instead got %d", s.ServiceFailureCount())
	}

	serviceFailures := s.Results()
	for _, name := range failedServiceNames {
		var failure *Result
		for _, f := range serviceFailures {
			if name == fmt.Sprintf(keyFormat, f.OrgName, f.SpaceName, f.ServiceName, f.Service) {
				failure = &f
				break
			}
		}

		if failure == nil {
			t.Fatalf("Could not find failed service " + name)
		}
		if failure.Message != errServiceError.Error() {
			t.Fatalf(fmt.Sprintf("We expected %s error to be: %s but instead we got %s.", name, errServiceError, failure.Message))
		}
	}
}

func TestAddFailedServicesConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	s := NewSummary(&bytes.Buffer{})
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(failedServiceNum int) {
			failedServiceName := fmt.Sprintf("FailedService%d", failedServiceNum)
			s.AddFailedService("", "", failedServiceName, "p.mysql", errServiceError)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if s.ServiceFailureCount() != 50 {
		t.Fatalf("Expected there to be 50 failed services, but instead got %d", s.ServiceFailureCount())
	}
	if s.ServiceSuccessCount() != 0 {
		t.Fatalf("Expected there to be 0 successful services, but instead got %d", s.ServiceSuccessCount())
	}
}

func TestAddSkippedServicesConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	s := NewSummary(&bytes.Buffer{})
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(skippedServiceNum int) {
			skippedServiceName := fmt.Sprintf("SkippedService%d", skippedServiceNum)
			s.AddSkippedService("", "", skippedServiceName, "p.mysql", errServiceError)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if s.ServiceSkippedCount() != 50 {
		t.Fatalf("Expected there to be 50 skipped services, but instead got %d", s.ServiceSkippedCount())
	}
	if s.ServiceSuccessCount() != 0 {
		t.Fatalf("Expected there to be 0 successful services, but instead got %d", s.ServiceSuccessCount())
	}
	if s.ServiceFailureCount() != 0 {
		t.Fatalf("Expected there to be 0 failed services, but instead got %d", s.ServiceFailureCount())
	}
}

func TestAddSuccessfulServicesConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	s := NewSummary(&bytes.Buffer{})
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func() {
			s.AddSuccessfulService("", "", "", "")
			wg.Done()
		}()
	}
	wg.Wait()

	if s.ServiceFailureCount() != 0 {
		t.Fatalf("Expected there to be 0 failed services, but instead got %d", s.ServiceFailureCount())
	}
	if s.ServiceSuccessCount() != 50 {
		t.Fatalf("Expected there to be 50 successful services, but instead got %d", s.ServiceSuccessCount())
	}
}

func TestSummary_Display(t *testing.T) {
	type fields struct {
		TableWriter    *bytes.Buffer
		SuccessResults []map[string]string
		FailureResults []map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "prints results table sorted by org, space, and service",
			fields: fields{
				TableWriter: &bytes.Buffer{},
				SuccessResults: []map[string]string{
					{"org": "blue", "space": "dev", "name": "my-good-service", "service": "p.mysql"},
					{"org": "blue", "space": "dev", "name": "another-good-service", "service": "credhub"},
					{"org": "blue", "space": "stage", "name": "my-good-service", "service": "sqlserver"},
				},
				FailureResults: []map[string]string{
					{"org": "red", "space": "dev", "name": "my-bad-service", "service": "ecs", "error": errServiceError.Error()},
				},
			},
			want: `Org       Space     Name                  Service    Result
blue      dev       another-good-service  credhub    successful
blue      dev       my-good-service       p.mysql    successful
blue      stage     my-good-service       sqlserver  successful
red       dev       my-bad-service        ecs        this is an example error
`,
		},
		{
			name: "do not display header when results are empty",
			fields: fields{
				TableWriter: &bytes.Buffer{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummary(tt.fields.TableWriter)
			for _, sr := range tt.fields.SuccessResults {
				s.AddSuccessfulService(sr["org"], sr["space"], sr["name"], sr["service"])
			}
			for _, sr := range tt.fields.FailureResults {
				s.AddFailedService(sr["org"], sr["space"], sr["name"], sr["service"], errors.New(sr["error"]))
			}
			s.Display()
			assert.Equal(t, tt.want, tt.fields.TableWriter.String())
		})
	}
}
