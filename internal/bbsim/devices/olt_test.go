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
	"github.com/opencord/bbsim/internal/common"
	"net"
	"testing"

	"github.com/opencord/voltha-protos/v2/go/openolt"
	"gotest.tools/assert"
)

func createMockOlt(numPon int, numOnu int, services []ServiceIf) *OltDevice {
	olt := &OltDevice{
		ID: 0,
	}

	for i := 0; i < numPon; i++ {
		pon := PonPort{
			ID: uint32(i),
		}

		for j := 0; j < numOnu; j++ {
			onuId := uint32(i + j)
			onu := Onu{
				ID:        onuId,
				PonPort:   &pon,
				PonPortID: pon.ID,
			}

			for k, s := range services {
				service := s.(*Service)
				service.HwAddress = net.HardwareAddr{0x2e, 0x60, byte(olt.ID), byte(pon.ID), byte(onuId), byte(k)}
				service.Onu = &onu
				onu.Services = append(onu.Services, service)
			}

			onu.SerialNumber = onu.NewSN(olt.ID, pon.ID, onu.ID)
			pon.Onus = append(pon.Onus, &onu)
		}
		olt.Pons = append(olt.Pons, &pon)
	}
	return olt
}

// check the creation of an OLT with a single Service
func TestCreateOLT(t *testing.T) {

	common.Services = []common.ServiceYaml{
		{Name: "hsia", CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), NeedsEapol: true, NeedsDchp: true, NeedsIgmp: true},
	}

	common.Config = &common.GlobalConfig{
		Olt: common.OltConfig{
			ID:          1,
			PonPorts:    2,
			OnusPonPort: 2,
		},
	}

	olt := CreateOLT(*common.Config, common.Services, true)

	assert.Equal(t, len(olt.Pons), int(common.Config.Olt.PonPorts))

	// count the ONUs
	onus := 0
	for _, p := range olt.Pons {
		onus = onus + len(p.Onus)
	}

	assert.Equal(t, onus, int(common.Config.Olt.PonPorts*common.Config.Olt.OnusPonPort))

	// count the services
	services := 0
	for _, p := range olt.Pons {
		for _, o := range p.Onus {
			services = services + len(o.Services)
		}
	}

	assert.Equal(t, services, int(common.Config.Olt.PonPorts)*int(common.Config.Olt.OnusPonPort)*len(common.Services))

	s1 := olt.Pons[0].Onus[0].Services[0].(*Service)

	assert.Equal(t, s1.Name, "hsia")
	assert.Equal(t, s1.CTag, 900)
	assert.Equal(t, s1.STag, 900)
	assert.Equal(t, s1.HwAddress.String(), "2e:60:01:00:01:00")
	assert.Equal(t, olt.Pons[0].Onus[0].ID, uint32(1))

	s2 := olt.Pons[0].Onus[1].Services[0].(*Service)
	assert.Equal(t, s2.CTag, 901)
	assert.Equal(t, s2.STag, 900)
	assert.Equal(t, s2.HwAddress.String(), "2e:60:01:00:02:00")
	assert.Equal(t, olt.Pons[0].Onus[1].ID, uint32(2))

	s3 := olt.Pons[1].Onus[0].Services[0].(*Service)
	assert.Equal(t, s3.CTag, 902)
	assert.Equal(t, s3.STag, 900)
	assert.Equal(t, s3.HwAddress.String(), "2e:60:01:01:01:00")
	assert.Equal(t, olt.Pons[1].Onus[0].ID, uint32(1))

	s4 := olt.Pons[1].Onus[1].Services[0].(*Service)
	assert.Equal(t, s4.CTag, 903)
	assert.Equal(t, s4.STag, 900)
	assert.Equal(t, s4.HwAddress.String(), "2e:60:01:01:02:00")
	assert.Equal(t, olt.Pons[1].Onus[1].ID, uint32(2))
}

func Test_Olt_FindOnuBySn_Success(t *testing.T) {

	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, []ServiceIf{})

	onu, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err, nil)
	assert.Equal(t, onu.Sn(), "BBSM00000303")
	assert.Equal(t, onu.ID, uint32(3))
	assert.Equal(t, onu.PonPortID, uint32(3))
}

func Test_Olt_FindOnuBySn_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, []ServiceIf{})

	_, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err.Error(), "cannot-find-onu-by-serial-number-BBSM00000303")
}

func Test_Olt_FindOnuByMacAddress_Success(t *testing.T) {
	numPon := 4
	numOnu := 4

	services := []ServiceIf{
		&Service{Name: "hsia"},
		&Service{Name: "voip"},
		&Service{Name: "vod"},
	}

	olt := createMockOlt(numPon, numOnu, services)

	mac := net.HardwareAddr{0x2e, 0x60, byte(olt.ID), byte(3), byte(6), byte(1)}
	s, err := olt.FindServiceByMacAddress(mac)

	assert.NilError(t, err)

	service := s.(*Service)

	assert.Equal(t, err, nil)
	assert.Equal(t, service.Onu.Sn(), "BBSM00000306")
	assert.Equal(t, service.Onu.ID, uint32(6))
	assert.Equal(t, service.Onu.PonPortID, uint32(3))

	assert.Equal(t, service.Name, "voip")
}

func Test_Olt_FindOnuByMacAddress_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, []ServiceIf{})

	mac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(3), byte(3)}

	_, err := olt.FindServiceByMacAddress(mac)

	assert.Equal(t, err.Error(), "cannot-find-service-by-mac-address-2e:60:70:13:03:03")
}

func Test_Olt_GetOnuByFlowId(t *testing.T) {
	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, []ServiceIf{})

	// Add the flows to onus (to be found)
	onu1, _ := olt.FindOnuBySn("BBSM00000303")
	flow1 := openolt.Flow{
		FlowId:     64,
		Classifier: &openolt.Classifier{},
	}
	msg1 := OnuFlowUpdateMessage{
		OnuID:     onu1.ID,
		PonPortID: onu1.PonPortID,
		Flow:      &flow1,
	}
	onu1.handleFlowAdd(msg1)

	onu2, _ := olt.FindOnuBySn("BBSM00000103")
	flow2 := openolt.Flow{
		FlowId:     72,
		Classifier: &openolt.Classifier{},
	}
	msg2 := OnuFlowUpdateMessage{
		OnuID:     onu2.ID,
		PonPortID: onu2.PonPortID,
		Flow:      &flow2,
	}
	onu2.handleFlowAdd(msg2)

	found, err := olt.GetOnuByFlowId(flow1.FlowId)

	assert.Equal(t, err, nil)
	assert.Equal(t, found.Sn(), onu1.Sn())
}
