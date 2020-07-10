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
	"github.com/opencord/bbsim/internal/common"
	"gotest.tools/assert"
)

func createMockDevices() (*devices.OltDevice, *devices.Onu) {
	olt := &devices.OltDevice{
		ID: 0,
	}

	onu := &devices.Onu{
		ID:        1,
		PonPortID: 1,
		STag:      900,
		CTag:      923,
		HwAddress: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(1), byte(1)},
		PortNo:    0,
	}
	onu.SerialNumber = onu.NewSN(0, onu.PonPortID, onu.ID)

	return olt, onu
}

func TestSadisServer_GetOnuEntryV1(t *testing.T) {

	olt, onu := createMockDevices()

	uni := "1"

	res, err := GetOnuEntryV1(olt, onu, uni)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))
	assert.Equal(t, res.CTag, 923)
	assert.Equal(t, res.STag, 900)
	assert.Equal(t, res.RemoteID, string(olt.SerialNumber))
	assert.Equal(t, res.DownstreamBandwidthProfile, "Default")
	assert.Equal(t, res.UpstreamBandwidthProfile, "User_Bandwidth1")
	assert.Equal(t, res.TechnologyProfileID, 64)

}

func TestSadisServer_GetOnuEntryV2_Att(t *testing.T) {
	olt, onu := createMockDevices()

	uni := "1"

	res, err := GetOnuEntryV2(olt, onu, uni)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))
	assert.Equal(t, res.RemoteID, fmt.Sprintf("%s-%s", onu.Sn(), uni))

	// assert the correct type
	uniTagList, ok := res.UniTagList[0].(SadisUniTagAtt)
	if !ok {
		t.Fatal("UniTagList has the wrong type")
	}

	assert.Equal(t, uniTagList.PonCTag, 923)
	assert.Equal(t, uniTagList.PonSTag, 900)
	assert.Equal(t, uniTagList.DownstreamBandwidthProfile, "User_Bandwidth1")
	assert.Equal(t, uniTagList.UpstreamBandwidthProfile, "Default")
	assert.Equal(t, uniTagList.TechnologyProfileID, 64)
	assert.Equal(t, uniTagList.IsDhcpRequired, false)
	assert.Equal(t, uniTagList.IsIgmpRequired, false)
}

func TestSadisServer_GetOnuEntryV2_Dt(t *testing.T) {
	common.Options.BBSim.SadisFormat = common.SadisFormatDt
	olt, onu := createMockDevices()

	uni := "1"

	res, err := GetOnuEntryV2(olt, onu, uni)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, res.ID, fmt.Sprintf("%s-%s", onu.Sn(), uni))
	assert.Equal(t, res.RemoteID, fmt.Sprintf("%s-%s", onu.Sn(), uni))

	// assert the correct type
	uniTagList, ok := res.UniTagList[0].(SadisUniTagDt)
	if !ok {
		t.Fatal("UniTagList has the wrong type")
	}

	assert.Equal(t, uniTagList.PonCTag, 4096)
	assert.Equal(t, uniTagList.PonSTag, 900)
	assert.Equal(t, uniTagList.DownstreamBandwidthProfile, "User_Bandwidth1")
	assert.Equal(t, uniTagList.UpstreamBandwidthProfile, "Default")
	assert.Equal(t, uniTagList.TechnologyProfileID, 64)
	assert.Equal(t, uniTagList.UniTagMatch, 4096)
}
