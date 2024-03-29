// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/cf"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/config"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/migrate"
)

type FakeFactory struct {
	NewStub        func(string, string, *cf.ServiceInstance, config.OpsManager, config.Loader, migrate.ClientHolder, string, bool) (migrate.ServiceInstanceMigrator, error)
	newMutex       sync.RWMutex
	newArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 *cf.ServiceInstance
		arg4 config.OpsManager
		arg5 config.Loader
		arg6 migrate.ClientHolder
		arg7 string
		arg8 bool
	}
	newReturns struct {
		result1 migrate.ServiceInstanceMigrator
		result2 error
	}
	newReturnsOnCall map[int]struct {
		result1 migrate.ServiceInstanceMigrator
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeFactory) New(arg1 string, arg2 string, arg3 *cf.ServiceInstance, arg4 config.OpsManager, arg5 config.Loader, arg6 migrate.ClientHolder, arg7 string, arg8 bool) (migrate.ServiceInstanceMigrator, error) {
	fake.newMutex.Lock()
	ret, specificReturn := fake.newReturnsOnCall[len(fake.newArgsForCall)]
	fake.newArgsForCall = append(fake.newArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 *cf.ServiceInstance
		arg4 config.OpsManager
		arg5 config.Loader
		arg6 migrate.ClientHolder
		arg7 string
		arg8 bool
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	stub := fake.NewStub
	fakeReturns := fake.newReturns
	fake.recordInvocation("New", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	fake.newMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeFactory) NewCallCount() int {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	return len(fake.newArgsForCall)
}

func (fake *FakeFactory) NewCalls(stub func(string, string, *cf.ServiceInstance, config.OpsManager, config.Loader, migrate.ClientHolder, string, bool) (migrate.ServiceInstanceMigrator, error)) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = stub
}

func (fake *FakeFactory) NewArgsForCall(i int) (string, string, *cf.ServiceInstance, config.OpsManager, config.Loader, migrate.ClientHolder, string, bool) {
	fake.newMutex.RLock()
	defer fake.newMutex.RUnlock()
	argsForCall := fake.newArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7, argsForCall.arg8
}

func (fake *FakeFactory) NewReturns(result1 migrate.ServiceInstanceMigrator, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	fake.newReturns = struct {
		result1 migrate.ServiceInstanceMigrator
		result2 error
	}{result1, result2}
}

func (fake *FakeFactory) NewReturnsOnCall(i int, result1 migrate.ServiceInstanceMigrator, result2 error) {
	fake.newMutex.Lock()
	defer fake.newMutex.Unlock()
	fake.NewStub = nil
	if fake.newReturnsOnCall == nil {
		fake.newReturnsOnCall = make(map[int]struct {
			result1 migrate.ServiceInstanceMigrator
			result2 error
		})
	}
	fake.newReturnsOnCall[i] = struct {
		result1 migrate.ServiceInstanceMigrator
		result2 error
	}{result1, result2}
}

func (fake *FakeFactory) Invocations() map[string][][]interface{} {
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

func (fake *FakeFactory) recordInvocation(key string, args []interface{}) {
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

var _ migrate.Factory = new(FakeFactory)
