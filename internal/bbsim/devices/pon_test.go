/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package devices

import (
	"github.com/opencord/voltha-protos/v4/go/openolt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var sn1 = NewSN(0, 0, 1)
var sn2 = NewSN(0, 0, 2)
var sn3 = NewSN(0, 0, 3)

// TODO run this test from the makefile!!

// NOTE that we are using a benchmark test to actually test concurrency
func Benchmark_storeGemPort(b *testing.B) {
	pon := PonPort{
		allocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	// concurrently add multiple ports
	go func(wg *sync.WaitGroup) { pon.storeGemPort(1, sn1); wg.Done() }(&wg)
	go func(wg *sync.WaitGroup) { pon.storeGemPort(2, sn2); wg.Done() }(&wg)
	go func(wg *sync.WaitGroup) { pon.storeGemPort(3, sn3); wg.Done() }(&wg)

	wg.Wait()

	assert.Equal(b, len(pon.allocatedGemPorts), 3)
}

func Benchmark_removeGemPort(b *testing.B) {
	pon := PonPort{
		allocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
	}

	pon.storeGemPort(1, sn1)
	pon.storeGemPort(2, sn2)
	pon.storeGemPort(3, sn3)

	assert.Equal(b, len(pon.allocatedGemPorts), 3)

	wg := sync.WaitGroup{}
	wg.Add(3)

	// concurrently add multiple ports
	go func(wg *sync.WaitGroup) { pon.removeGemPort(1); wg.Done() }(&wg)
	go func(wg *sync.WaitGroup) { pon.removeGemPort(2); wg.Done() }(&wg)
	go func(wg *sync.WaitGroup) { pon.removeGemPort(3); wg.Done() }(&wg)

	wg.Wait()

	assert.Equal(b, len(pon.allocatedGemPorts), 0)
}

func Test_removeGemPort(t *testing.T) {
	pon := &PonPort{
		allocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
	}

	pon.storeGemPort(1, sn1)
	pon.storeGemPort(2, sn2)
	assert.Equal(t, len(pon.allocatedGemPorts), 2)

	// remove a non exiting gemPort
	pon.removeGemPort(3)
	assert.Equal(t, len(pon.allocatedGemPorts), 2)

	// remove an existing gemPort
	pon.removeGemPort(1)
	assert.Equal(t, len(pon.allocatedGemPorts), 1)

}

func Test_isGemPortAllocated(t *testing.T) {
	pon := &PonPort{
		allocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
	}

	pon.storeGemPort(1, sn1)

	assert.Equal(t, len(pon.allocatedGemPorts), 1)

	free, sn := pon.isGemPortAllocated(1)

	assert.Equal(t, free, true)
	assert.Equal(t, sn, sn1)

	used, sn_ := pon.isGemPortAllocated(2)

	assert.Equal(t, used, false)
	assert.Nil(t, sn_)
}
