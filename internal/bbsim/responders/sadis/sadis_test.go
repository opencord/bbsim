/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the License);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an AS IS BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sadis

import (
	"fmt"
	"net"
	"testing"

	"github.com/opencord/bbsim/internal/bbsim/devices"
	"gotest.tools/assert"
)

func createMockDevices() (*devices.OltDevice, *devices.Onu) {
	olt := &devices.OltDevice{
		ID: 0,
	}

	onu := &devices.Onu{
		ID:        1,
		PonPortID: 1,
		PortNo:    0,
	}

	mac := net.HardwareAddr{0x2e, 0x60, 0x01, byte(1), byte(1), byte(0)}

	onu.SerialNumber = onu.NewSN(0, onu.PonPortID, onu.ID)
	onu.Services = []devices.ServiceIf{
		&devices.Service{Name: "hsia", CTag: 923, STag: 900, NeedsEapol: true, NeedsDhcp: true, NeedsIgmp: true, HwAddress: mac, TechnologyProfileID: 64},
	}

	return olt, onu
}

func TestSadisServer_GetOnuEntryV2(t *testing.T) {

	olt, onu := createMockDevices()

	uni := "1"

	entry, err := GetOnuEntryV2(olt, onu, uni)

	assert.NilError(t, err)

	assert.Equal(t, entry.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))
	assert.Equal(t, entry.RemoteID, fmt.Sprintf("%s-%s", onu.Sn(), uni))

	assert.Equal(t, entry.UniTagList[0].PonCTag, 923)
	assert.Equal(t, entry.UniTagList[0].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[0].DownstreamBandwidthProfile, "User_Bandwidth2")
	assert.Equal(t, entry.UniTagList[0].UpstreamBandwidthProfile, "User_Bandwidth1")
	assert.Equal(t, entry.UniTagList[0].TechnologyProfileID, 64)
	assert.Equal(t, entry.UniTagList[0].IsDhcpRequired, true)
	assert.Equal(t, entry.UniTagList[0].IsIgmpRequired, true)
}

func TestSadisServer_GetOnuEntryV2_multi_service(t *testing.T) {

	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(2)}

	hsia := devices.Service{Name: "hsia", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)},
		CTag: 900, STag: 900, TechnologyProfileID: 64}

	voip := devices.Service{Name: "voip", HwAddress: mac,
		CTag: 901, STag: 900, TechnologyProfileID: 65}

	vod := devices.Service{Name: "vod", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(3)},
		CTag: 902, STag: 900, TechnologyProfileID: 66}

	olt, onu := createMockDevices()

	onu.Services = []devices.ServiceIf{&hsia, &voip, &vod}

	uni := "1"

	entry, err := GetOnuEntryV2(olt, onu, uni)

	assert.NilError(t, err)

	assert.Equal(t, entry.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))
	assert.Equal(t, entry.RemoteID, fmt.Sprintf("%s-%s", onu.Sn(), uni))

	assert.Equal(t, len(entry.UniTagList), 3)

	assert.Equal(t, entry.UniTagList[0].PonCTag, 900)
	assert.Equal(t, entry.UniTagList[0].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[0].TechnologyProfileID, 64)

	assert.Equal(t, entry.UniTagList[1].PonCTag, 901)
	assert.Equal(t, entry.UniTagList[1].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[1].TechnologyProfileID, 65)

	assert.Equal(t, entry.UniTagList[2].PonCTag, 902)
	assert.Equal(t, entry.UniTagList[2].PonSTag, 900)
	assert.Equal(t, entry.UniTagList[2].TechnologyProfileID, 66)
}
