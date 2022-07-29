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

package httpclient

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/log"
	"time"

	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
)

type attemptRetryStrategy struct {
	maxAttempts int
	delay       time.Duration
	retryable   boshretry.Retryable
}

func NewAttemptRetryStrategy(
	maxAttempts int,
	delay time.Duration,
	retryable boshretry.Retryable,
) boshretry.RetryStrategy {
	return &attemptRetryStrategy{
		maxAttempts: maxAttempts,
		delay:       delay,
		retryable:   retryable,
	}
}

func (s *attemptRetryStrategy) Try() error {
	var err error
	var shouldRetry bool

	for i := 0; i < s.maxAttempts; i++ {
		log.Debugf("Making attempt #%d for %T", i, s.retryable)

		shouldRetry, err = s.retryable.Attempt()
		if !shouldRetry {
			return err
		}

		time.Sleep(s.delay)
	}

	return err
}
