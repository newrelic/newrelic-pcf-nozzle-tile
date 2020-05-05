// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// None ...
var None = NewAttributes()

// Attribute ...
type Attribute struct {
	name  string
	value interface{}
}

func (a *Attribute) String() string {
	return fmt.Sprintf("{%s:%v}", a.name, a.value)
}

// Name ...
func (a *Attribute) Name() string {
	return a.name
}

// New ...
func New(name string, value interface{}) *Attribute {
	return &Attribute{
		name:  name,
		value: value,
	}
}

// Value ...
func (a *Attribute) Value() interface{} {
	if reflect.ValueOf(a.value).Kind() == reflect.Ptr {
		value := reflect.ValueOf(a.value).Elem()
		return reflect.Indirect(value).Interface()
	}
	return a.value
}

// Contains ...
func (a *Attribute) Contains(s string) bool {
	return strings.Contains(a.Name(), s)
}

// Attributes key value map
// singature will be set and static on first call of Singature()
type Attributes struct {
	signature uid.ID
	Map       map[string]*Attribute
	sync      *sync.RWMutex
}

// SetAttribute ...
func (a *Attributes) SetAttribute(name string, v interface{}) *Attribute {
	attr := New(name, v)
	a.sync.Lock()
	defer a.sync.Unlock()

	a.Map[name] = attr
	return attr
}

// ForEach iterates attributes with callback function
func (a *Attributes) ForEach(fn func(attr *Attribute)) {
	a.sync.Lock()
	for _, attr := range a.Map {
		fn(attr)
	}
	a.sync.Unlock()
}

// Length of attribute collection size
func (a *Attributes) Length() int {
	return len(a.Map)
}

// Has returns index or false in bool
func (a *Attributes) Has(name string) (attr *Attribute) {
	a.sync.RLock()
	defer a.sync.RUnlock()
	attr, _ = a.Map[name]
	return attr
}

// NewAttributes ...
func NewAttributes(list ...*Attribute) *Attributes {
	attributes := Attributes{
		Map:  map[string]*Attribute{},
		sync: &sync.RWMutex{},
	}
	for _, attr := range list {
		attributes.Append(attr)
	}
	return &attributes
}

// Append ...
// TODO: why is this not setting sync?
func (a *Attributes) Append(v *Attribute) {
	if attr := a.Has(v.Name()); attr != nil {
		return
	}
	a.sync.Lock()
	a.Map[v.Name()] = v
	a.sync.Unlock()
}

// AppendAll ...
func (a *Attributes) AppendAll(other *Attributes) {
	other.ForEach(func(attr *Attribute) {
		a.Append(attr)
	})
}

// AttributeByName ... can panic if attribute not found!
func (a *Attributes) AttributeByName(name string) (attr *Attribute) {
	a.sync.RLock()
	defer a.sync.RUnlock()
	attr, _ = a.Map[name]
	return
}

// Get ...
func (a *Attributes) Get() []*Attribute {
	attrs := make([]*Attribute, 0, len(a.Map))
	for _, val := range a.Map {
		attrs = append(attrs, val)
	}
	return attrs
}

// FloatValueOf ... TODO: fix thread safe
func (a *Attributes) FloatValueOf(name string) float64 {
	if attr, found := a.Map[name]; found {
		i := reflect.TypeOf(attr.value)
		if i.ConvertibleTo(reflect.TypeOf(float64(0))) {
			v := reflect.ValueOf(attr.value)
			f := v.Convert(reflect.TypeOf(float64(0)))
			return f.Float()
		}
	}
	return 0
}

// Signature ...
// first call sets signature, new attributes added
// after first call will not change signature
func (a *Attributes) Signature() uid.ID {
	if len(a.signature) > 0 {
		return a.signature
	}
	a.sync.Lock()
	defer a.sync.Unlock()
	keys := a.sortedKeys()
	for _, k := range keys {
		a.signature.Concat(a.Map[k].Value())
	}
	return a.signature
}

func (a *Attributes) sortedKeys() []string {
	keys := make([]string, len(a.Map))
	i := 0
	for k := range a.Map {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// Marshal ...
func (a *Attributes) Marshal() map[string]interface{} {
	a.sync.RLock()
	defer a.sync.RUnlock()
	result := map[string]interface{}{}
	for name, attr := range a.Map {
		result[name] = attr.Value()
	}
	return result
}
