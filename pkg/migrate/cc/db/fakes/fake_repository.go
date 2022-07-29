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

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate/cc/db"
)

type FakeRepository struct {
	CreateServiceBindingStub        func(cfclient.ServiceBinding, string, string) error
	createServiceBindingMutex       sync.RWMutex
	createServiceBindingArgsForCall []struct {
		arg1 cfclient.ServiceBinding
		arg2 string
		arg3 string
	}
	createServiceBindingReturns struct {
		result1 error
	}
	createServiceBindingReturnsOnCall map[int]struct {
		result1 error
	}
	CreateServiceInstanceStub        func(cfclient.ServiceInstance, cfclient.Space, cfclient.ServicePlan, cfclient.Service, string) error
	createServiceInstanceMutex       sync.RWMutex
	createServiceInstanceArgsForCall []struct {
		arg1 cfclient.ServiceInstance
		arg2 cfclient.Space
		arg3 cfclient.ServicePlan
		arg4 cfclient.Service
		arg5 string
	}
	createServiceInstanceReturns struct {
		result1 error
	}
	createServiceInstanceReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteServiceInstanceStub        func(string, string) (bool, error)
	deleteServiceInstanceMutex       sync.RWMutex
	deleteServiceInstanceArgsForCall []struct {
		arg1 string
		arg2 string
	}
	deleteServiceInstanceReturns struct {
		result1 bool
		result2 error
	}
	deleteServiceInstanceReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	ServiceInstanceExistsStub        func(string) (bool, error)
	serviceInstanceExistsMutex       sync.RWMutex
	serviceInstanceExistsArgsForCall []struct {
		arg1 string
	}
	serviceInstanceExistsReturns struct {
		result1 bool
		result2 error
	}
	serviceInstanceExistsReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRepository) CreateServiceBinding(arg1 cfclient.ServiceBinding, arg2 string, arg3 string) error {
	fake.createServiceBindingMutex.Lock()
	ret, specificReturn := fake.createServiceBindingReturnsOnCall[len(fake.createServiceBindingArgsForCall)]
	fake.createServiceBindingArgsForCall = append(fake.createServiceBindingArgsForCall, struct {
		arg1 cfclient.ServiceBinding
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.CreateServiceBindingStub
	fakeReturns := fake.createServiceBindingReturns
	fake.recordInvocation("CreateServiceBinding", []interface{}{arg1, arg2, arg3})
	fake.createServiceBindingMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) CreateServiceBindingCallCount() int {
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	return len(fake.createServiceBindingArgsForCall)
}

func (fake *FakeRepository) CreateServiceBindingCalls(stub func(cfclient.ServiceBinding, string, string) error) {
	fake.createServiceBindingMutex.Lock()
	defer fake.createServiceBindingMutex.Unlock()
	fake.CreateServiceBindingStub = stub
}

func (fake *FakeRepository) CreateServiceBindingArgsForCall(i int) (cfclient.ServiceBinding, string, string) {
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	argsForCall := fake.createServiceBindingArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeRepository) CreateServiceBindingReturns(result1 error) {
	fake.createServiceBindingMutex.Lock()
	defer fake.createServiceBindingMutex.Unlock()
	fake.CreateServiceBindingStub = nil
	fake.createServiceBindingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) CreateServiceBindingReturnsOnCall(i int, result1 error) {
	fake.createServiceBindingMutex.Lock()
	defer fake.createServiceBindingMutex.Unlock()
	fake.CreateServiceBindingStub = nil
	if fake.createServiceBindingReturnsOnCall == nil {
		fake.createServiceBindingReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createServiceBindingReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) CreateServiceInstance(arg1 cfclient.ServiceInstance, arg2 cfclient.Space, arg3 cfclient.ServicePlan, arg4 cfclient.Service, arg5 string) error {
	fake.createServiceInstanceMutex.Lock()
	ret, specificReturn := fake.createServiceInstanceReturnsOnCall[len(fake.createServiceInstanceArgsForCall)]
	fake.createServiceInstanceArgsForCall = append(fake.createServiceInstanceArgsForCall, struct {
		arg1 cfclient.ServiceInstance
		arg2 cfclient.Space
		arg3 cfclient.ServicePlan
		arg4 cfclient.Service
		arg5 string
	}{arg1, arg2, arg3, arg4, arg5})
	stub := fake.CreateServiceInstanceStub
	fakeReturns := fake.createServiceInstanceReturns
	fake.recordInvocation("CreateServiceInstance", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.createServiceInstanceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRepository) CreateServiceInstanceCallCount() int {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	return len(fake.createServiceInstanceArgsForCall)
}

func (fake *FakeRepository) CreateServiceInstanceCalls(stub func(cfclient.ServiceInstance, cfclient.Space, cfclient.ServicePlan, cfclient.Service, string) error) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = stub
}

func (fake *FakeRepository) CreateServiceInstanceArgsForCall(i int) (cfclient.ServiceInstance, cfclient.Space, cfclient.ServicePlan, cfclient.Service, string) {
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	argsForCall := fake.createServiceInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeRepository) CreateServiceInstanceReturns(result1 error) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = nil
	fake.createServiceInstanceReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) CreateServiceInstanceReturnsOnCall(i int, result1 error) {
	fake.createServiceInstanceMutex.Lock()
	defer fake.createServiceInstanceMutex.Unlock()
	fake.CreateServiceInstanceStub = nil
	if fake.createServiceInstanceReturnsOnCall == nil {
		fake.createServiceInstanceReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createServiceInstanceReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRepository) DeleteServiceInstance(arg1 string, arg2 string) (bool, error) {
	fake.deleteServiceInstanceMutex.Lock()
	ret, specificReturn := fake.deleteServiceInstanceReturnsOnCall[len(fake.deleteServiceInstanceArgsForCall)]
	fake.deleteServiceInstanceArgsForCall = append(fake.deleteServiceInstanceArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.DeleteServiceInstanceStub
	fakeReturns := fake.deleteServiceInstanceReturns
	fake.recordInvocation("DeleteServiceInstance", []interface{}{arg1, arg2})
	fake.deleteServiceInstanceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) DeleteServiceInstanceCallCount() int {
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	return len(fake.deleteServiceInstanceArgsForCall)
}

func (fake *FakeRepository) DeleteServiceInstanceCalls(stub func(string, string) (bool, error)) {
	fake.deleteServiceInstanceMutex.Lock()
	defer fake.deleteServiceInstanceMutex.Unlock()
	fake.DeleteServiceInstanceStub = stub
}

func (fake *FakeRepository) DeleteServiceInstanceArgsForCall(i int) (string, string) {
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	argsForCall := fake.deleteServiceInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRepository) DeleteServiceInstanceReturns(result1 bool, result2 error) {
	fake.deleteServiceInstanceMutex.Lock()
	defer fake.deleteServiceInstanceMutex.Unlock()
	fake.DeleteServiceInstanceStub = nil
	fake.deleteServiceInstanceReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) DeleteServiceInstanceReturnsOnCall(i int, result1 bool, result2 error) {
	fake.deleteServiceInstanceMutex.Lock()
	defer fake.deleteServiceInstanceMutex.Unlock()
	fake.DeleteServiceInstanceStub = nil
	if fake.deleteServiceInstanceReturnsOnCall == nil {
		fake.deleteServiceInstanceReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.deleteServiceInstanceReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) ServiceInstanceExists(arg1 string) (bool, error) {
	fake.serviceInstanceExistsMutex.Lock()
	ret, specificReturn := fake.serviceInstanceExistsReturnsOnCall[len(fake.serviceInstanceExistsArgsForCall)]
	fake.serviceInstanceExistsArgsForCall = append(fake.serviceInstanceExistsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ServiceInstanceExistsStub
	fakeReturns := fake.serviceInstanceExistsReturns
	fake.recordInvocation("ServiceInstanceExists", []interface{}{arg1})
	fake.serviceInstanceExistsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeRepository) ServiceInstanceExistsCallCount() int {
	fake.serviceInstanceExistsMutex.RLock()
	defer fake.serviceInstanceExistsMutex.RUnlock()
	return len(fake.serviceInstanceExistsArgsForCall)
}

func (fake *FakeRepository) ServiceInstanceExistsCalls(stub func(string) (bool, error)) {
	fake.serviceInstanceExistsMutex.Lock()
	defer fake.serviceInstanceExistsMutex.Unlock()
	fake.ServiceInstanceExistsStub = stub
}

func (fake *FakeRepository) ServiceInstanceExistsArgsForCall(i int) string {
	fake.serviceInstanceExistsMutex.RLock()
	defer fake.serviceInstanceExistsMutex.RUnlock()
	argsForCall := fake.serviceInstanceExistsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeRepository) ServiceInstanceExistsReturns(result1 bool, result2 error) {
	fake.serviceInstanceExistsMutex.Lock()
	defer fake.serviceInstanceExistsMutex.Unlock()
	fake.ServiceInstanceExistsStub = nil
	fake.serviceInstanceExistsReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) ServiceInstanceExistsReturnsOnCall(i int, result1 bool, result2 error) {
	fake.serviceInstanceExistsMutex.Lock()
	defer fake.serviceInstanceExistsMutex.Unlock()
	fake.ServiceInstanceExistsStub = nil
	if fake.serviceInstanceExistsReturnsOnCall == nil {
		fake.serviceInstanceExistsReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.serviceInstanceExistsReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createServiceBindingMutex.RLock()
	defer fake.createServiceBindingMutex.RUnlock()
	fake.createServiceInstanceMutex.RLock()
	defer fake.createServiceInstanceMutex.RUnlock()
	fake.deleteServiceInstanceMutex.RLock()
	defer fake.deleteServiceInstanceMutex.RUnlock()
	fake.serviceInstanceExistsMutex.RLock()
	defer fake.serviceInstanceExistsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRepository) recordInvocation(key string, args []interface{}) {
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

var _ db.Repository = new(FakeRepository)