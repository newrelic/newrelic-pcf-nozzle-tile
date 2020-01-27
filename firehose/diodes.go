// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package firehose

import (
	"code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

// OneToOneEnvelope ...
type OneToOneEnvelope struct {
	d *diodes.Poller
}

// NewOneToOneEnvelope ...
func NewOneToOneEnvelope(size int, alerter diodes.Alerter) *OneToOneEnvelope {
	return &OneToOneEnvelope{
		d: diodes.NewPoller(diodes.NewManyToOne(size, alerter)),
	}
}

// Set inserts the given V2 envelope into the diode.
func (d *OneToOneEnvelope) Set(data *loggregator_v2.Envelope) {
	d.d.Set(diodes.GenericDataType(data))
}

// TryNext returns the next V2 envelope to be read from the diode.  If the
// diode is empty it will return a nil envelope and false for the bool.
func (d *OneToOneEnvelope) TryNext() (*loggregator_v2.Envelope, bool) {
	if data, notEmpty := d.d.TryNext(); notEmpty {
		return (*loggregator_v2.Envelope)(data), true
	}
	return nil, false
}

// Next will return the next V2 envelope to be read from the diode. If the
// diode is empty this method will block until an envelope is available to be
// read.
func (d *OneToOneEnvelope) Next() *loggregator_v2.Envelope {
	data := d.d.Next()
	return (*loggregator_v2.Envelope)(data)
}
