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

	onu := CreateONU(&olt, pon, 1, 900, 900, true, false,0,false)

	assert.Equal(t, onu.Sn(), "BBSM00000101")
	assert.Equal(t, onu.STag, 900)
	assert.Equal(t, onu.CTag, 900)
	assert.Equal(t, onu.Auth, true)
	assert.Equal(t, onu.Dhcp, false)
	assert.Equal(t, onu.HwAddress.String(), "2e:60:70:00:01:01")
}