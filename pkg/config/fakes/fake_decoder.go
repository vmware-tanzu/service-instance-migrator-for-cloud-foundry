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
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"sync"
)

type FakeDecoder struct {
	DecodeStub        func(config.Migration, string) interface{}
	decodeMutex       sync.RWMutex
	decodeArgsForCall []struct {
		arg1 config.Migration
		arg2 string
	}
	decodeReturns struct {
		result1 interface{}
	}
	decodeReturnsOnCall map[int]struct {
		result1 interface{}
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDecoder) Decode(arg1 config.Migration, arg2 string) interface{} {
	fake.decodeMutex.Lock()
	ret, specificReturn := fake.decodeReturnsOnCall[len(fake.decodeArgsForCall)]
	fake.decodeArgsForCall = append(fake.decodeArgsForCall, struct {
		arg1 config.Migration
		arg2 string
	}{arg1, arg2})
	stub := fake.DecodeStub
	fakeReturns := fake.decodeReturns
	fake.recordInvocation("Decode", []interface{}{arg1, arg2})
	fake.decodeMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDecoder) DecodeCallCount() int {
	fake.decodeMutex.RLock()
	defer fake.decodeMutex.RUnlock()
	return len(fake.decodeArgsForCall)
}

func (fake *FakeDecoder) DecodeCalls(stub func(config.Migration, string) interface{}) {
	fake.decodeMutex.Lock()
	defer fake.decodeMutex.Unlock()
	fake.DecodeStub = stub
}

func (fake *FakeDecoder) DecodeArgsForCall(i int) (config.Migration, string) {
	fake.decodeMutex.RLock()
	defer fake.decodeMutex.RUnlock()
	argsForCall := fake.decodeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDecoder) DecodeReturns(result1 interface{}) {
	fake.decodeMutex.Lock()
	defer fake.decodeMutex.Unlock()
	fake.DecodeStub = nil
	fake.decodeReturns = struct {
		result1 interface{}
	}{result1}
}

func (fake *FakeDecoder) DecodeReturnsOnCall(i int, result1 interface{}) {
	fake.decodeMutex.Lock()
	defer fake.decodeMutex.Unlock()
	fake.DecodeStub = nil
	if fake.decodeReturnsOnCall == nil {
		fake.decodeReturnsOnCall = make(map[int]struct {
			result1 interface{}
		})
	}
	fake.decodeReturnsOnCall[i] = struct {
		result1 interface{}
	}{result1}
}

func (fake *FakeDecoder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.decodeMutex.RLock()
	defer fake.decodeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDecoder) recordInvocation(key string, args []interface{}) {
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

var _ config.MigrationDecoder = new(FakeDecoder)
