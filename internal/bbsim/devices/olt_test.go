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
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/types"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/stretchr/testify/assert"
)

func createMockOlt(numPon int, numOnu int, numUni int, services []ServiceIf) *OltDevice {
	olt := &OltDevice{
		ID:               0,
		AllocIDs:         make(map[uint32]map[uint32]map[uint32]map[int32]map[uint64]bool),
		GemPortIDs:       make(map[uint32]map[uint32]map[uint32]map[int32]map[uint64]bool),
		OmciResponseRate: 10,
	}

	for i := 0; i < numPon; i++ {

		// initialize the resource maps for every PON Ports
		olt.AllocIDs[uint32(i)] = make(map[uint32]map[uint32]map[int32]map[uint64]bool)
		olt.GemPortIDs[uint32(i)] = make(map[uint32]map[uint32]map[int32]map[uint64]bool)

		pon := PonPort{
			ID: uint32(i),
		}

		for j := 0; j < numOnu; j++ {

			// initialize the resource maps for every ONU and the first UNI
			olt.AllocIDs[uint32(i)][uint32(j)] = make(map[uint32]map[int32]map[uint64]bool)
			olt.GemPortIDs[uint32(i)][uint32(j)] = make(map[uint32]map[int32]map[uint64]bool)

			onuId := uint32(i + j)
			onu := Onu{
				ID:        onuId,
				PonPort:   &pon,
				PonPortID: pon.ID,
				InternalState: fsm.NewFSM(
					OnuStateCreated,
					// this is fake state machine, we don't care about transition in the OLT
					// unit tests, we'll use SetState to emulate cases
					fsm.Events{
						{Name: OnuTxEnable, Src: []string{}, Dst: OnuStateEnabled},
						{Name: OnuTxDisable, Src: []string{}, Dst: OnuStateDisabled},
					},
					fsm.Callbacks{},
				),
				Channel: make(chan bbsim.Message, 2048),
			}

			for k := 0; k < numUni; k++ {
				uni := UniPort{
					ID:     uint32(k + 1),
					Onu:    &onu,
					logger: uniLogger,
				}
				for l, s := range services {
					service := s.(*Service)
					service.HwAddress = net.HardwareAddr{0x2e, byte(olt.ID), byte(pon.ID), byte(onuId), byte(k), byte(l)}
					service.UniPort = &uni
					uni.Services = append(uni.Services, service)
				}
				onu.UniPorts = append(onu.UniPorts, &uni)
			}

			onu.SerialNumber = NewSN(olt.ID, pon.ID, onu.ID)
			pon.Onus = append(pon.Onus, &onu)
		}
		olt.Pons = append(olt.Pons, &pon)
	}
	return olt
}

