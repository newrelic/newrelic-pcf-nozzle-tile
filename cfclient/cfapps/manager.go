// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cfapps

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
)

// Error ...
type Error string

func (e Error) Error() string {
	return string(e)
}

var instance *CFAppManager
var once sync.Once

// CFAppManager ...
type CFAppManager struct {
	app         *app.Application
	client      *cfclient.Client
	clientLock  *sync.RWMutex
	Cache       *Cache
	rateManager *rateManager
	closeChan   chan bool
}

// Start CFAppManager
func Start(app *app.Application) *CFAppManager {

	once.Do(func() {
		app.Log.Info("started CFAppManager")
		instance = &CFAppManager{
			app:         app,
			client:      newClient(app),
			clientLock:  &sync.RWMutex{},
			Cache:       NewCache(),
			rateManager: newRateManager(),
		}
	})
	return instance
}

// GetInstance gets singleton of CFAppManager
func GetInstance() *CFAppManager {
	return instance
}

// GetAppInstanceAttributes ...
func (c *CFAppManager) GetAppInstanceAttributes(appID string, instanceID int32) (attrs *attributes.Attributes) {
	return c.GetApp(appID).GetInstanceAttributes(instanceID)
}

// GetApp ...
func (c *CFAppManager) GetApp(guid string) (app *CFApp) {
	var found bool
	if app, found = c.Cache.Get(guid); found {
		return app
	}
	app = NewCFApp(guid)
	c.app.Log.Debug("Adding new app: ", guid)
	c.Cache.Put(app)
	return app
}

func (c *CFAppManager) updateAppAsync(app *CFApp) {
	go func() {
		if err := c.FetchApp(app); err != nil {
			if atomic.LoadInt32(&app.retryCount) > 2 {
				c.app.Log.Warn("Max retries trying to fetch app: ", app.GUID)
				return
			}
			atomic.AddInt32(&app.retryCount, 1)
			c.updateAppAsync(app)
		} else {
			atomic.StoreInt32(&app.retryCount, 0)
		}
	}()

}

// GetAppInstances ...
func (c *CFAppManager) GetAppInstances(guid string) (map[string]cfclient.AppInstance, error) {
	c.clientLock.RLock()
	defer c.clientLock.RUnlock()
	return c.client.GetAppInstances(guid)
}

// GetAppEnv ...
func (c *CFAppManager) GetAppEnv(guid string) (cfclient.AppEnv, error) {
	c.clientLock.RLock()
	defer c.clientLock.RUnlock()
	return c.client.GetAppEnv(guid)
}

// FetchApp ...
func (c *CFAppManager) FetchApp(a *CFApp) error {

	c.app.Log.Tracer("å")

	defer c.rateManager.Done()
	if timeout := c.rateManager.Wait(); timeout != nil {
		err := errors.New("timeout on update container app details: " + a.GUID)
		c.app.Log.Warn(err)
		return err
	}

	c.clientLock.RLock()
	result, err := c.client.GetAppByGuidNoInlineCall(a.GUID)
	c.clientLock.RUnlock()
	c.app.Log.Tracer("^")

	if err != nil {
		err = fmt.Errorf("CF api error %s on GUID %s", err.Error(), a.GUID)
		c.app.Log.Warn(err)
		if isAuthError(err) {
			//token is no longer valid (expired, or signature can't be verified e.g. after UAA key rotation) -- refresh the client
			app.Get().Log.Warnf("cfClient auth error. Refreshing client due to this error: %s", err.Error())
			go c.UpdateClient()
		}
		return err
	}

	a.Lock.Lock()
	defer a.Lock.Unlock()

	a.App = &result

	a.Attributes.SetAttribute(AppInstancesDesired, result.Instances)
	a.Attributes.SetAttribute(AppName, result.Name)
	a.Attributes.SetAttribute(AppOrgName, result.SpaceData.Entity.OrgData.Entity.Name)
	a.Attributes.SetAttribute(AppSpaceName, result.SpaceData.Entity.Name)

	a.LastPull = time.Now()

	c.app.Log.Tracer("Å")

	// need these requests in the back of the stack
	go a.UpdateInstances()
	go a.GetAppEnv()

	return nil

}

// Close CFAppManager
func (c *CFAppManager) Close() {
	c.closeChan <- true
	c.closeChan <- true
	c.app.Log.Info("closed CFAppManager")
}

func newClient(app *app.Application) *cfclient.Client {
	config := &cfclient.Config{
		ApiAddress:        app.Config.GetString("CF_API_URL"),
		Username:          app.Config.GetString("CF_API_USERNAME"),
		Password:          app.Config.GetString("CF_API_PASSWORD"),
		SkipSslValidation: app.Config.GetBool("CF_SKIP_SSL"),
	}

	client, err := cfclient.NewClient(config)
	if err != nil {
		app.Log.Fatalf("unable to connect to cf-client: %s", err.Error())
	}
	return client
}

// isAuthError reports whether err indicates the cached CF API token is no longer usable.
// Besides a plain 401, UAA signing-key rotation surfaces as an oauth2 token error (e.g.
// `oauth2: "invalid_token" "Could not verify token signature."`) when the client silently
// tries to refresh a token signed under a key UAA no longer trusts.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "401 Unauthorized") || strings.Contains(msg, "oauth2:")
}

// UpdateClient when the token expires we need a fresh client
func (c *CFAppManager) UpdateClient() {
	c.clientLock.Lock()
	defer c.clientLock.Unlock()

	client := newClient(app.Get())
	c.client = client
}
