// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/credhub"
)

type FakeClientFactory struct {
	NewStub        func(string, string, string, string, []byte, string, string) credhub.Client
	newMutex       sync.RWMutex
	newArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 []byte
		arg6 string
		arg7 string
	}
	newReturns struct {
		result1 credhub.Client
	}
	newReturnsOnCall map[int]struct {
		result1 credhub.Client
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClientFactory) New(arg1 string, arg2 string, arg3 string, arg4 string, arg5 []byte, arg6 string, arg7 string) credhub.Client {
	var arg5Copy []byte
	if arg5 != nil {
		arg5Copy = make([]byte, len(arg5))
		copy(arg5Copy, arg5)
	}
	fake.newMutex.Lock()
	ret, specificReturn := fake.newReturnsOnCall[len(fake.newArgsForCall)]
	fake.newArgsForCall = append(fake.newArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
		arg5 []byte
		arg6 string
		arg7 string
	}{arg1, arg2, arg3, arg4, arg5Copy, arg6, arg7})
	stub := fake.NewStub
	fakeReturns := fake.newReturns
	fake.recordInvocation("New", []interface{}{arg1, arg2, arg3, arg4, arg5Copy, arg6, arg7})
	fake.newMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeClientFactory) NewCallCount() int {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	return len(fake.newArgsForCall)
}

func (fake *FakeClientFactory) NewCalls(stub func(string, string, string, string, []byte, string, string) credhub.Client) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = stub
}

func (fake *FakeClientFactory) NewArgsForCall(i int) (string, string, string, string, []byte, string, string) {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	argsForCall := fake.newArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7
}

func (fake *FakeClientFactory) NewReturns(result1 credhub.Client) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	fake.newReturns = struct {
		result1 credhub.Client
	}{result1}
}

func (fake *FakeClientFactory) NewReturnsOnCall(i int, result1 credhub.Client) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	if fake.newReturnsOnCall == nil {
		fake.newReturnsOnCall = make(map[int]struct {
			result1 credhub.Client
		})
	}
	fake.newReturnsOnCall[i] = struct {
		result1 credhub.Client
	}{result1}
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

var _ credhub.ClientFactory = new(FakeClientFactory)