// check the creation of an OLT with a single Service
func TestCreateOLT(t *testing.T) {

	common.Services = []common.ServiceYaml{
		{Name: "hsia", CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), NeedsEapol: true, NeedsDhcp: true, NeedsIgmp: true},
	}

	common.Config = &common.GlobalConfig{
		Olt: common.OltConfig{
			ID:          1,
			PonPorts:    2,
			OnusPonPort: 2,
			UniPorts:    4,
		},
	}

	allocIdPerOnu := uint32(common.Config.Olt.UniPorts * uint32(len(common.Services)))
	common.PonsConfig = &common.PonPortsConfig{
		Number: common.Config.Olt.PonPorts,
		Ranges: []common.PonRangeConfig{
			{
				PonRange:     common.IdRange{StartId: 0, EndId: common.Config.Olt.PonPorts - 1},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 1 + (common.Config.Olt.OnusPonPort - 1)},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1024 + (common.Config.Olt.OnusPonPort * allocIdPerOnu)},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + common.Config.Olt.OnusPonPort*allocIdPerOnu*8},
			},
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

	// counte the UNIs
	unis := 0
	for _, p := range olt.Pons {
		for _, o := range p.Onus {
			unis = unis + len(o.UniPorts)
		}
	}
	// NOTE when unis will be configurable this test will need to adapt
	assert.Equal(t, unis, int(common.Config.Olt.PonPorts*common.Config.Olt.OnusPonPort*common.Config.Olt.UniPorts))

	// count the services
	services := 0
	for _, p := range olt.Pons {
		for _, o := range p.Onus {
			for _, u := range o.UniPorts {
				uni := u.(*UniPort)
				services = services + len(uni.Services)
			}
		}
	}
	// NOTE when unis will be configurable this test will need to adapt
	assert.Equal(t, services, int(common.Config.Olt.PonPorts)*int(common.Config.Olt.OnusPonPort)*int(common.Config.Olt.UniPorts)*len(common.Services))

	s1 := olt.Pons[0].Onus[0].UniPorts[0].(*UniPort).Services[0].(*Service)

	assert.Equal(t, s1.Name, "hsia")
	assert.Equal(t, s1.CTag, 900)
	assert.Equal(t, s1.STag, 900)
	assert.Equal(t, "2e:01:00:01:00:00", s1.HwAddress.String())
	assert.Equal(t, olt.Pons[0].Onus[0].ID, uint32(1))

	// each ONU has 4 UNIs, taking up the c-tags
	s2 := olt.Pons[0].Onus[1].UniPorts[0].(*UniPort).Services[0].(*Service)
	assert.Equal(t, s2.CTag, 904)
	assert.Equal(t, s2.STag, 900)
	assert.Equal(t, s2.HwAddress.String(), "2e:01:00:02:00:00")
	assert.Equal(t, olt.Pons[0].Onus[1].ID, uint32(2))

	s3 := olt.Pons[1].Onus[0].UniPorts[0].(*UniPort).Services[0].(*Service)
	assert.Equal(t, s3.CTag, 908)
	assert.Equal(t, s3.STag, 900)
	assert.Equal(t, s3.HwAddress.String(), "2e:01:01:01:00:00")
	assert.Equal(t, olt.Pons[1].Onus[0].ID, uint32(1))

	s4 := olt.Pons[1].Onus[1].UniPorts[0].(*UniPort).Services[0].(*Service)
	assert.Equal(t, s4.CTag, 912)
	assert.Equal(t, s4.STag, 900)
	assert.Equal(t, s4.HwAddress.String(), "2e:01:01:02:00:00")
	assert.Equal(t, olt.Pons[1].Onus[1].ID, uint32(2))
}

