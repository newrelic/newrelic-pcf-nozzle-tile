// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package uid

import "fmt"

// ID ...
type ID string

// Concat ...
func (i *ID) Concat(v ...interface{}) {
	for _, v := range v {
		*i = ID(fmt.Sprintf("%s/%s", *i, v))
	}
}

func (i *ID) String() string {
	return string(*i)
}
