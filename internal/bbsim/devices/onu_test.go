/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Onu_CreateOnu(t *testing.T) {
	nextCtag := map[string]int{}
	nextStag := map[string]int{}

	olt := OltDevice{
		ID:             0,
		NumUni:         4,
		NumPots:        1,
		NniDhcpTrapVid: 60,
	}
	pon := PonPort{
		ID:  1,
		Olt: &olt,
	}

	onu := CreateONU(&olt, &pon, 1, 0, nextCtag, nextStag, false)

	assert.Equal(t, "BBSM00000101", onu.Sn())
	assert.Equal(t, 4, len(onu.UniPorts))
	assert.Equal(t, 1, len(onu.PotsPorts))
	assert.Equal(t, 60, olt.NniDhcpTrapVid)

}