func TestGetDeviceInfo(t *testing.T) {

	var onusPerPon uint32 = 4

	common.Services = []common.ServiceYaml{
		{Name: "hsia", CTag: 900, CTagAllocation: common.TagAllocationUnique.String(), STag: 900, STagAllocation: common.TagAllocationShared.String(), NeedsEapol: true, NeedsDhcp: true, NeedsIgmp: true},
	}

	common.Config = &common.GlobalConfig{
		Olt: common.OltConfig{
			ID:          1,
			PonPorts:    2,
			OnusPonPort: onusPerPon,
			UniPorts:    4,
		},
	}

	common.PonsConfig = &common.PonPortsConfig{
		Number: 16,
		Ranges: []common.PonRangeConfig{
			{
				PonRange:     common.IdRange{StartId: 0, EndId: 1},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1024 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 2, EndId: 2},
				Technology:   common.GPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 3, EndId: 3},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 4, EndId: 6},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 7, EndId: 8},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 9, EndId: 9},
				Technology:   common.GPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 10, EndId: 10},
				Technology:   common.GPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
			{
				PonRange:     common.IdRange{StartId: 11, EndId: 15},
				Technology:   common.XGSPON.String(),
				OnuRange:     common.IdRange{StartId: 1, EndId: 4},
				AllocIdRange: common.IdRange{StartId: 1024, EndId: 1028 + 4},
				GemportRange: common.IdRange{StartId: 1024, EndId: 1024 + 32},
			},
		},
	}

	olt := CreateOLT(*common.Config, common.Services, true)

	res, err := olt.GetDeviceInfo(context.Background(), &openolt.Empty{})

	assert.NoError(t, err, "GetDeviceInfo returned error")

	assert.Equal(t, len(res.Ranges), len(common.PonsConfig.Ranges))

	for _, resRange := range res.Ranges {
		for _, ponId := range resRange.IntfIds {
			conf, err := common.GetPonConfigById(ponId)
			assert.NoError(t, err, fmt.Sprintf("Cannot get pon configuration by id %d", ponId))

			for _, pool := range resRange.Pools {
				switch pool.Type {
				case openolt.DeviceInfo_DeviceResourceRanges_Pool_ONU_ID:
					assert.Equal(t, pool.Start, conf.OnuRange.StartId)
					assert.Equal(t, pool.End, conf.OnuRange.EndId)
				case openolt.DeviceInfo_DeviceResourceRanges_Pool_ALLOC_ID:
					assert.Equal(t, pool.Start, conf.AllocIdRange.StartId)
					assert.Equal(t, pool.End, conf.AllocIdRange.EndId)
				case openolt.DeviceInfo_DeviceResourceRanges_Pool_GEMPORT_ID:
					assert.Equal(t, pool.Start, conf.GemportRange.StartId)
					assert.Equal(t, pool.End, conf.GemportRange.EndId)
				}
			}
		}
	}
}

func Test_Olt_FindOnuBySn_Success(t *testing.T) {

	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	onu, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err, nil)
	assert.Equal(t, onu.Sn(), "BBSM00000303")
	assert.Equal(t, onu.ID, uint32(3))
	assert.Equal(t, onu.PonPortID, uint32(3))
}

func Test_Olt_FindOnuBySn_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	_, err := olt.FindOnuBySn("BBSM00000303")

	assert.Equal(t, err.Error(), "cannot-find-onu-by-serial-number-BBSM00000303")
}

func Test_Olt_FindOnuByMacAddress_Success(t *testing.T) {
	numPon := 4
	numOnu := 4
	numUni := 4

	services := []ServiceIf{
		&Service{Name: "hsia"},
		&Service{Name: "voip"},
		&Service{Name: "vod"},
	}

	olt := createMockOlt(numPon, numOnu, numUni, services)

	mac := net.HardwareAddr{0x2e, byte(olt.ID), byte(3), byte(6), byte(3), byte(1)}
	s, err := olt.FindServiceByMacAddress(mac)

	assert.NoError(t, err)

	service := s.(*Service)

	assert.Equal(t, err, nil)
	assert.Equal(t, service.UniPort.Onu.Sn(), "BBSM00000306")
	assert.Equal(t, service.UniPort.ID, uint32(4))
	assert.Equal(t, service.UniPort.Onu.ID, uint32(6))
	assert.Equal(t, service.UniPort.Onu.PonPortID, uint32(3))

	assert.Equal(t, service.Name, "voip")
}

func Test_Olt_FindOnuByMacAddress_Error(t *testing.T) {

	numPon := 1
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	mac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(3), byte(3)}

	_, err := olt.FindServiceByMacAddress(mac)

	assert.Equal(t, err.Error(), "cannot-find-service-by-mac-address-2e:60:70:13:03:03")
}

