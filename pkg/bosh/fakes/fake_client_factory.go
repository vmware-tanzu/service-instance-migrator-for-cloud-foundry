// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/bosh"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

type FakeClientFactory struct {
	NewStub        func(string, string, []byte, bosh.CertAppender, config.Authentication) (bosh.Client, error)
	newMutex       sync.RWMutex
	newArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 []byte
		arg4 bosh.CertAppender
		arg5 config.Authentication
	}
	newReturns struct {
		result1 bosh.Client
		result2 error
	}
	newReturnsOnCall map[int]struct {
		result1 bosh.Client
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClientFactory) New(arg1 string, arg2 string, arg3 []byte, arg4 bosh.CertAppender, arg5 config.Authentication) (bosh.Client, error) {
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.newMutex.Lock()
	ret, specificReturn := fake.newReturnsOnCall[len(fake.newArgsForCall)]
	fake.newArgsForCall = append(fake.newArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 []byte
		arg4 bosh.CertAppender
		arg5 config.Authentication
	}{arg1, arg2, arg3Copy, arg4, arg5})
	stub := fake.NewStub
	fakeReturns := fake.newReturns
	fake.recordInvocation("New", []interface{}{arg1, arg2, arg3Copy, arg4, arg5})
	fake.newMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeClientFactory) NewCallCount() int {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	return len(fake.newArgsForCall)
}

func (fake *FakeClientFactory) NewCalls(stub func(string, string, []byte, bosh.CertAppender, config.Authentication) (bosh.Client, error)) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = stub
}

func (fake *FakeClientFactory) NewArgsForCall(i int) (string, string, []byte, bosh.CertAppender, config.Authentication) {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	argsForCall := fake.newArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeClientFactory) NewReturns(result1 bosh.Client, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	fake.newReturns = struct {
		result1 bosh.Client
		result2 error
	}{result1, result2}
}

func (fake *FakeClientFactory) NewReturnsOnCall(i int, result1 bosh.Client, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	if fake.newReturnsOnCall == nil {
		fake.newReturnsOnCall = make(map[int]struct {
			result1 bosh.Client
			result2 error
		})
	}
	fake.newReturnsOnCall[i] = struct {
		result1 bosh.Client
		result2 error
	}{result1, result2}
}

func (fake *FakeClientFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClientFactory) recordInvocation(key string, args []interface{}) {
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

var _ bosh.ClientFactory = new(FakeClientFactory)
