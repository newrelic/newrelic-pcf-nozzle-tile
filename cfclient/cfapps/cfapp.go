// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cfapps

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
)

var startValue = "WAITING ON DATA"
var cfg = app.Get().Config

// nolint
var (
	AppName             = cfg.GetString(config.EnvAppName)
	AppSpaceName        = cfg.GetString(config.EnvAppSpaceName)
	AppOrgName          = cfg.GetString(config.EnvAppOrgName)
	AppID               = cfg.GetString(config.EnvAppID)
	AppInstanceIndex    = cfg.GetString(config.EnvAppInstanceIndex)
	AppInstanceState    = cfg.GetString(config.EnvAppInstanceState)
	AppInstanceUID      = cfg.GetString(config.EnvAppInstanceUID)
	AppInstancesDesired = cfg.GetString(config.EnvAppInstancesDesired)
)

// CFApp Extended
type CFApp struct {
	Attributes   *attributes.Attributes
	GUID         string
	App          *cfclient.App
	Summaries    map[int32]string
	VcapServices map[string]interface{}
	LastPull     time.Time
	Lock         *sync.RWMutex
	retryCount   int32
}

// NewSummary ...
func NewSummary() *attributes.Attributes {
	return attributes.NewAttributes(
		attributes.New(AppInstancesDesired, startValue),
		attributes.New(AppName, startValue),
		attributes.New(AppSpaceName, startValue),
		attributes.New(AppOrgName, startValue),
		attributes.New(AppInstanceState, startValue),
	)
}

// NewCFApp ...
func NewCFApp(guid string) *CFApp {
	return &CFApp{
		Attributes:   NewSummary(),
		GUID:         guid,
		Summaries:    map[int32]string{},
		VcapServices: map[string]interface{}{},
		LastPull:     time.Now(),
		Lock:         &sync.RWMutex{},
		retryCount:   0,
	}
}

// SummaryByInstance ...
/*
func (a *CFApp) SummaryByInstance(index int32) *attributes.Attributes {
	indexString := strconv.FormatInt(int64(index), 10)
	if attrs, found := a.Summaries[indexString]; found {
		return attrs
	}
	a.Summaries[indexString] = NewSummary(indexString)
	return a.Summaries[indexString]
}
*/

// GetInstanceAttributes ...
func (a *CFApp) GetInstanceAttributes(id int32) (attrs *attributes.Attributes) {
	attrs = attributes.NewAttributes()
	a.Lock.RLock()
	defer a.Lock.RUnlock()
	attrs.AppendAll(a.Attributes)
	if appInstance, found := a.Summaries[id]; found {
		attrs.SetAttribute(AppInstanceState, appInstance)
		if a.App != nil {
			attrs.SetAttribute(AppInstanceUID, fmt.Sprintf("%s:%d", a.App.Name, id))
		}
		return attrs
	}
	attrs.SetAttribute(AppInstanceState, startValue)
	return attrs
}

// GetAttributes returns Attribute struct with methods ...
// func (a *CFApp) GetAttributes() *attributes.Attributes {
// 	return a.Attributes
// }

// UpdateInstances ...
func (a *CFApp) UpdateInstances() {

	defer GetInstance().rateManager.Done()
	if timeout := GetInstance().rateManager.Wait(); timeout != nil {
		app.Get().Log.Errorln("API timeout, app instances failed to update states")
		return
	}

	states, err := GetInstance().GetAppInstances(a.GUID)

	a.Lock.Lock()
	defer a.Lock.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "401 Unauthorized") {
			//401 unauthorized -- token has expired so we need to refresh the client
			app.Get().Log.Warn("cfClient 401 error. Refreshing client due to this error: %s", err.Error())
			go GetInstance().UpdateClient()
		}
		if _, found := a.Summaries[0]; found {
			a.Summaries[0] = err.Error()
		}
		return
	}

	a.Attributes.SetAttribute(AppInstancesDesired, len(states))

	for k, v := range states {
		if index64, err := strconv.ParseInt(k, 10, 32); err == nil {
			index := int32(index64)
			//			attrs := a.GetInstanceAttributes(index)
			//			if a.App != nil {
			//				attrs.SetAttribute(AppInstanceUID, fmt.Sprintf("%s:%s", a.App.Name, k))
			//			}
			a.Summaries[index] = v.State
		} else {
			app.Get().Log.Fatal(err)
		}
	}

}

// GetAppEnv calls the client to get the system environment.  This is added to the pcfapp and
// consumed by applicaton specific accumulators (ContainerMetric and LogMessage)
func (a *CFApp) GetAppEnv() {

	defer GetInstance().rateManager.Done()
	if timeout := GetInstance().rateManager.Wait(); timeout != nil {
		app.Get().Log.Errorln("api timeout, GetAppEnv failed to update")
		return
	}
	env, err := GetInstance().GetAppEnv(a.GUID)
	if err != nil {
		app.Get().Log.Errorf("GetAppEnv failed: %v", err)
		if strings.Contains(err.Error(), "401 Unauthorized") {
			//401 unauthorized -- token has expired so we need to refresh the client
			app.Get().Log.Warn("cfClient 401 error. Refreshing client due to this error: %s", err.Error())
			go GetInstance().UpdateClient()
		}
		return
	}
	a.Lock.Lock()
	a.VcapServices = env.SystemEnv["VCAP_SERVICES"].(map[string]interface{})
	a.Lock.Unlock()
	GetInstance().app.Log.Tracer("V")
}