func Test_Olt_GetOnuByFlowId(t *testing.T) {
	numPon := 4
	numOnu := 4

	services := []ServiceIf{
		&Service{Name: "hsia"},
	}

	olt := createMockOlt(numPon, numOnu, 1, services)

	// Add the flows to onus (to be found)
	onu1, _ := olt.FindOnuBySn("BBSM00000303")
	flow1 := openolt.Flow{
		FlowId:     64,
		Classifier: &openolt.Classifier{},
		UniId:      1,
	}
	msg1 := types.OnuFlowUpdateMessage{
		OnuID:     onu1.ID,
		PonPortID: onu1.PonPortID,
		Flow:      &flow1,
	}
	onu1.handleFlowAdd(msg1)

	onu2, _ := olt.FindOnuBySn("BBSM00000103")
	flow2 := openolt.Flow{
		FlowId:     72,
		Classifier: &openolt.Classifier{},
		UniId:      1,
	}
	msg2 := types.OnuFlowUpdateMessage{
		OnuID:     onu2.ID,
		PonPortID: onu2.PonPortID,
		Flow:      &flow2,
	}
	onu2.handleFlowAdd(msg2)

	found, err := olt.GetOnuByFlowId(flow1.FlowId)

	assert.Equal(t, err, nil)
	assert.Equal(t, found.Sn(), onu1.Sn())
}

func Test_Olt_storeGemPortId(t *testing.T) {

	const (
		pon  = 1
		onu  = 1
		uni  = 16
		gem1 = 1024
		gem2 = 1025
	)

	numPon := 2
	numOnu := 2

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	// add a first flow on the ONU
	flow1 := &openolt.Flow{
		AccessIntfId: pon,
		OnuId:        onu,
		PortNo:       uni,
		FlowId:       1,
		GemportId:    gem1,
	}

	olt.storeGemPortIdByFlow(flow1)
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 1)       // we have 1 gem port
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem1]), 1) // and one flow referencing it

	// add a second flow on the ONU (same gem)
	flow2 := &openolt.Flow{
		AccessIntfId: pon,
		OnuId:        onu,
		PortNo:       uni,
		FlowId:       2,
		GemportId:    gem1,
	}

	olt.storeGemPortIdByFlow(flow2)
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 1)       // we have 1 gem port
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem1]), 2) // and two flows referencing it

	// add a third flow on the ONU (different gem)
	flow3 := &openolt.Flow{
		AccessIntfId: pon,
		OnuId:        onu,
		PortNo:       uni,
		FlowId:       2,
		GemportId:    1025,
	}

	olt.storeGemPortIdByFlow(flow3)
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 2)       // we have 2 gem ports
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem1]), 2) // two flows referencing the first one
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem2]), 1) // and one flow referencing the second one
}

func Test_Olt_storeGemPortIdReplicatedFlow(t *testing.T) {
	const (
		pon  = 1
		onu  = 1
		uni  = 16
		gem1 = 1024
		gem2 = 1025
	)

	numPon := 2
	numOnu := 2

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	// add a flow that needs replication
	pbitToGemPortMap := make(map[uint32]uint32)
	pbitToGemPortMap[0] = gem1
	pbitToGemPortMap[1] = gem2
	flow1 := &openolt.Flow{
		AccessIntfId:  pon,
		OnuId:         onu,
		PortNo:        uni,
		FlowId:        1,
		GemportId:     0,
		ReplicateFlow: true,
		PbitToGemport: pbitToGemPortMap,
	}

	olt.storeGemPortIdByFlow(flow1)
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 2)       // we have 2 gem ports in the flow
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem1]), 1) // and one flow referencing them
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem2]), 1) // and one flow referencing them
}

