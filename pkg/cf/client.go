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

package cf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
)

const (
	// DefaultRetryTimeout sets the amount of time before a retry times out
	DefaultRetryTimeout = time.Minute
	// DefaultRetryPause sets the amount of time to wait before retrying
	DefaultRetryPause = 3 * time.Second
)

var ErrRetry = errors.New("retry")

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakes . Client

type Client interface {
	AppByName(appName, spaceGuid, orgGuid string) (cfclient.App, error)
	CreateApp(req cfclient.AppCreateRequest) (cfclient.App, error)
	DeleteApp(guid string) error
	CreateOrg(req cfclient.OrgRequest) (cfclient.Org, error)
	CreateServiceBinding(appGUID, serviceInstanceGUID string) (*cfclient.ServiceBinding, error)
	CreateServiceInstance(req cfclient.ServiceInstanceRequest) (cfclient.ServiceInstance, error)
	CreateServiceKey(req cfclient.CreateServiceKeyRequest) (cfclient.ServiceKey, error)
	CreateSpace(req cfclient.SpaceRequest) (cfclient.Space, error)
	CreateUserProvidedServiceInstance(req cfclient.UserProvidedServiceInstanceRequest) (*cfclient.UserProvidedServiceInstance, error)
	DeleteOrg(guid string, recursive, async bool) error
	DeleteServiceBinding(guid string) error
	DeleteServiceInstance(guid string, recursive, async bool) error
	GetAppByGuidNoInlineCall(guid string) (cfclient.App, error)
	GetClientConfig() *cfclient.Config
	GetOrgByGuid(guid string) (cfclient.Org, error)
	GetOrgByName(name string) (cfclient.Org, error)
	GetSpaceByGuid(spaceGUID string) (cfclient.Space, error)
	GetSpaceByName(name string, orgGUID string) (cfclient.Space, error)
	GetServiceByGuid(guid string) (cfclient.Service, error)
	GetServicePlanByGUID(guid string) (*cfclient.ServicePlan, error)
	GetServiceInstanceByGuid(guid string) (cfclient.ServiceInstance, error)
	GetServiceInstanceParams(guid string) (map[string]interface{}, error)
	GetUserProvidedServiceInstanceByGuid(guid string) (cfclient.UserProvidedServiceInstance, error)
	ListOrgs() ([]cfclient.Org, error)
	ListUserProvidedServiceInstancesByQuery(query url.Values) ([]cfclient.UserProvidedServiceInstance, error)
	ListServices() ([]cfclient.Service, error)
	ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error)
	ListServiceKeysByQuery(query url.Values) ([]cfclient.ServiceKey, error)
	ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error)
	ListSpaceServiceInstances(spaceGUID string) ([]cfclient.ServiceInstance, error)
	ListServicePlans() ([]cfclient.ServicePlan, error)
	ListSpaces() ([]cfclient.Space, error)
	ListSpacesByQuery(query url.Values) ([]cfclient.Space, error)
	ListServiceBrokers() ([]cfclient.ServiceBroker, error)
	ListServicePlansByQuery(values url.Values) ([]cfclient.ServicePlan, error)
	UpdateUserProvidedServiceInstance(guid string, req cfclient.UserProvidedServiceInstanceRequest) (*cfclient.UserProvidedServiceInstance, error)
	UpdateSI(serviceInstanceGuid string, req cfclient.ServiceInstanceUpdateRequest, async bool) error
	NewRequest(method, path string) *cfclient.Request
	DoRequest(req *cfclient.Request) (*http.Response, error)
	DoWithRetry(f func() error) error
}

type Config struct {
	SSLDisabled  bool
	URL          string
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
}

type ClientImpl struct {
	CachingClient *CachingClient
	RetryPause    time.Duration
	RetryTimeout  time.Duration
	Config        *Config
	cfConfig      *cfclient.Config
	*cfclient.Client
}

func NewClient(c *Config, options ...func(*ClientImpl)) (*ClientImpl, error) {
	client := &ClientImpl{
		Config:       c,
		RetryTimeout: DefaultRetryTimeout,
		RetryPause:   DefaultRetryPause,
	}

	for _, o := range options {
		o(client)
	}

	return client, nil
}

func (c *ClientImpl) lazyLoadCacheClientOrDie() *CachingClient {
	if c.CachingClient == nil {
		cf, err := c.lazyLoadCFClient(c.Config)
		if err != nil {
			panic(err)
		}
		c.CachingClient = NewCachingClient(cf)
	}
	return c.CachingClient
}

func (c *ClientImpl) lazyLoadCFClient(cfg *Config) (*cfclient.Client, error) {
	if c.Client == nil {
		cf, err := cfclient.NewClient(c.lazyLoadClientConfig(cfg))
		if err != nil {
			return nil, err
		}
		c.Client = cf
	}

	return c.Client, nil
}

func (c *ClientImpl) lazyLoadClientConfig(config *Config) *cfclient.Config {
	if c.cfConfig == nil {
		cfg := &cfclient.Config{
			ApiAddress:        config.URL,
			Username:          config.Username,
			Password:          config.Password,
			ClientID:          config.ClientID,
			ClientSecret:      config.ClientSecret,
			SkipSslValidation: true,
		}
		c.cfConfig = cfg
	}

	return c.cfConfig
}

func WithRetryPause(pause time.Duration) func(*ClientImpl) {
	return func(cf *ClientImpl) {
		cf.RetryPause = pause
	}
}

func WithRetryTimeout(timeout time.Duration) func(*ClientImpl) {
	return func(cf *ClientImpl) {
		cf.RetryTimeout = timeout
	}
}

