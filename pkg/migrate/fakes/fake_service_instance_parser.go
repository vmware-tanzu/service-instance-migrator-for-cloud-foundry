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

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/io"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
)

type FakeServiceInstanceParser struct {
	MarshalStub        func(interface{}, io.FileDescriptor) error
	marshalMutex       sync.RWMutex
	marshalArgsForCall []struct {
		arg1 interface{}
		arg2 io.FileDescriptor
	}
	marshalReturns struct {
		result1 error
	}
	marshalReturnsOnCall map[int]struct {
		result1 error
	}
	UnmarshalStub        func(interface{}, io.FileDescriptor) error
	unmarshalMutex       sync.RWMutex
	unmarshalArgsForCall []struct {
		arg1 interface{}
		arg2 io.FileDescriptor
	}
	unmarshalReturns struct {
		result1 error
	}
	unmarshalReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeServiceInstanceParser) Marshal(arg1 interface{}, arg2 io.FileDescriptor) error {
	fake.marshalMutex.Lock()
	ret, specificReturn := fake.marshalReturnsOnCall[len(fake.marshalArgsForCall)]
	fake.marshalArgsForCall = append(fake.marshalArgsForCall, struct {
		arg1 interface{}
		arg2 io.FileDescriptor
	}{arg1, arg2})
	stub := fake.MarshalStub
	fakeReturns := fake.marshalReturns
	fake.recordInvocation("Marshal", []interface{}{arg1, arg2})
	fake.marshalMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceInstanceParser) MarshalCallCount() int {
	fake.marshalMutex.RLock()
	defer fake.marshalMutex.RUnlock()
	return len(fake.marshalArgsForCall)
}

func (fake *FakeServiceInstanceParser) MarshalCalls(stub func(interface{}, io.FileDescriptor) error) {
	fake.marshalMutex.Lock()
	defer fake.marshalMutex.Unlock()
	fake.MarshalStub = stub
}

func (fake *FakeServiceInstanceParser) MarshalArgsForCall(i int) (interface{}, io.FileDescriptor) {
	fake.marshalMutex.RLock()
	defer fake.marshalMutex.RUnlock()
	argsForCall := fake.marshalArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceInstanceParser) MarshalReturns(result1 error) {
	fake.marshalMutex.Lock()
	defer fake.marshalMutex.Unlock()
	fake.MarshalStub = nil
	fake.marshalReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceInstanceParser) MarshalReturnsOnCall(i int, result1 error) {
	fake.marshalMutex.Lock()
	defer fake.marshalMutex.Unlock()
	fake.MarshalStub = nil
	if fake.marshalReturnsOnCall == nil {
		fake.marshalReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.marshalReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceInstanceParser) Unmarshal(arg1 interface{}, arg2 io.FileDescriptor) error {
	fake.unmarshalMutex.Lock()
	ret, specificReturn := fake.unmarshalReturnsOnCall[len(fake.unmarshalArgsForCall)]
	fake.unmarshalArgsForCall = append(fake.unmarshalArgsForCall, struct {
		arg1 interface{}
		arg2 io.FileDescriptor
	}{arg1, arg2})
	stub := fake.UnmarshalStub
	fakeReturns := fake.unmarshalReturns
	fake.recordInvocation("Unmarshal", []interface{}{arg1, arg2})
	fake.unmarshalMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeServiceInstanceParser) UnmarshalCallCount() int {
	fake.unmarshalMutex.RLock()
	defer fake.unmarshalMutex.RUnlock()
	return len(fake.unmarshalArgsForCall)
}

func (fake *FakeServiceInstanceParser) UnmarshalCalls(stub func(interface{}, io.FileDescriptor) error) {
	fake.unmarshalMutex.Lock()
	defer fake.unmarshalMutex.Unlock()
	fake.UnmarshalStub = stub
}

func (fake *FakeServiceInstanceParser) UnmarshalArgsForCall(i int) (interface{}, io.FileDescriptor) {
	fake.unmarshalMutex.RLock()
	defer fake.unmarshalMutex.RUnlock()
	argsForCall := fake.unmarshalArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeServiceInstanceParser) UnmarshalReturns(result1 error) {
	fake.unmarshalMutex.Lock()
	defer fake.unmarshalMutex.Unlock()
	fake.UnmarshalStub = nil
	fake.unmarshalReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceInstanceParser) UnmarshalReturnsOnCall(i int, result1 error) {
	fake.unmarshalMutex.Lock()
	defer fake.unmarshalMutex.Unlock()
	fake.UnmarshalStub = nil
	if fake.unmarshalReturnsOnCall == nil {
		fake.unmarshalReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unmarshalReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeServiceInstanceParser) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.marshalMutex.RLock()
	defer fake.marshalMutex.RUnlock()
	fake.unmarshalMutex.RLock()
	defer fake.unmarshalMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeServiceInstanceParser) recordInvocation(key string, args []interface{}) {
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

var _ migrate.ServiceInstanceParser = new(FakeServiceInstanceParser)