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

func createMockOlt(numPon int, numOnu int) OltDevice {
	olt := OltDevice{
		ID: 0,
	}

	for i := 0; i < numPon; i++ {
		pon := PonPort{
			ID: uint32(i),
		}

		for j := 0; j < numOnu; j++ {
			onu := Onu{
				ID:        uint32(i + j),
				PonPort:   pon,
				PonPortID: uint32(i),
			}
			onu.SerialNumber = onu.NewSN(olt.ID, pon.ID, onu.ID)
			pon.Onus = append(pon.Onus, onu)
		}
		olt.Pons = append(olt.Pons, pon)
	}
	return olt
}

func Test_Olt_FindOnu_Success(t *testing.T) {

	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	onu, err := olt.FindOnu("BBSM00000303")

	assert.Equal(t, err, nil)
	assert.Equal(t, onu.Sn(), "BBSM00000303")
	assert.Equal(t, onu.ID, uint32(3))
	assert.Equal(t, onu.PonPortID, uint32(3))
}

func Test_Olt_FindOnu_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu)

	_, err := olt.FindOnu("BBSM00000303")

	assert.Equal(t, err.Error(), "cannot-find-onu-BBSM00000303")
}
