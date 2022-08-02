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
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
)

const (
	separator string = "%"
	keyFormat string = "%s%%%s%%%s%%%s"
)

// Result is a specific service migration result: success or failure
type Result struct {
	OrgName     string
	SpaceName   string
	ServiceName string
	Service     string
	Message     string
}

// Summary is a thread safe sink of execution results for service migrations
type Summary struct {
	results      map[string]string
	successCount int
	failureCount int
	skippedCount int
	resMutex     sync.RWMutex
	sucMutex     sync.RWMutex
	errMutex     sync.RWMutex
	skipMutex    sync.RWMutex
	TableWriter  io.Writer
}

// NewSummary creates a new initialized summary instance
func NewSummary(w io.Writer) *Summary {
	return &Summary{
		results:     make(map[string]string),
		TableWriter: w,
	}
}

// ServiceFailureCount is the number of total service failures that have occurred
func (s *Summary) ServiceFailureCount() int {
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	return s.failureCount
}

// ServiceSkippedCount is the number of total service skipped that have occurred
func (s *Summary) ServiceSkippedCount() int {
	s.skipMutex.Lock()
	defer s.skipMutex.Unlock()
	return s.skippedCount
}

// ServiceSuccessCount is the number of total service successes that have occurred
func (s *Summary) ServiceSuccessCount() int {
	s.sucMutex.Lock()
	defer s.sucMutex.Unlock()
	return s.successCount
}

// Results returns a copy of all the service migrations that have occurred
func (s *Summary) Results() []Result {
	s.resMutex.Lock()
	defer s.resMutex.Unlock()

	var r []Result
	for k, m := range s.results {
		value := strings.Split(k, separator)
		r = append(r, Result{
			OrgName:     value[0],
			SpaceName:   value[1],
			ServiceName: value[2],
			Service:     value[3],
			Message:     m,
		})
	}
	sort.Slice(r, func(i, j int) bool {
		var sortedByOrgName, sortedBySpaceName bool

		sortedByOrgName = r[i].OrgName < r[j].OrgName
		if r[i].OrgName == r[j].OrgName {
			sortedBySpaceName = r[i].SpaceName < r[j].SpaceName
			if r[i].SpaceName == r[j].SpaceName {
				return r[i].ServiceName < r[j].ServiceName
			}
			return sortedBySpaceName
		}

		return sortedByOrgName
	})

	return r
}

// AddFailedService adds a failed service along with its error to the summary result
func (s *Summary) AddFailedService(org, space, serviceName, serviceType string, err error) {
	if len(serviceName) == 0 {
		return
	}
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	s.failureCount++

	s.resMutex.Lock()
	defer s.resMutex.Unlock()

	s.results[fmt.Sprintf(keyFormat, org, space, serviceName, serviceType)] = err.Error()
}

// AddSkippedService adds a skipped service along with its error to the summary result
func (s *Summary) AddSkippedService(org, space, serviceName, serviceType string, err error) {
	if len(serviceName) == 0 {
		return
	}
	s.skipMutex.Lock()
	defer s.skipMutex.Unlock()
	s.skippedCount++

	s.resMutex.Lock()
	defer s.resMutex.Unlock()

	s.results[fmt.Sprintf(keyFormat, org, space, serviceName, serviceType)] = fmt.Sprintf("skipped: %v", err)
}

// AddSuccessfulService adds a successful service and increments the count of successful services
func (s *Summary) AddSuccessfulService(org, space, serviceName, serviceType string) {
	s.sucMutex.Lock()
	defer s.sucMutex.Unlock()
	s.successCount++

	s.resMutex.Lock()
	defer s.resMutex.Unlock()

	s.results[fmt.Sprintf(keyFormat, org, space, serviceName, serviceType)] = "successful"
}

func (s *Summary) Display() {
	if len(s.Results()) == 0 {
		log.Infoln("Migration summary: no results found")
		return
	}

	tw := tabwriter.NewWriter(s.TableWriter, 10, 2, 2, ' ', 0)
	// Header
	_, _ = fmt.Fprintln(tw, "Org\tSpace\tName\tService\tResult")
	log.Infof("Migration summary: %d successes, %d skipped, %d errors.", s.ServiceSuccessCount(), s.ServiceSkippedCount(), s.ServiceFailureCount())
	fmt.Println()

	for _, f := range s.Results() {
		row := []string{
			f.OrgName,
			f.SpaceName,
			f.ServiceName,
			f.Service,
			strings.Split(f.Message, ":")[0],
		}
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	_ = tw.Flush()
	fmt.Println()
}
