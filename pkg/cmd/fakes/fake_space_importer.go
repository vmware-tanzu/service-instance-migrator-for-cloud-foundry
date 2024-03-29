// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cmd"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

type FakeSpaceImporter struct {
	ImportStub        func(context.Context, config.OpsManager, string, string, string) error
	importMutex       sync.RWMutex
	importArgsForCall []struct {
		arg1 context.Context
		arg2 config.OpsManager
		arg3 string
		arg4 string
		arg5 string
	}
	importReturns struct {
		result1 error
	}
	importReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSpaceImporter) Import(arg1 context.Context, arg2 config.OpsManager, arg3 string, arg4 string, arg5 string) error {
	fake.importMutex.Lock()
	ret, specificReturn := fake.importReturnsOnCall[len(fake.importArgsForCall)]
	fake.importArgsForCall = append(fake.importArgsForCall, struct {
		arg1 context.Context
		arg2 config.OpsManager
		arg3 string
		arg4 string
		arg5 string
	}{arg1, arg2, arg3, arg4, arg5})
	stub := fake.ImportStub
	fakeReturns := fake.importReturns
	fake.recordInvocation("Import", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.importMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeSpaceImporter) ImportCallCount() int {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	return len(fake.importArgsForCall)
}

func (fake *FakeSpaceImporter) ImportCalls(stub func(context.Context, config.OpsManager, string, string, string) error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = stub
}

func (fake *FakeSpaceImporter) ImportArgsForCall(i int) (context.Context, config.OpsManager, string, string, string) {
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	argsForCall := fake.importArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeSpaceImporter) ImportReturns(result1 error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = nil
	fake.importReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpaceImporter) ImportReturnsOnCall(i int, result1 error) {
	fake.importMutex.Lock()
	defer fake.importMutex.Unlock()
	fake.ImportStub = nil
	if fake.importReturnsOnCall == nil {
		fake.importReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.importReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSpaceImporter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.importMutex.RLock()
	defer fake.importMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSpaceImporter) recordInvocation(key string, args []interface{}) {
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

var _ cmd.SpaceImporter = new(FakeSpaceImporter)
