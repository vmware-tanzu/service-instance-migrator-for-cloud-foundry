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

package httpclient_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakelogger "github.com/cloudfoundry/bosh-utils/logger/fakes"
	. "github.com/cloudfoundry/bosh-utils/retrystrategy"
)

var _ = Describe("AttemptRetryStrategy", func() {
	var (
		logger *fakelogger.FakeLogger
	)

	BeforeEach(func() {
		logger = &fakelogger.FakeLogger{}
	})

	Describe("Try", func() {
		It("includes type of retryable in log message", func() {
			retryable := newSimpleRetryable([]attemptOutput{
				{
					ShouldRetry: false,
					AttemptErr:  nil,
				},
			})

			attemptRetryStrategy := NewAttemptRetryStrategy(3, 0, retryable, logger)
			err := attemptRetryStrategy.Try()
			Expect(err).ToNot(HaveOccurred())
			Expect(logger.DebugCallCount()).To(Equal(1))

			_, message, args := logger.DebugArgsForCall(0)
			Expect(fmt.Sprintf(message, args...)).To(Equal("Making attempt #0 for *httpclient_test.simpleRetryable"))
		})

		Context("when the request is retryable", func() {
			Context("and the request has errored", func() {
				It("retries until the max attempts are used and returns an error", func() {
					retryable := newSimpleRetryable([]attemptOutput{
						{
							ShouldRetry: true,
							AttemptErr:  errors.New("one"),
						},
						{
							ShouldRetry: true,
							AttemptErr:  errors.New("two"),
						},
						{
							ShouldRetry: true,
							AttemptErr:  errors.New("three"),
						},
					})

					attemptRetryStrategy := NewAttemptRetryStrategy(3, 0, retryable, logger)
					err := attemptRetryStrategy.Try()
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("three"))
					Expect(retryable.Attempts).To(Equal(3))
				})
			})

			Context("and the attempt does not error", func() {
				It("retries until the max attempts are used and does not return an error", func() {
					retryable := newSimpleRetryable([]attemptOutput{
						{
							ShouldRetry: true,
							AttemptErr:  nil,
						},
						{
							ShouldRetry: true,
							AttemptErr:  nil,
						},
						{
							ShouldRetry: true,
							AttemptErr:  nil,
						},
					})
					attemptRetryStrategy := NewAttemptRetryStrategy(3, 0, retryable, logger)
					err := attemptRetryStrategy.Try()
					Expect(err).NotTo(HaveOccurred())
					Expect(retryable.Attempts).To(Equal(3))
				})
			})
		})

		Context("when the attempt is not retryable", func() {
			It("stops trying", func() {
				retryable := newSimpleRetryable([]attemptOutput{
					{
						ShouldRetry: true,
						AttemptErr:  errors.New("first-error"),
					},
					{
						ShouldRetry: false,
						AttemptErr:  errors.New("second-error"),
					},
				})
				attemptRetryStrategy := NewAttemptRetryStrategy(10, 0, retryable, logger)
				err := attemptRetryStrategy.Try()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("second-error"))
				Expect(retryable.Attempts).To(Equal(2))
			})

			It("stops trying", func() {
				retryable := newSimpleRetryable([]attemptOutput{
					{
						ShouldRetry: false,
						AttemptErr:  nil,
					},
				})

				attemptRetryStrategy := NewAttemptRetryStrategy(10, 0, retryable, logger)
				err := attemptRetryStrategy.Try()
				Expect(err).NotTo(HaveOccurred())
				Expect(retryable.Attempts).To(Equal(1))
			})
		})
	})
})