func (c *ClientImpl) GetAppByGuidNoInlineCall(guid string) (cfclient.App, error) {
	return c.lazyLoadCacheClientOrDie().GetAppByGuidNoInlineCall(guid)
}

func (c *ClientImpl) GetClientConfig() *cfclient.Config {
	return c.lazyLoadClientConfig(c.Config)
}

func (c *ClientImpl) AppByName(appName, spaceGuid, orgGuid string) (cfclient.App, error) {
	return c.lazyLoadCacheClientOrDie().AppByName(appName, spaceGuid, orgGuid)
}

func (c *ClientImpl) CreateOrg(request cfclient.OrgRequest) (cfclient.Org, error) {
	return c.lazyLoadCacheClientOrDie().CreateOrg(request)
}

func (c *ClientImpl) GetOrgByGuid(guid string) (cfclient.Org, error) {
	return c.lazyLoadCacheClientOrDie().GetOrgByGuid(guid)
}

func (c *ClientImpl) GetOrgByName(name string) (cfclient.Org, error) {
	return c.lazyLoadCacheClientOrDie().GetOrgByName(name)
}

func (c *ClientImpl) GetSpaceByGuid(spaceGUID string) (cfclient.Space, error) {
	return c.lazyLoadCacheClientOrDie().GetSpaceByGuid(spaceGUID)
}

func (c *ClientImpl) GetSpaceByName(name string, orgGUID string) (cfclient.Space, error) {
	return c.lazyLoadCacheClientOrDie().GetSpaceByName(name, orgGUID)
}

func (c *ClientImpl) GetServiceByGuid(guid string) (cfclient.Service, error) {
	return c.lazyLoadCacheClientOrDie().GetServiceByGuid(guid)
}

func (c *ClientImpl) GetServicePlanByGUID(guid string) (*cfclient.ServicePlan, error) {
	return c.lazyLoadCacheClientOrDie().GetServicePlanByGUID(guid)
}

func (c *ClientImpl) GetServiceInstanceByGuid(guid string) (cfclient.ServiceInstance, error) {
	return c.lazyLoadCacheClientOrDie().GetServiceInstanceByGuid(guid)
}

func (c *ClientImpl) GetServiceInstanceParams(guid string) (map[string]interface{}, error) {
	return c.lazyLoadCacheClientOrDie().GetServiceInstanceParams(guid)
}

func (c *ClientImpl) ListOrgs() ([]cfclient.Org, error) {
	return c.lazyLoadCacheClientOrDie().ListOrgs()
}

func (c *ClientImpl) ListUserProvidedServiceInstancesByQuery(query url.Values) ([]cfclient.UserProvidedServiceInstance, error) {
	return c.lazyLoadCacheClientOrDie().ListUserProvidedServiceInstancesByQuery(query)
}

func (c *ClientImpl) ListServices() ([]cfclient.Service, error) {
	return c.lazyLoadCacheClientOrDie().ListServices()
}

func (c *ClientImpl) ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error) {
	return c.lazyLoadCacheClientOrDie().ListServiceBindingsByQuery(query)
}

func (c *ClientImpl) ListSpaceServiceInstances(spaceGUID string) ([]cfclient.ServiceInstance, error) {
	return c.lazyLoadCacheClientOrDie().ListSpaceServiceInstances(spaceGUID)
}

func (c *ClientImpl) ListServiceKeysByQuery(query url.Values) ([]cfclient.ServiceKey, error) {
	return c.lazyLoadCacheClientOrDie().ListServiceKeysByQuery(query)
}

func (c *ClientImpl) ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error) {
	return c.lazyLoadCacheClientOrDie().ListServiceInstancesByQuery(query)
}

func (c *ClientImpl) ListServicePlans() ([]cfclient.ServicePlan, error) {
	return c.lazyLoadCacheClientOrDie().ListServicePlans()
}

func (c *ClientImpl) ListSpaces() ([]cfclient.Space, error) {
	return c.lazyLoadCacheClientOrDie().ListSpaces()
}

func (c *ClientImpl) ListSpacesByQuery(query url.Values) ([]cfclient.Space, error) {
	var spaces []cfclient.Space
	err := c.DoWithRetry(func() error {
		var err error
		spaces, err = c.lazyLoadCacheClientOrDie().ListSpacesByQuery(query)
		if err != nil {
			cfErr := cfclient.CloudFoundryHTTPError{}
			if errors.As(err, &cfErr) {
				if cfErr.StatusCode >= 500 && cfErr.StatusCode <= 599 {
					return ErrRetry
				}
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return spaces, nil
}

func (c *ClientImpl) ListServiceBrokers() ([]cfclient.ServiceBroker, error) {
	return c.lazyLoadCacheClientOrDie().ListServiceBrokers()
}

func (c *ClientImpl) ListServicePlansByQuery(values url.Values) ([]cfclient.ServicePlan, error) {
	return c.lazyLoadCacheClientOrDie().ListServicePlansByQuery(values)
}

func (c *ClientImpl) NewRequest(method, path string) *cfclient.Request {
	return c.lazyLoadCacheClientOrDie().NewRequest(method, path)
}

func (c *ClientImpl) DoRequest(req *cfclient.Request) (*http.Response, error) {
	return c.lazyLoadCacheClientOrDie().DoRequest(req)
}

func (c *ClientImpl) DoWithRetry(f func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.RetryTimeout)
	defer cancel()

	for {
		err := f()
		if err == nil {
			return nil
		}

		dnsErr := &net.DNSError{}
		ok := errors.As(err, &dnsErr)
		if ok || errors.Is(err, ErrRetry) {
			select {
			case <-time.After(c.RetryPause):
				continue
			case <-ctx.Done():
				return fmt.Errorf("timed out retrying operation: %w", err)
			}
		}

		return err
	}
}
