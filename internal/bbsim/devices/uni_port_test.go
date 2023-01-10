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
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func createTestUni() *UniPort {
	onu := &Onu{
		ID:           0,
		SerialNumber: NewSN(1, 1, 1),
		PonPortID:    0,
	}

	services := []*Service{
		{Name: "hsia", HwAddress: net.HardwareAddr{0x2e, 0x00}},
	}

	uni := UniPort{
		ID:     1,
		MeId:   omcilib.GenerateUniPortEntityId(1),
		PortNo: 16,
		Onu:    onu,
		logger: uniLogger,
	}

	for _, s := range services {
		uni.Services = append(uni.Services, s)
	}
	return &uni
}

func TestNewUniPortAtt(t *testing.T) {

	const (
		hsia = "hsia"
		voip = "voip"
		vod  = "vod"
		mc   = "mc"
	)

	type args struct {
		services []common.ServiceYaml
		nextCtag map[string]int
		nextStag map[string]int
	}

	type wants struct {
		expectedCtags map[string]int
		expectedStags map[string]int
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{"newUniPort-att",
			args{
				services: []common.ServiceYaml{
					{Name: hsia, CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), NeedsEapol: true, NeedsDhcp: true, NeedsIgmp: true},
				},
				nextCtag: map[string]int{hsia: 920},
				nextStag: map[string]int{hsia: 900},
			},
			wants{
				expectedCtags: map[string]int{hsia: 921},
				expectedStags: map[string]int{hsia: 900},
			},
		},
		{"newUniPort-dt",
			args{
				services: []common.ServiceYaml{
					{Name: hsia, CTag: 900, CTagAllocation: common.TagAllocationShared.String(), STag: 900, STagAllocation: common.TagAllocationUnique.String(), UniTagMatch: 4096},
				},
				nextCtag: map[string]int{hsia: 920},
				nextStag: map[string]int{hsia: 900},
			},
			wants{
				expectedCtags: map[string]int{hsia: 920},
				expectedStags: map[string]int{hsia: 901},
			},
		},
		{"newUniPort-tt",
			args{
				services: []common.ServiceYaml{
					{Name: hsia, CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 35, TechnologyProfileID: 64},
					{Name: voip, CTag: 444, CTagAllocation: common.TagAllocationShared.String(), STag: 333, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 65, TechnologyProfileID: 65, ConfigureMacAddress: true, UsPonCTagPriority: 7, UsPonSTagPriority: 7, DsPonCTagPriority: 7, DsPonSTagPriority: 7},
					{Name: vod, CTag: 55, CTagAllocation: common.TagAllocationShared.String(), STag: 555, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 55, TechnologyProfileID: 66, NeedsDhcp: true, NeedsIgmp: true, ConfigureMacAddress: true, UsPonCTagPriority: 5, UsPonSTagPriority: 5, DsPonCTagPriority: 5, DsPonSTagPriority: 5},
				},
				nextCtag: map[string]int{hsia: 920},
				nextStag: map[string]int{hsia: 900},
			},
			wants{
				expectedCtags: map[string]int{hsia: 921, voip: 444, vod: 55},
				expectedStags: map[string]int{hsia: 900, voip: 333, vod: 555},
			},
		},
		{"newUniPort-tt-maclearning-pppoe",
			args{
				services: []common.ServiceYaml{
					{Name: hsia, CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 35, TechnologyProfileID: 64},
					{Name: voip, CTag: 444, CTagAllocation: common.TagAllocationShared.String(), STag: 333, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 65, TechnologyProfileID: 65, EnableMacLearning: true, UsPonCTagPriority: 7, UsPonSTagPriority: 7, DsPonCTagPriority: 7, DsPonSTagPriority: 7},
					{Name: vod, CTag: 55, CTagAllocation: common.TagAllocationShared.String(), STag: 555, STagAllocation: common.TagAllocationShared.String(), UniTagMatch: 55, TechnologyProfileID: 66, NeedsDhcp: true, NeedsIgmp: true, NeedsPPPoE: true, EnableMacLearning: true, UsPonCTagPriority: 5, UsPonSTagPriority: 5, DsPonCTagPriority: 5, DsPonSTagPriority: 5},
				},
				nextCtag: map[string]int{hsia: 920},
				nextStag: map[string]int{hsia: 900},
			},
			wants{
				expectedCtags: map[string]int{hsia: 921, voip: 444, vod: 55},
				expectedStags: map[string]int{hsia: 900, voip: 333, vod: 555},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.Services = tt.args.services

			onu := &Onu{
				ID:           0,
				SerialNumber: NewSN(1, 1, 1),
			}

			uni, err := NewUniPort(1, onu, tt.args.nextCtag, tt.args.nextStag)

			assert.NoError(t, err)

			assert.Equal(t, uint32(1), uni.ID)
			assert.Equal(t, uint16(258), uni.MeId.ToUint16())
			assert.Equal(t, len(tt.args.services), len(uni.Services))

			assert.Equal(t, len(tt.args.services), len(uni.Services))

			for i, configuredService := range tt.args.services {
				service := uni.Services[i].(*Service)
				assert.Equal(t, tt.wants.expectedCtags[service.Name], service.CTag)
				assert.Equal(t, tt.wants.expectedStags[service.Name], service.STag)
				assert.Equal(t, configuredService.NeedsEapol, service.NeedsEapol)
				assert.Equal(t, configuredService.NeedsDhcp, service.NeedsDhcp)
				assert.Equal(t, configuredService.NeedsIgmp, service.NeedsIgmp)
				assert.Equal(t, configuredService.NeedsPPPoE, service.NeedsPPPoE)
				assert.Equal(t, configuredService.UniTagMatch, service.UniTagMatch)
				assert.Equal(t, configuredService.TechnologyProfileID, service.TechnologyProfileID)
				assert.Equal(t, configuredService.ConfigureMacAddress, service.ConfigureMacAddress)
				assert.Equal(t, configuredService.EnableMacLearning, service.EnableMacLearning)
				assert.Equal(t, configuredService.UsPonCTagPriority, service.UsPonCTagPriority)
				assert.Equal(t, configuredService.DsPonCTagPriority, service.DsPonCTagPriority)
				assert.Equal(t, configuredService.UsPonSTagPriority, service.UsPonSTagPriority)
				assert.Equal(t, configuredService.DsPonSTagPriority, service.DsPonSTagPriority)
			}
		})
	}
}

