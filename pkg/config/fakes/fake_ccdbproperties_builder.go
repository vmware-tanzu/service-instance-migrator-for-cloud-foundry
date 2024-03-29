// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

type FakeCCDBPropertiesBuilder struct {
	BuildStub        func() *config.CCDBProperties
	buildMutex       sync.RWMutex
	buildArgsForCall []struct {
	}
	buildReturns struct {
		result1 *config.CCDBProperties
	}
	buildReturnsOnCall map[int]struct {
		result1 *config.CCDBProperties
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCCDBPropertiesBuilder) Build() *config.CCDBProperties {
	fake.buildMutex.Lock()
	ret, specificReturn := fake.buildReturnsOnCall[len(fake.buildArgsForCall)]
	fake.buildArgsForCall = append(fake.buildArgsForCall, struct {
	}{})
	stub := fake.BuildStub
	fakeReturns := fake.buildReturns
	fake.recordInvocation("Build", []interface{}{})
	fake.buildMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCCDBPropertiesBuilder) BuildCallCount() int {
	fake.buildMutex.RLock()
	defer fake.buildMutex.RUnlock()
	return len(fake.buildArgsForCall)
}

func (fake *FakeCCDBPropertiesBuilder) BuildCalls(stub func() *config.CCDBProperties) {
	fake.buildMutex.Lock()
	defer fake.buildMutex.Unlock()
	fake.BuildStub = stub
}

func (fake *FakeCCDBPropertiesBuilder) BuildReturns(result1 *config.CCDBProperties) {
	fake.buildMutex.Lock()
	defer fake.buildMutex.Unlock()
	fake.BuildStub = nil
	fake.buildReturns = struct {
		result1 *config.CCDBProperties
	}{result1}
}

func (fake *FakeCCDBPropertiesBuilder) BuildReturnsOnCall(i int, result1 *config.CCDBProperties) {
	fake.buildMutex.Lock()
	defer fake.buildMutex.Unlock()
	fake.BuildStub = nil
	if fake.buildReturnsOnCall == nil {
		fake.buildReturnsOnCall = make(map[int]struct {
			result1 *config.CCDBProperties
		})
	}
	fake.buildReturnsOnCall[i] = struct {
		result1 *config.CCDBProperties
	}{result1}
}

func (fake *FakeCCDBPropertiesBuilder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.buildMutex.RLock()
	defer fake.buildMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCCDBPropertiesBuilder) recordInvocation(key string, args []interface{}) {
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

var _ config.CCDBPropertiesBuilder = new(FakeCCDBPropertiesBuilder)
