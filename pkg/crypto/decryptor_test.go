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

package crypto_test

import (
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/crypto"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDecrypt(t *testing.T) {
	spec.Run(t, "Decrypt", testDecrypt, spec.Report(report.Terminal{}))
}

func testDecrypt(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("using legacy cloud controller encryption", func() {
		it("can decrypt the key", func() {
			s, err := crypto.Decrypt("W9gqsW+Z+tcE61WDFbuCPvdzJIHgv5cILkTJzZWAtBhga5hfnI9Q5YNt6MZSzSI1", "153ce768", "W_XislJAhmKDiMQT7oHybm63_yyd9HLG")
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("AGfbLuf8FeXczfqVKk1_uB6w9LJOW6Y0"))
		})

		it("can encrypt the key", func() {
			s, err := crypto.Encrypt("AGfbLuf8FeXczfqVKk1_uB6w9LJOW6Y0", "153ce768", "W_XislJAhmKDiMQT7oHybm63_yyd9HLG")
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("W9gqsW+Z+tcE61WDFbuCPvdzJIHgv5cILkTJzZWAtBhga5hfnI9Q5YNt6MZSzSI1"))
		})

		when("using invalid salt", func() {
			it("fails to decrypt", func() {
				s, err := crypto.Decrypt("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r", "00000000", "zP2vzTyH_wvwH-NlhSFTn2vB88QQT3Mf")
				Expect(err).To(HaveOccurred())
				Expect(s).To(BeZero())
			})
		})
	})

	when("using current cloud controller encryption", func() {
		it("can decrypt the key", func() {
			s, err := crypto.Decrypt("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r", "359f7a8c88fe1aea", "zP2vzTyH_wvwH-NlhSFTn2vB88QQT3Mf")
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("cbwZxqI0c2GgAkzLYs_NvUHo1Bf6aC07"))
		})

		it("can encrypt the key", func() {
			s, err := crypto.Encrypt("cbwZxqI0c2GgAkzLYs_NvUHo1Bf6aC07", "359f7a8c88fe1aea", "zP2vzTyH_wvwH-NlhSFTn2vB88QQT3Mf")
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r"))
		})

		when("using invalid salt", func() {
			it("fails to decrypt", func() {
				s, err := crypto.Decrypt("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r", "0000000000000000", "zP2vzTyH_wvwH-NlhSFTn2vB88QQT3Mf")
				Expect(err).To(HaveOccurred())
				Expect(s).To(BeZero())
			})
		})
	})

	when("using invalid encrypted data", func() {
		it("fails", func() {
			s, err := crypto.Decrypt("this-is-invalid", "359f7a8c88fe1aea", "zP2vzTyH_wvwH-NlhSFTn2vB88QQT3Mf")
			Expect(err).To(HaveOccurred())
			Expect(s).To(BeZero())
		})
	})

	when("using an invalid encryption key", func() {
		it("fails", func() {
			s, err := crypto.Decrypt("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r", "153ce768", "invalid-key")
			Expect(err).To(HaveOccurred())
			Expect(s).To(BeZero())

			s, err = crypto.Decrypt("igNofYRgGrq8i9su+5mNYTrc+YIDw3NUIgIuPkRBhkx3Z9Y+EJKXDAu++WXaK7+r", "359f7a8c88fe1aea", "invalid-key")
			Expect(err).To(HaveOccurred())
			Expect(s).To(BeZero())
		})
	})
}
