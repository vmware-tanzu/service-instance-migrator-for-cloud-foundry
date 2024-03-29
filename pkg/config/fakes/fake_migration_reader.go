// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
)

type FakeMigrationReader struct {
	GetMigrationStub        func() (*config.Migration, error)
	getMigrationMutex       sync.RWMutex
	getMigrationArgsForCall []struct {
	}
	getMigrationReturns struct {
		result1 *config.Migration
		result2 error
	}
	getMigrationReturnsOnCall map[int]struct {
		result1 *config.Migration
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMigrationReader) GetMigration() (*config.Migration, error) {
	fake.getMigrationMutex.Lock()
	ret, specificReturn := fake.getMigrationReturnsOnCall[len(fake.getMigrationArgsForCall)]
	fake.getMigrationArgsForCall = append(fake.getMigrationArgsForCall, struct {
	}{})
	stub := fake.GetMigrationStub
	fakeReturns := fake.getMigrationReturns
	fake.recordInvocation("GetMigration", []interface{}{})
	fake.getMigrationMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeMigrationReader) GetMigrationCallCount() int {
	fake.getMigrationMutex.RLock()
	defer fake.getMigrationMutex.RUnlock()
	return len(fake.getMigrationArgsForCall)
}

func (fake *FakeMigrationReader) GetMigrationCalls(stub func() (*config.Migration, error)) {
	fake.getMigrationMutex.Lock()
	defer fake.getMigrationMutex.Unlock()
	fake.GetMigrationStub = stub
}

func (fake *FakeMigrationReader) GetMigrationReturns(result1 *config.Migration, result2 error) {
	fake.getMigrationMutex.Lock()
	defer fake.getMigrationMutex.Unlock()
	fake.GetMigrationStub = nil
	fake.getMigrationReturns = struct {
		result1 *config.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrationReader) GetMigrationReturnsOnCall(i int, result1 *config.Migration, result2 error) {
	fake.getMigrationMutex.Lock()
	defer fake.getMigrationMutex.Unlock()
	fake.GetMigrationStub = nil
	if fake.getMigrationReturnsOnCall == nil {
		fake.getMigrationReturnsOnCall = make(map[int]struct {
			result1 *config.Migration
			result2 error
		})
	}
	fake.getMigrationReturnsOnCall[i] = struct {
		result1 *config.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrationReader) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getMigrationMutex.RLock()
	defer fake.getMigrationMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMigrationReader) recordInvocation(key string, args []interface{}) {
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

var _ config.MigrationReader = new(FakeMigrationReader)