func Test_Olt_freeGemPortId(t *testing.T) {
	const (
		pon   = 1
		onu   = 1
		uni   = 16
		gem1  = 1024
		gem2  = 1025
		flow1 = 1
		flow2 = 2
		flow3 = 3
	)

	numPon := 2
	numOnu := 2

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	olt.GemPortIDs[pon][onu][uni] = make(map[int32]map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem1] = make(map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem1][flow1] = true
	olt.GemPortIDs[pon][onu][uni][gem1][flow2] = true
	olt.GemPortIDs[pon][onu][uni][gem2] = make(map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem2][flow3] = true

	// remove one flow on the first gem, check that the gem is still allocated as there is still a flow referencing it
	// NOTE that the flow remove only carries the flow ID, no other information
	flowGem1 := &openolt.Flow{
		FlowId: flow1,
	}

	olt.freeGemPortId(flowGem1)
	// we still have two unis in the map
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 2)

	// we should now have a single gem referenced on this UNI
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni][gem1]), 1, "gemport-not-removed")

	// the gem should still reference flow 2
	assert.Equal(t, olt.GemPortIDs[pon][onu][uni][gem1][flow2], true)
	// but should not reference flow1
	_, flow1Exists := olt.GemPortIDs[pon][onu][uni][gem1][flow1]
	assert.Equal(t, flow1Exists, false)

	// this is the only flow remaining on this gem, the gem should be removed
	flowGem2 := &openolt.Flow{
		FlowId: flow2,
	}
	olt.freeGemPortId(flowGem2)

	// we should now have a single gem referenced on this UNI
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 1, "gemport-not-removed")

	// and it should be gem2
	_, gem1exists := olt.GemPortIDs[pon][onu][uni][gem1]
	assert.Equal(t, gem1exists, false)
	_, gem2exists := olt.GemPortIDs[pon][onu][uni][gem2]
	assert.Equal(t, gem2exists, true)
}

func Test_Olt_freeGemPortIdReplicatedflow(t *testing.T) {
	const (
		pon   = 1
		onu   = 1
		uni   = 16
		gem1  = 1024
		gem2  = 1025
		flow1 = 1
	)

	numPon := 2
	numOnu := 2

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	olt.GemPortIDs[pon][onu][uni] = make(map[int32]map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem1] = make(map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem1][flow1] = true
	olt.GemPortIDs[pon][onu][uni][gem2] = make(map[uint64]bool)
	olt.GemPortIDs[pon][onu][uni][gem2][flow1] = true

	// this flow was a replicated flow, remove all the gems that are referenced by that flow
	flowMultiGem := &openolt.Flow{
		FlowId: flow1,
	}

	olt.freeGemPortId(flowMultiGem)

	// this flow removes all the gems, so no UNI should be left
	assert.Equal(t, len(olt.GemPortIDs[pon][onu][uni]), 0)
}

// testing that we can validate flows while we are adding them
func Benchmark_validateAndAddFlows(b *testing.B) {
	const (
		pon   = 0
		start = 0
		end   = 512
	)

	for r := 0; r < b.N; r++ {
		olt := createMockOlt(1, 512, 4, []ServiceIf{})

		wg := sync.WaitGroup{}

		// concurrently adding 1K gems
		for i := start; i < end; i++ {
			wg.Add(1)
			flow := &openolt.Flow{
				AccessIntfId: pon,
				OnuId:        int32(i),
				PortNo:       uint32(i),
				GemportId:    int32(i),
				FlowId:       uint64(i),
			}
			go func(wg *sync.WaitGroup) {
				olt.storeGemPortIdByFlow(flow)
				olt.storeAllocId(flow)
				wg.Done()
			}(&wg)
		}

		// at the same time validate flows
		for i := start; i < end; i++ {
			wg.Add(1)
			flow := &openolt.Flow{
				AccessIntfId: pon,
				OnuId:        int32(i),
				PortNo:       uint32(i),
				GemportId:    1,
				FlowId:       uint64(i),
			}
			go func(wg *sync.WaitGroup) {
				_ = olt.validateFlow(flow)
				wg.Done()
			}(&wg)
		}

		wg.Wait()
		// NOTE this tests only fails if there is concurrent access to the map
	}
}

