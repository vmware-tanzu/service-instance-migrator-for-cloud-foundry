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

// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
)

type FakeClient struct {
	FindDeploymentStub        func(string) (director.DeploymentResp, bool, error)
	findDeploymentMutex       sync.RWMutex
	findDeploymentArgsForCall []struct {
		arg1 string
	}
	findDeploymentReturns struct {
		result1 director.DeploymentResp
		result2 bool
		result3 error
	}
	findDeploymentReturnsOnCall map[int]struct {
		result1 director.DeploymentResp
		result2 bool
		result3 error
	}
	FindVMStub        func(string, string) (director.VMInfo, bool, error)
	findVMMutex       sync.RWMutex
	findVMArgsForCall []struct {
		arg1 string
		arg2 string
	}
	findVMReturns struct {
		result1 director.VMInfo
		result2 bool
		result3 error
	}
	findVMReturnsOnCall map[int]struct {
		result1 director.VMInfo
		result2 bool
		result3 error
	}
	VerifyAuthStub        func() error
	verifyAuthMutex       sync.RWMutex
	verifyAuthArgsForCall []struct {
	}
	verifyAuthReturns struct {
		result1 error
	}
	verifyAuthReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) FindDeployment(arg1 string) (director.DeploymentResp, bool, error) {
	fake.findDeploymentMutex.Lock()
	ret, specificReturn := fake.findDeploymentReturnsOnCall[len(fake.findDeploymentArgsForCall)]
	fake.findDeploymentArgsForCall = append(fake.findDeploymentArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.FindDeploymentStub
	fakeReturns := fake.findDeploymentReturns
	fake.recordInvocation("FindDeployment", []interface{}{arg1})
	fake.findDeploymentMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeClient) FindDeploymentCallCount() int {
	fake.findDeploymentMutex.RLock()
	defer fake.findDeploymentMutex.RUnlock()
	return len(fake.findDeploymentArgsForCall)
}

func (fake *FakeClient) FindDeploymentCalls(stub func(string) (director.DeploymentResp, bool, error)) {
	fake.findDeploymentMutex.Lock()
	defer fake.findDeploymentMutex.Unlock()
	fake.FindDeploymentStub = stub
}

func (fake *FakeClient) FindDeploymentArgsForCall(i int) string {
	fake.findDeploymentMutex.RLock()
	defer fake.findDeploymentMutex.RUnlock()
	argsForCall := fake.findDeploymentArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeClient) FindDeploymentReturns(result1 director.DeploymentResp, result2 bool, result3 error) {
	fake.findDeploymentMutex.Lock()
	defer fake.findDeploymentMutex.Unlock()
	fake.FindDeploymentStub = nil
	fake.findDeploymentReturns = struct {
		result1 director.DeploymentResp
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeClient) FindDeploymentReturnsOnCall(i int, result1 director.DeploymentResp, result2 bool, result3 error) {
	fake.findDeploymentMutex.Lock()
	defer fake.findDeploymentMutex.Unlock()
	fake.FindDeploymentStub = nil
	if fake.findDeploymentReturnsOnCall == nil {
		fake.findDeploymentReturnsOnCall = make(map[int]struct {
			result1 director.DeploymentResp
			result2 bool
			result3 error
		})
	}
	fake.findDeploymentReturnsOnCall[i] = struct {
		result1 director.DeploymentResp
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeClient) FindVM(arg1 string, arg2 string) (director.VMInfo, bool, error) {
	fake.findVMMutex.Lock()
	ret, specificReturn := fake.findVMReturnsOnCall[len(fake.findVMArgsForCall)]
	fake.findVMArgsForCall = append(fake.findVMArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.FindVMStub
	fakeReturns := fake.findVMReturns
	fake.recordInvocation("FindVM", []interface{}{arg1, arg2})
	fake.findVMMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeClient) FindVMCallCount() int {
	fake.findVMMutex.RLock()
	defer fake.findVMMutex.RUnlock()
	return len(fake.findVMArgsForCall)
}

func (fake *FakeClient) FindVMCalls(stub func(string, string) (director.VMInfo, bool, error)) {
	fake.findVMMutex.Lock()
	defer fake.findVMMutex.Unlock()
	fake.FindVMStub = stub
}

func (fake *FakeClient) FindVMArgsForCall(i int) (string, string) {
	fake.findVMMutex.RLock()
	defer fake.findVMMutex.RUnlock()
	argsForCall := fake.findVMArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeClient) FindVMReturns(result1 director.VMInfo, result2 bool, result3 error) {
	fake.findVMMutex.Lock()
	defer fake.findVMMutex.Unlock()
	fake.FindVMStub = nil
	fake.findVMReturns = struct {
		result1 director.VMInfo
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeClient) FindVMReturnsOnCall(i int, result1 director.VMInfo, result2 bool, result3 error) {
	fake.findVMMutex.Lock()
	defer fake.findVMMutex.Unlock()
	fake.FindVMStub = nil
	if fake.findVMReturnsOnCall == nil {
		fake.findVMReturnsOnCall = make(map[int]struct {
			result1 director.VMInfo
			result2 bool
			result3 error
		})
	}
	fake.findVMReturnsOnCall[i] = struct {
		result1 director.VMInfo
		result2 bool
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeClient) VerifyAuth() error {
	fake.verifyAuthMutex.Lock()
	ret, specificReturn := fake.verifyAuthReturnsOnCall[len(fake.verifyAuthArgsForCall)]
	fake.verifyAuthArgsForCall = append(fake.verifyAuthArgsForCall, struct {
	}{})
	stub := fake.VerifyAuthStub
	fakeReturns := fake.verifyAuthReturns
	fake.recordInvocation("VerifyAuth", []interface{}{})
	fake.verifyAuthMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeClient) VerifyAuthCallCount() int {
	fake.verifyAuthMutex.RLock()
	defer fake.verifyAuthMutex.RUnlock()
	return len(fake.verifyAuthArgsForCall)
}

func (fake *FakeClient) VerifyAuthCalls(stub func() error) {
	fake.verifyAuthMutex.Lock()
	defer fake.verifyAuthMutex.Unlock()
	fake.VerifyAuthStub = stub
}

func (fake *FakeClient) VerifyAuthReturns(result1 error) {
	fake.verifyAuthMutex.Lock()
	defer fake.verifyAuthMutex.Unlock()
	fake.VerifyAuthStub = nil
	fake.verifyAuthReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) VerifyAuthReturnsOnCall(i int, result1 error) {
	fake.verifyAuthMutex.Lock()
	defer fake.verifyAuthMutex.Unlock()
	fake.VerifyAuthStub = nil
	if fake.verifyAuthReturnsOnCall == nil {
		fake.verifyAuthReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.verifyAuthReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.findDeploymentMutex.RLock()
	defer fake.findDeploymentMutex.RUnlock()
	fake.findVMMutex.RLock()
	defer fake.findVMMutex.RUnlock()
	fake.verifyAuthMutex.RLock()
	defer fake.verifyAuthMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ bosh.Client = new(FakeClient)