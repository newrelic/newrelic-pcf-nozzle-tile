// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cfapps

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

type rateManager struct {
	burstLimit   int32
	active       int32
	queued       int32
	delay        time.Duration
	noQueueDelay time.Duration
	timeout      time.Duration
	nextChan     chan bool
}

func newRateManager() *rateManager {
	rm := &rateManager{
		burstLimit:   app.Get().Config.GetInt32("FIREHOSE_RATE_BURST"),
		queued:       0,
		active:       0,
		delay:        50 * time.Millisecond,
		noQueueDelay: time.Second,
		timeout:      app.Get().Config.GetDuration("FIREHOSE_RATE_TIMEOUT_SECS") * time.Second,
		nextChan:     make(chan bool),
	}
	go func() {
		for {
			if atomic.LoadInt32(&rm.active) <= rm.burstLimit {
				atomic.AddInt32(&rm.active, 1)
				rm.nextChan <- true
				continue
			}

			if rm.HasQueue() {
				time.Sleep(rm.delay)
			} else {
				time.Sleep(rm.noQueueDelay)
			}

		}
	}()
	return rm
}

func (b *rateManager) Done() {
	atomic.AddInt32(&b.active, -1)
}

// HasQueue ...
func (b *rateManager) HasQueue() bool {
	return atomic.LoadInt32(&b.queued) > 0
}

func (b *rateManager) Wait() error {

	atomic.AddInt32(&b.queued, 1)
	defer atomic.AddInt32(&b.queued, -1)

	select {

	case <-b.nextChan:
		return nil

	case <-time.After(b.timeout):
		return errors.New("rate-limiter timeout")

	}

}