func Test_Olt_validateFlow(t *testing.T) {

	const (
		pon0            = 0
		pon1            = 1
		onu0            = 0
		onu1            = 1
		uniPort         = 0
		usedGemIdPon0   = 1024
		usedGemIdPon1   = 1025
		usedAllocIdPon0 = 1
		usedAllocIdPon1 = 2
		flowId          = 1
	)

	numPon := 2
	numOnu := 2

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	olt.GemPortIDs[pon0][onu0][uniPort] = make(map[int32]map[uint64]bool)
	olt.GemPortIDs[pon1][onu0][uniPort] = make(map[int32]map[uint64]bool)

	olt.GemPortIDs[pon0][onu0][uniPort][usedGemIdPon0] = make(map[uint64]bool)
	olt.GemPortIDs[pon0][onu0][uniPort][usedGemIdPon0][flowId] = true
	olt.GemPortIDs[pon1][onu0][uniPort][usedGemIdPon1] = make(map[uint64]bool)
	olt.GemPortIDs[pon1][onu0][uniPort][usedGemIdPon1][flowId] = true

	olt.AllocIDs[pon0][onu0][uniPort] = make(map[int32]map[uint64]bool)
	olt.AllocIDs[pon1][onu0][uniPort] = make(map[int32]map[uint64]bool)
	olt.AllocIDs[pon0][onu0][uniPort][usedAllocIdPon0] = make(map[uint64]bool)
	olt.AllocIDs[pon0][onu0][uniPort][usedAllocIdPon0][flowId] = true
	olt.AllocIDs[pon1][onu0][uniPort][usedAllocIdPon1] = make(map[uint64]bool)
	olt.AllocIDs[pon1][onu0][uniPort][usedAllocIdPon1][flowId] = true

	// a GemPortID can be referenced across multiple flows on the same ONU
	validGemFlow := &openolt.Flow{
		AccessIntfId: pon0,
		OnuId:        onu0,
		GemportId:    usedGemIdPon0,
	}

	err := olt.validateFlow(validGemFlow)
	assert.NoError(t, err)

	// a GemPortID can NOT be referenced across different ONUs on the same PON
	invalidGemFlow := &openolt.Flow{
		AccessIntfId: pon0,
		OnuId:        onu1,
		GemportId:    usedGemIdPon0,
	}
	err = olt.validateFlow(invalidGemFlow)
	assert.Error(t, err, "gem-1024-already-in-use-on-uni-0-onu-0")

	// if a flow reference the same GEM on a different PON it's a valid flow
	invalidGemDifferentPonFlow := &openolt.Flow{
		AccessIntfId: pon1,
		OnuId:        onu1,
		GemportId:    usedGemIdPon0,
	}
	err = olt.validateFlow(invalidGemDifferentPonFlow)
	assert.NoError(t, err)

	// an allocId can be referenced across multiple flows on the same ONU
	validAllocFlow := &openolt.Flow{
		AccessIntfId: pon0,
		OnuId:        onu0,
		AllocId:      usedAllocIdPon0,
	}
	err = olt.validateFlow(validAllocFlow)
	assert.NoError(t, err)

	// an allocId can NOT be referenced across different ONUs on the same PON
	invalidAllocFlow := &openolt.Flow{
		AccessIntfId: pon0,
		OnuId:        onu1,
		AllocId:      usedAllocIdPon0,
	}
	err = olt.validateFlow(invalidAllocFlow)
	assert.Error(t, err, "allocId-1-already-in-use-on-uni-0-onu-0")

	// if a flow reference the same AllocId on a different PON it's a valid flow
	invalidAllocDifferentPonFlow := &openolt.Flow{
		AccessIntfId: pon1,
		OnuId:        onu1,
		AllocId:      usedAllocIdPon0,
	}
	err = olt.validateFlow(invalidAllocDifferentPonFlow)
	assert.NoError(t, err)
}