func Test_AddGemPortToService_eapol(t *testing.T) {

	uni := createTestUni()
	hsia := Service{Name: "hsia", NeedsEapol: true, CTag: 900, UniPort: uni}
	voip := Service{Name: "voip", NeedsEapol: false, CTag: 55, UniPort: uni}
	uni.Services = []ServiceIf{&hsia, &voip}
	uni.addGemPortToService(1024, uint32(layers.EthernetTypeEAPOL), 0, 0)

	assert.Equal(t, hsia.GemPort, uint32(1024))
	assert.Equal(t, voip.GemPort, uint32(0))
}

func Test_AddGemPortToService_dhcp(t *testing.T) {

	uni := createTestUni()
	hsia := Service{Name: "hsia", NeedsEapol: true, UniPort: uni}
	voip := Service{Name: "voip", NeedsDhcp: true, CTag: 900, UniPort: uni}
	mc := Service{Name: "mc", CTag: 900, UniPort: uni}
	uni.Services = []ServiceIf{&hsia, &voip, &mc}
	uni.addGemPortToService(1025, uint32(layers.EthernetTypeIPv4), 900, 0)

	assert.Equal(t, hsia.GemPort, uint32(0))
	assert.Equal(t, voip.GemPort, uint32(1025))
	assert.Equal(t, mc.GemPort, uint32(0))
}

func Test_AddGemPortToService_dataplane(t *testing.T) {

	uni := createTestUni()
	hsia := Service{Name: "hsia", NeedsEapol: true, CTag: 900, STag: 500, UniPort: uni}
	voip := Service{Name: "voip", NeedsDhcp: true, CTag: 900, UniPort: uni}
	uni.Services = []ServiceIf{&hsia, &voip}
	uni.addGemPortToService(1024, uint32(layers.EthernetTypeLLC), 500, 900)

	assert.Equal(t, hsia.GemPort, uint32(1024))
	assert.Equal(t, voip.GemPort, uint32(0))
}

func Test_FindServiceByMacAddress(t *testing.T) {

	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(2)}

	uni := createTestUni()
	hsia := Service{Name: "hsia", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}, UniPort: uni}
	voip := Service{Name: "voip", HwAddress: mac, UniPort: uni}
	vod := Service{Name: "vod", HwAddress: net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(3)}, UniPort: uni}
	uni.Services = []ServiceIf{&hsia, &voip, &vod}

	service, err := uni.findServiceByMacAddress(mac)
	assert.NoError(t, err)
	assert.Equal(t, service.HwAddress.String(), mac.String())
}
