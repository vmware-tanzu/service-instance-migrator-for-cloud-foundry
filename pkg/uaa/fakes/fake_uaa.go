// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/uaa"
)

type FakeUAA struct {
	ClientCredentialsGrantStub        func() (uaa.AccessToken, error)
	clientCredentialsGrantMutex       sync.RWMutex
	clientCredentialsGrantArgsForCall []struct {
	}
	clientCredentialsGrantReturns struct {
		result1 uaa.AccessToken
		result2 error
	}
	clientCredentialsGrantReturnsOnCall map[int]struct {
		result1 uaa.AccessToken
		result2 error
	}
	OwnerPasswordCredentialsGrantStub        func([]uaa.TokenParameters) (uaa.AccessToken, error)
	ownerPasswordCredentialsGrantMutex       sync.RWMutex
	ownerPasswordCredentialsGrantArgsForCall []struct {
		arg1 []uaa.TokenParameters
	}
	ownerPasswordCredentialsGrantReturns struct {
		result1 uaa.AccessToken
		result2 error
	}
	ownerPasswordCredentialsGrantReturnsOnCall map[int]struct {
		result1 uaa.AccessToken
		result2 error
	}
	RefreshTokenGrantStub        func(string) (uaa.AccessToken, error)
	refreshTokenGrantMutex       sync.RWMutex
	refreshTokenGrantArgsForCall []struct {
		arg1 string
	}
	refreshTokenGrantReturns struct {
		result1 uaa.AccessToken
		result2 error
	}
	refreshTokenGrantReturnsOnCall map[int]struct {
		result1 uaa.AccessToken
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeUAA) ClientCredentialsGrant() (uaa.AccessToken, error) {
	fake.clientCredentialsGrantMutex.Lock()
	ret, specificReturn := fake.clientCredentialsGrantReturnsOnCall[len(fake.clientCredentialsGrantArgsForCall)]
	fake.clientCredentialsGrantArgsForCall = append(fake.clientCredentialsGrantArgsForCall, struct {
	}{})
	stub := fake.ClientCredentialsGrantStub
	fakeReturns := fake.clientCredentialsGrantReturns
	fake.recordInvocation("ClientCredentialsGrant", []interface{}{})
	fake.clientCredentialsGrantMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeUAA) ClientCredentialsGrantCallCount() int {
	fake.clientCredentialsGrantMutex.RLock()
	defer fake.clientCredentialsGrantMutex.RUnlock()
	return len(fake.clientCredentialsGrantArgsForCall)
}

func (fake *FakeUAA) ClientCredentialsGrantCalls(stub func() (uaa.AccessToken, error)) {
	fake.clientCredentialsGrantMutex.Lock()
	defer fake.clientCredentialsGrantMutex.Unlock()
	fake.ClientCredentialsGrantStub = stub
}

func (fake *FakeUAA) ClientCredentialsGrantReturns(result1 uaa.AccessToken, result2 error) {
	fake.clientCredentialsGrantMutex.Lock()
	defer fake.clientCredentialsGrantMutex.Unlock()
	fake.ClientCredentialsGrantStub = nil
	fake.clientCredentialsGrantReturns = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) ClientCredentialsGrantReturnsOnCall(i int, result1 uaa.AccessToken, result2 error) {
	fake.clientCredentialsGrantMutex.Lock()
	defer fake.clientCredentialsGrantMutex.Unlock()
	fake.ClientCredentialsGrantStub = nil
	if fake.clientCredentialsGrantReturnsOnCall == nil {
		fake.clientCredentialsGrantReturnsOnCall = make(map[int]struct {
			result1 uaa.AccessToken
			result2 error
		})
	}
	fake.clientCredentialsGrantReturnsOnCall[i] = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrant(arg1 []uaa.TokenParameters) (uaa.AccessToken, error) {
	var arg1Copy []uaa.TokenParameters
	if arg1 != nil {
		arg1Copy = make([]uaa.TokenParameters, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.ownerPasswordCredentialsGrantMutex.Lock()
	ret, specificReturn := fake.ownerPasswordCredentialsGrantReturnsOnCall[len(fake.ownerPasswordCredentialsGrantArgsForCall)]
	fake.ownerPasswordCredentialsGrantArgsForCall = append(fake.ownerPasswordCredentialsGrantArgsForCall, struct {
		arg1 []uaa.TokenParameters
	}{arg1Copy})
	stub := fake.OwnerPasswordCredentialsGrantStub
	fakeReturns := fake.ownerPasswordCredentialsGrantReturns
	fake.recordInvocation("OwnerPasswordCredentialsGrant", []interface{}{arg1Copy})
	fake.ownerPasswordCredentialsGrantMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantCallCount() int {
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	return len(fake.ownerPasswordCredentialsGrantArgsForCall)
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantCalls(stub func([]uaa.TokenParameters) (uaa.AccessToken, error)) {
	fake.ownerPasswordCredentialsGrantMutex.Lock()
	defer fake.ownerPasswordCredentialsGrantMutex.Unlock()
	fake.OwnerPasswordCredentialsGrantStub = stub
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantArgsForCall(i int) []uaa.TokenParameters {
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	argsForCall := fake.ownerPasswordCredentialsGrantArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantReturns(result1 uaa.AccessToken, result2 error) {
	fake.ownerPasswordCredentialsGrantMutex.Lock()
	defer fake.ownerPasswordCredentialsGrantMutex.Unlock()
	fake.OwnerPasswordCredentialsGrantStub = nil
	fake.ownerPasswordCredentialsGrantReturns = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantReturnsOnCall(i int, result1 uaa.AccessToken, result2 error) {
	fake.ownerPasswordCredentialsGrantMutex.Lock()
	defer fake.ownerPasswordCredentialsGrantMutex.Unlock()
	fake.OwnerPasswordCredentialsGrantStub = nil
	if fake.ownerPasswordCredentialsGrantReturnsOnCall == nil {
		fake.ownerPasswordCredentialsGrantReturnsOnCall = make(map[int]struct {
			result1 uaa.AccessToken
			result2 error
		})
	}
	fake.ownerPasswordCredentialsGrantReturnsOnCall[i] = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) RefreshTokenGrant(arg1 string) (uaa.AccessToken, error) {
	fake.refreshTokenGrantMutex.Lock()
	ret, specificReturn := fake.refreshTokenGrantReturnsOnCall[len(fake.refreshTokenGrantArgsForCall)]
	fake.refreshTokenGrantArgsForCall = append(fake.refreshTokenGrantArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.RefreshTokenGrantStub
	fakeReturns := fake.refreshTokenGrantReturns
	fake.recordInvocation("RefreshTokenGrant", []interface{}{arg1})
	fake.refreshTokenGrantMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeUAA) RefreshTokenGrantCallCount() int {
	fake.refreshTokenGrantMutex.RLock()
	defer fake.refreshTokenGrantMutex.RUnlock()
	return len(fake.refreshTokenGrantArgsForCall)
}

func (fake *FakeUAA) RefreshTokenGrantCalls(stub func(string) (uaa.AccessToken, error)) {
	fake.refreshTokenGrantMutex.Lock()
	defer fake.refreshTokenGrantMutex.Unlock()
	fake.RefreshTokenGrantStub = stub
}

func (fake *FakeUAA) RefreshTokenGrantArgsForCall(i int) string {
	fake.refreshTokenGrantMutex.RLock()
	defer fake.refreshTokenGrantMutex.RUnlock()
	argsForCall := fake.refreshTokenGrantArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeUAA) RefreshTokenGrantReturns(result1 uaa.AccessToken, result2 error) {
	fake.refreshTokenGrantMutex.Lock()
	defer fake.refreshTokenGrantMutex.Unlock()
	fake.RefreshTokenGrantStub = nil
	fake.refreshTokenGrantReturns = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) RefreshTokenGrantReturnsOnCall(i int, result1 uaa.AccessToken, result2 error) {
	fake.refreshTokenGrantMutex.Lock()
	defer fake.refreshTokenGrantMutex.Unlock()
	fake.RefreshTokenGrantStub = nil
	if fake.refreshTokenGrantReturnsOnCall == nil {
		fake.refreshTokenGrantReturnsOnCall = make(map[int]struct {
			result1 uaa.AccessToken
			result2 error
		})
	}
	fake.refreshTokenGrantReturnsOnCall[i] = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.clientCredentialsGrantMutex.RLock()
	defer fake.clientCredentialsGrantMutex.RUnlock()
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	fake.refreshTokenGrantMutex.RLock()
	defer fake.refreshTokenGrantMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeUAA) recordInvocation(key string, args []interface{}) {
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

var _ uaa.UAA = new(FakeUAA)