func Test_Olt_validateReplicatedFlow(t *testing.T) {

	const (
		pon0            = 0
		onu0            = 0
		onu1            = 1
		uniPort         = 0
		usedGemId1      = 1024
		usedGemId2      = 1025
		usedAllocIdPon0 = 1
		flowId          = 1
	)

	numPon := 1
	numOnu := 1

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	// both the gemports referenced in this flow are already allocated
	olt.GemPortIDs[pon0][onu0][uniPort] = make(map[int32]map[uint64]bool)
	olt.GemPortIDs[pon0][onu0][uniPort][usedGemId1] = make(map[uint64]bool)
	olt.GemPortIDs[pon0][onu0][uniPort][usedGemId1][flowId] = true
	olt.GemPortIDs[pon0][onu0][uniPort][usedGemId2] = make(map[uint64]bool)
	olt.GemPortIDs[pon0][onu0][uniPort][usedGemId2][flowId] = true

	olt.AllocIDs[pon0][onu0][uniPort] = make(map[int32]map[uint64]bool)
	olt.AllocIDs[pon0][onu0][uniPort][usedAllocIdPon0] = make(map[uint64]bool)
	olt.AllocIDs[pon0][onu0][uniPort][usedAllocIdPon0][flowId] = true

	pbitToGemPortMap := make(map[uint32]uint32)
	pbitToGemPortMap[0] = usedGemId1
	pbitToGemPortMap[1] = usedGemId2

	// this flow should fail vlidation as the gems are already allocated to Onu0
	invalidGemFlow := &openolt.Flow{
		AccessIntfId:  pon0,
		OnuId:         onu1,
		PortNo:        uniPort,
		GemportId:     0,
		ReplicateFlow: true,
		PbitToGemport: pbitToGemPortMap,
	}

	err := olt.validateFlow(invalidGemFlow)
	assert.NotNil(t, err)

	// PbitToGemport is a map, so any of the two gemPorts can fail first and determine the error message
	foundError := false
	switch err.Error() {
	case fmt.Sprintf("gem-%d-already-in-use-on-uni-%d-onu-%d-replicated-flow-%d", usedGemId2, uniPort, onu0, invalidGemFlow.FlowId):
		foundError = true
	case fmt.Sprintf("gem-%d-already-in-use-on-uni-%d-onu-%d-replicated-flow-%d", usedGemId1, uniPort, onu0, invalidGemFlow.FlowId):
		foundError = true

	}
	assert.True(t, foundError)
}

func Test_Olt_OmciMsgOut(t *testing.T) {
	numPon := 4
	numOnu := 4

	olt := createMockOlt(numPon, numOnu, 1, []ServiceIf{})

	// a malformed packet should return an error
	msg := &openolt.OmciMsg{
		IntfId: 1,
		OnuId:  1,
		Pkt:    []byte{},
	}
	ctx := context.TODO()
	_, err := olt.OmciMsgOut(ctx, msg)
	assert.Error(t, err, "olt-received-malformed-omci-packet")

	// a correct packet for a non exiting ONU should throw an error
	msg = &openolt.OmciMsg{
		IntfId: 10,
		OnuId:  25,
		Pkt:    makeOmciSetRequest(t),
	}
	_, err = olt.OmciMsgOut(ctx, msg)
	assert.Error(t, err, "Cannot find PonPort with id 10 in OLT 0")

	// a correct packet for a disabled ONU should be dropped
	// note that an error is not returned, this is valid in BBsim
	const (
		ponId = 1
		onuId = 1
	)
	pon, _ := olt.GetPonById(ponId)
	onu, _ := pon.GetOnuById(onuId)
	onu.InternalState.SetState(OnuStateDisabled)
	msg = &openolt.OmciMsg{
		IntfId: ponId,
		OnuId:  onuId,
		Pkt:    makeOmciSetRequest(t),
	}
	_, err = olt.OmciMsgOut(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, len(onu.Channel), 0) // check that no messages have been sent

	// test that the ONU receives a valid packet
	onu.InternalState.SetState(OnuStateEnabled)
	_, err = olt.OmciMsgOut(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, len(onu.Channel), 1) // check that one message have been sent

}
