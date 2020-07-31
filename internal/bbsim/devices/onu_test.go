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
	"gotest.tools/assert"
	"testing"
)

func Test_Onu_CreateOnu(t *testing.T) {

	olt := OltDevice{
		ID: 0,
	}
	pon := PonPort{
		ID:  1,
		Olt: &olt,
	}

	onu := CreateONU(&olt, &pon, 1, 900, 900, true, false, 0, false)

	assert.Equal(t, onu.Sn(), "BBSM00000101")
	assert.Equal(t, onu.STag, 900)
	assert.Equal(t, onu.CTag, 900)
	assert.Equal(t, onu.Auth, true)
	assert.Equal(t, onu.Dhcp, false)
	assert.Equal(t, onu.HwAddress.String(), "2e:60:70:00:01:01")
}

func Test_Onu_FindGemPort(t *testing.T) {
	onu := createTestOnu()

	gem1, _ := NewGemPort(onu,1024, 24)
	gem2, _ := NewGemPort(onu,1025, 25)
	gem3, _ := NewGemPort(onu,1026, 26)
	onu.GemPorts[gem1.ID] = gem1
	onu.GemPorts[gem2.ID] = gem2
	onu.GemPorts[gem3.ID] = gem3

	found, gem := onu.findGemPort(1026)

	assert.Equal(t, found, true)
	assert.Equal(t, gem.ID, gem3.ID)
	assert.Equal(t, gem.ID, gem3.ID)

	found, gem = onu.findGemPort(2048)
	assert.Equal(t, found,false)
}

func Test_Onu_AddGemPort(t *testing.T) {
	onu := createTestOnu()

	assert.Equal(t, len(onu.GemPorts), 0)

	onu.addGemPort(1024, 24)
	assert.Equal(t, len(onu.GemPorts), 1)

	// do not duplicate GemPorts
	onu.addGemPort(1024, 24)
	assert.Equal(t, len(onu.GemPorts), 1)

	// TODO test VLAN update
}

func Test_Onu_RemoveGemPort(t *testing.T) {
	onu := createTestOnu()

	gem1, _ := NewGemPort(onu,1024, 24)
	gem2, _ := NewGemPort(onu,1025, 25)
	gem3, _ := NewGemPort(onu,1026, 26)
	onu.GemPorts[gem1.ID] = gem1
	onu.GemPorts[gem2.ID] = gem2
	onu.GemPorts[gem3.ID] = gem3

	assert.Equal(t, len(onu.GemPorts), 3)

	err := onu.removeGemPort(1024)

	assert.NilError(t, err)
	assert.Equal(t, len(onu.GemPorts), 2)

	// delete  a non-existing gemPort
	err = onu.removeGemPort(2048)
	assert.NilError(t, err)
	assert.Equal(t, len(onu.GemPorts), 2)
}
