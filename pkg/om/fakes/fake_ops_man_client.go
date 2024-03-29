// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om"
	"github.com/vmware-tanzu/service-instance-migrator-for-cloud-foundry/pkg/om/httpclient"
)

type FakeOpsManClient struct {
	GetBOSHCredentialsStub        func() (string, error)
	getBOSHCredentialsMutex       sync.RWMutex
	getBOSHCredentialsArgsForCall []struct {
	}
	getBOSHCredentialsReturns struct {
		result1 string
		result2 error
	}
	getBOSHCredentialsReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	GetStagedProductPropertiesStub        func(string) (map[string]httpclient.ResponseProperty, error)
	getStagedProductPropertiesMutex       sync.RWMutex
	getStagedProductPropertiesArgsForCall []struct {
		arg1 string
	}
	getStagedProductPropertiesReturns struct {
		result1 map[string]httpclient.ResponseProperty
		result2 error
	}
	getStagedProductPropertiesReturnsOnCall map[int]struct {
		result1 map[string]httpclient.ResponseProperty
		result2 error
	}
	ListCertificateAuthoritiesStub        func() ([]httpclient.CA, error)
	listCertificateAuthoritiesMutex       sync.RWMutex
	listCertificateAuthoritiesArgsForCall []struct {
	}
	listCertificateAuthoritiesReturns struct {
		result1 []httpclient.CA
		result2 error
	}
	listCertificateAuthoritiesReturnsOnCall map[int]struct {
		result1 []httpclient.CA
		result2 error
	}
	ListDeployedProductCredentialsStub        func(string, string) (httpclient.DeployedProductCredential, error)
	listDeployedProductCredentialsMutex       sync.RWMutex
	listDeployedProductCredentialsArgsForCall []struct {
		arg1 string
		arg2 string
	}
	listDeployedProductCredentialsReturns struct {
		result1 httpclient.DeployedProductCredential
		result2 error
	}
	listDeployedProductCredentialsReturnsOnCall map[int]struct {
		result1 httpclient.DeployedProductCredential
		result2 error
	}
	ListDeployedProductsStub        func() ([]httpclient.DeployedProduct, error)
	listDeployedProductsMutex       sync.RWMutex
	listDeployedProductsArgsForCall []struct {
	}
	listDeployedProductsReturns struct {
		result1 []httpclient.DeployedProduct
		result2 error
	}
	listDeployedProductsReturnsOnCall map[int]struct {
		result1 []httpclient.DeployedProduct
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOpsManClient) GetBOSHCredentials() (string, error) {
	fake.getBOSHCredentialsMutex.Lock()
	ret, specificReturn := fake.getBOSHCredentialsReturnsOnCall[len(fake.getBOSHCredentialsArgsForCall)]
	fake.getBOSHCredentialsArgsForCall = append(fake.getBOSHCredentialsArgsForCall, struct {
	}{})
	stub := fake.GetBOSHCredentialsStub
	fakeReturns := fake.getBOSHCredentialsReturns
	fake.recordInvocation("GetBOSHCredentials", []interface{}{})
	fake.getBOSHCredentialsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManClient) GetBOSHCredentialsCallCount() int {
	fake.getBOSHCredentialsMutex.RLock()
	defer fake.getBOSHCredentialsMutex.RUnlock()
	return len(fake.getBOSHCredentialsArgsForCall)
}

func (fake *FakeOpsManClient) GetBOSHCredentialsCalls(stub func() (string, error)) {
	fake.getBOSHCredentialsMutex.Lock()
	defer fake.getBOSHCredentialsMutex.Unlock()
	fake.GetBOSHCredentialsStub = stub
}

func (fake *FakeOpsManClient) GetBOSHCredentialsReturns(result1 string, result2 error) {
	fake.getBOSHCredentialsMutex.Lock()
	defer fake.getBOSHCredentialsMutex.Unlock()
	fake.GetBOSHCredentialsStub = nil
	fake.getBOSHCredentialsReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) GetBOSHCredentialsReturnsOnCall(i int, result1 string, result2 error) {
	fake.getBOSHCredentialsMutex.Lock()
	defer fake.getBOSHCredentialsMutex.Unlock()
	fake.GetBOSHCredentialsStub = nil
	if fake.getBOSHCredentialsReturnsOnCall == nil {
		fake.getBOSHCredentialsReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getBOSHCredentialsReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) GetStagedProductProperties(arg1 string) (map[string]httpclient.ResponseProperty, error) {
	fake.getStagedProductPropertiesMutex.Lock()
	ret, specificReturn := fake.getStagedProductPropertiesReturnsOnCall[len(fake.getStagedProductPropertiesArgsForCall)]
	fake.getStagedProductPropertiesArgsForCall = append(fake.getStagedProductPropertiesArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetStagedProductPropertiesStub
	fakeReturns := fake.getStagedProductPropertiesReturns
	fake.recordInvocation("GetStagedProductProperties", []interface{}{arg1})
	fake.getStagedProductPropertiesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManClient) GetStagedProductPropertiesCallCount() int {
	fake.getStagedProductPropertiesMutex.RLock()
	defer fake.getStagedProductPropertiesMutex.RUnlock()
	return len(fake.getStagedProductPropertiesArgsForCall)
}

func (fake *FakeOpsManClient) GetStagedProductPropertiesCalls(stub func(string) (map[string]httpclient.ResponseProperty, error)) {
	fake.getStagedProductPropertiesMutex.Lock()
	defer fake.getStagedProductPropertiesMutex.Unlock()
	fake.GetStagedProductPropertiesStub = stub
}

func (fake *FakeOpsManClient) GetStagedProductPropertiesArgsForCall(i int) string {
	fake.getStagedProductPropertiesMutex.RLock()
	defer fake.getStagedProductPropertiesMutex.RUnlock()
	argsForCall := fake.getStagedProductPropertiesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOpsManClient) GetStagedProductPropertiesReturns(result1 map[string]httpclient.ResponseProperty, result2 error) {
	fake.getStagedProductPropertiesMutex.Lock()
	defer fake.getStagedProductPropertiesMutex.Unlock()
	fake.GetStagedProductPropertiesStub = nil
	fake.getStagedProductPropertiesReturns = struct {
		result1 map[string]httpclient.ResponseProperty
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) GetStagedProductPropertiesReturnsOnCall(i int, result1 map[string]httpclient.ResponseProperty, result2 error) {
	fake.getStagedProductPropertiesMutex.Lock()
	defer fake.getStagedProductPropertiesMutex.Unlock()
	fake.GetStagedProductPropertiesStub = nil
	if fake.getStagedProductPropertiesReturnsOnCall == nil {
		fake.getStagedProductPropertiesReturnsOnCall = make(map[int]struct {
			result1 map[string]httpclient.ResponseProperty
			result2 error
		})
	}
	fake.getStagedProductPropertiesReturnsOnCall[i] = struct {
		result1 map[string]httpclient.ResponseProperty
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListCertificateAuthorities() ([]httpclient.CA, error) {
	fake.listCertificateAuthoritiesMutex.Lock()
	ret, specificReturn := fake.listCertificateAuthoritiesReturnsOnCall[len(fake.listCertificateAuthoritiesArgsForCall)]
	fake.listCertificateAuthoritiesArgsForCall = append(fake.listCertificateAuthoritiesArgsForCall, struct {
	}{})
	stub := fake.ListCertificateAuthoritiesStub
	fakeReturns := fake.listCertificateAuthoritiesReturns
	fake.recordInvocation("ListCertificateAuthorities", []interface{}{})
	fake.listCertificateAuthoritiesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManClient) ListCertificateAuthoritiesCallCount() int {
	fake.listCertificateAuthoritiesMutex.RLock()
	defer fake.listCertificateAuthoritiesMutex.RUnlock()
	return len(fake.listCertificateAuthoritiesArgsForCall)
}

func (fake *FakeOpsManClient) ListCertificateAuthoritiesCalls(stub func() ([]httpclient.CA, error)) {
	fake.listCertificateAuthoritiesMutex.Lock()
	defer fake.listCertificateAuthoritiesMutex.Unlock()
	fake.ListCertificateAuthoritiesStub = stub
}

func (fake *FakeOpsManClient) ListCertificateAuthoritiesReturns(result1 []httpclient.CA, result2 error) {
	fake.listCertificateAuthoritiesMutex.Lock()
	defer fake.listCertificateAuthoritiesMutex.Unlock()
	fake.ListCertificateAuthoritiesStub = nil
	fake.listCertificateAuthoritiesReturns = struct {
		result1 []httpclient.CA
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListCertificateAuthoritiesReturnsOnCall(i int, result1 []httpclient.CA, result2 error) {
	fake.listCertificateAuthoritiesMutex.Lock()
	defer fake.listCertificateAuthoritiesMutex.Unlock()
	fake.ListCertificateAuthoritiesStub = nil
	if fake.listCertificateAuthoritiesReturnsOnCall == nil {
		fake.listCertificateAuthoritiesReturnsOnCall = make(map[int]struct {
			result1 []httpclient.CA
			result2 error
		})
	}
	fake.listCertificateAuthoritiesReturnsOnCall[i] = struct {
		result1 []httpclient.CA
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListDeployedProductCredentials(arg1 string, arg2 string) (httpclient.DeployedProductCredential, error) {
	fake.listDeployedProductCredentialsMutex.Lock()
	ret, specificReturn := fake.listDeployedProductCredentialsReturnsOnCall[len(fake.listDeployedProductCredentialsArgsForCall)]
	fake.listDeployedProductCredentialsArgsForCall = append(fake.listDeployedProductCredentialsArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.ListDeployedProductCredentialsStub
	fakeReturns := fake.listDeployedProductCredentialsReturns
	fake.recordInvocation("ListDeployedProductCredentials", []interface{}{arg1, arg2})
	fake.listDeployedProductCredentialsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManClient) ListDeployedProductCredentialsCallCount() int {
	fake.listDeployedProductCredentialsMutex.RLock()
	defer fake.listDeployedProductCredentialsMutex.RUnlock()
	return len(fake.listDeployedProductCredentialsArgsForCall)
}

func (fake *FakeOpsManClient) ListDeployedProductCredentialsCalls(stub func(string, string) (httpclient.DeployedProductCredential, error)) {
	fake.listDeployedProductCredentialsMutex.Lock()
	defer fake.listDeployedProductCredentialsMutex.Unlock()
	fake.ListDeployedProductCredentialsStub = stub
}

func (fake *FakeOpsManClient) ListDeployedProductCredentialsArgsForCall(i int) (string, string) {
	fake.listDeployedProductCredentialsMutex.RLock()
	defer fake.listDeployedProductCredentialsMutex.RUnlock()
	argsForCall := fake.listDeployedProductCredentialsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeOpsManClient) ListDeployedProductCredentialsReturns(result1 httpclient.DeployedProductCredential, result2 error) {
	fake.listDeployedProductCredentialsMutex.Lock()
	defer fake.listDeployedProductCredentialsMutex.Unlock()
	fake.ListDeployedProductCredentialsStub = nil
	fake.listDeployedProductCredentialsReturns = struct {
		result1 httpclient.DeployedProductCredential
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListDeployedProductCredentialsReturnsOnCall(i int, result1 httpclient.DeployedProductCredential, result2 error) {
	fake.listDeployedProductCredentialsMutex.Lock()
	defer fake.listDeployedProductCredentialsMutex.Unlock()
	fake.ListDeployedProductCredentialsStub = nil
	if fake.listDeployedProductCredentialsReturnsOnCall == nil {
		fake.listDeployedProductCredentialsReturnsOnCall = make(map[int]struct {
			result1 httpclient.DeployedProductCredential
			result2 error
		})
	}
	fake.listDeployedProductCredentialsReturnsOnCall[i] = struct {
		result1 httpclient.DeployedProductCredential
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListDeployedProducts() ([]httpclient.DeployedProduct, error) {
	fake.listDeployedProductsMutex.Lock()
	ret, specificReturn := fake.listDeployedProductsReturnsOnCall[len(fake.listDeployedProductsArgsForCall)]
	fake.listDeployedProductsArgsForCall = append(fake.listDeployedProductsArgsForCall, struct {
	}{})
	stub := fake.ListDeployedProductsStub
	fakeReturns := fake.listDeployedProductsReturns
	fake.recordInvocation("ListDeployedProducts", []interface{}{})
	fake.listDeployedProductsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManClient) ListDeployedProductsCallCount() int {
	fake.listDeployedProductsMutex.RLock()
	defer fake.listDeployedProductsMutex.RUnlock()
	return len(fake.listDeployedProductsArgsForCall)
}

func (fake *FakeOpsManClient) ListDeployedProductsCalls(stub func() ([]httpclient.DeployedProduct, error)) {
	fake.listDeployedProductsMutex.Lock()
	defer fake.listDeployedProductsMutex.Unlock()
	fake.ListDeployedProductsStub = stub
}

func (fake *FakeOpsManClient) ListDeployedProductsReturns(result1 []httpclient.DeployedProduct, result2 error) {
	fake.listDeployedProductsMutex.Lock()
	defer fake.listDeployedProductsMutex.Unlock()
	fake.ListDeployedProductsStub = nil
	fake.listDeployedProductsReturns = struct {
		result1 []httpclient.DeployedProduct
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) ListDeployedProductsReturnsOnCall(i int, result1 []httpclient.DeployedProduct, result2 error) {
	fake.listDeployedProductsMutex.Lock()
	defer fake.listDeployedProductsMutex.Unlock()
	fake.ListDeployedProductsStub = nil
	if fake.listDeployedProductsReturnsOnCall == nil {
		fake.listDeployedProductsReturnsOnCall = make(map[int]struct {
			result1 []httpclient.DeployedProduct
			result2 error
		})
	}
	fake.listDeployedProductsReturnsOnCall[i] = struct {
		result1 []httpclient.DeployedProduct
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getBOSHCredentialsMutex.RLock()
	defer fake.getBOSHCredentialsMutex.RUnlock()
	fake.getStagedProductPropertiesMutex.RLock()
	defer fake.getStagedProductPropertiesMutex.RUnlock()
	fake.listCertificateAuthoritiesMutex.RLock()
	defer fake.listCertificateAuthoritiesMutex.RUnlock()
	fake.listDeployedProductCredentialsMutex.RLock()
	defer fake.listDeployedProductCredentialsMutex.RUnlock()
	fake.listDeployedProductsMutex.RLock()
	defer fake.listDeployedProductsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeOpsManClient) recordInvocation(key string, args []interface{}) {
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

var _ om.OpsManClient = new(FakeOpsManClient)
