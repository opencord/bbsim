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
	"github.com/google/gopacket/layers"
	"gotest.tools/assert"
	"net"
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

	onu := CreateONU(&olt, &pon, 1, 0, false)

	assert.Equal(t, onu.Sn(), "BBSM00000101")
}

func Test_AddGemPortToService_eapol(t *testing.T) {

	hsia := Service{Name: "hsia", NeedsEapol: true, CTag: 900}
	voip := Service{Name: "voip", NeedsEapol: false, CTag: 55}

	onu := createTestOnu()

	onu.Services = []ServiceIf{&hsia, &voip}

	onu.addGemPortToService(1024, uint32(layers.EthernetTypeEAPOL), 0, 0)

	assert.Equal(t, hsia.GemPort, uint32(1024))
	assert.Equal(t, voip.GemPort, uint32(0))
}

func Test_AddGemPortToService_dhcp(t *testing.T) {

	hsia := Service{Name: "hsia", NeedsEapol: true}
	voip := Service{Name: "voip", NeedsDhcp: true, CTag: 900}
	mc := Service{Name: "mc", CTag: 900}

	onu := createTestOnu()

	onu.Services = []ServiceIf{&hsia, &voip, &mc}

	onu.addGemPortToService(1025, uint32(layers.EthernetTypeIPv4), 900, 0)

	assert.Equal(t, hsia.GemPort, uint32(0))
	assert.Equal(t, voip.GemPort, uint32(1025))
	assert.Equal(t, mc.GemPort, uint32(0))
}

func Test_AddGemPortToService_dataplane(t *testing.T) {

	hsia := Service{Name: "hsia", NeedsEapol: true, CTag: 900, STag: 500}
	voip := Service{Name: "voip", NeedsDhcp: true, CTag: 900}

	onu := createTestOnu()

	onu.Services = []ServiceIf{&hsia, &voip}

	onu.addGemPortToService(1024, uint32(layers.EthernetTypeLLC), 500, 900)

	assert.Equal(t, hsia.GemPort, uint32(1024))
	assert.Equal(t, voip.GemPort, uint32(0))
}

func Test_FindServiceByMacAddress(t *testing.T) {

	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(2)}

	hsia := Service{Name: "hsia", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}}
	voip := Service{Name: "voip", HwAddress: mac}
	vod := Service{Name: "vod", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(3)}}

	onu := createTestOnu()

	onu.Services = []ServiceIf{&hsia, &voip, &vod}

	service, err := onu.findServiceByMacAddress(mac)
	assert.NilError(t, err)
	assert.Equal(t, service.HwAddress.String(), mac.String())
}
