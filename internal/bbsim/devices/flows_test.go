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
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/stretchr/testify/assert"
)

func getTestOlt(t *testing.T, ctx context.Context, services []common.ServiceYaml) (olt *OltDevice, pon *PonPort, onu *Onu, uni *UniPort, stream *mockStream) {
	common.Services = services

	common.Config = &common.GlobalConfig{
		Olt: common.OltConfig{
			ID:          1,
			NniPorts:    1,
			PonPorts:    1,
			OnusPonPort: 1,
			UniPorts:    1,
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

	olt = CreateOLT(*common.Config, common.Services, true)

	stream = &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	olt.OpenoltStream = stream
	olt.enableContext = ctx

	if len(olt.Pons) <= 0 {
		assert.Fail(t, "No PONs on OLT")
		return
	}
	pon = olt.Pons[0]

	if len(pon.Onus) <= 0 {
		assert.Fail(t, "No ONUs on PON")
		return
	}
	onu = pon.Onus[0]

	if len(onu.UniPorts) <= 0 {
		assert.Fail(t, "No UNIs on ONU")
		return
	}
	uni = onu.UniPorts[0].(*UniPort)

	_, err := olt.ActivateOnu(ctx, &openolt.Onu{
		IntfId:       pon.ID,
		OnuId:        onu.ID,
		SerialNumber: onu.SerialNumber,
	})

	assert.Nil(t, err)

	assert.Equal(t, len(services), len(uni.Services))
	for i, s := range uni.Services {
		service := s.(*Service)
		assert.Equal(t, services[i].Name, service.Name)
		assert.Equal(t, services[i].NeedsDhcp, service.NeedsDhcp)
	}

	err = onu.InternalState.Event(OnuTxInitialize)
	assert.Nil(t, err)

	//A mock ONU won't start processing messages, so we have to call it manually
	go onu.ProcessOnuMessages(ctx, stream, nil)

	err = onu.InternalState.Event(OnuTxDiscover)
	assert.Nil(t, err)
	err = onu.InternalState.Event(OnuTxEnable)
	assert.Nil(t, err)

	err = uni.Enable()
	assert.Nil(t, err)

	return
}

func addTestFlow(t *testing.T, ctx context.Context, olt *OltDevice, onu *Onu, flow openolt.Flow) {
	//Check if the flow is correctly added
	_, err := olt.FlowAdd(ctx, &flow)
	assert.Nil(t, err)
}

func removeTestFlow(t *testing.T, ctx context.Context, olt *OltDevice, onu *Onu, flow openolt.Flow) {
	//Check if the flow is correctly removed
	_, err := olt.FlowRemove(ctx, &flow)
	assert.Nil(t, err)
}

func Test_Flows_FttbTrapRules(t *testing.T) {
	const (
		VID_VENDOR_MGMT      = 6
		VID_NETWORK_DPU_MGMT = 60

		allocId   = int32(1024)
		gemportId = int32(1024)
		uniPortNo = uint32(256)
		nniId     = 0
		nniPortNo = 0x1000000
		pbitNone  = 255

		dhcpServiceName = "dpu_dhcp"
		hsiaServicename = "hsia"
	)

	testServices := []common.ServiceYaml{
		{Name: dhcpServiceName, CTag: 60, CTagAllocation: common.TagAllocationShared.String(), STagAllocation: common.TagAllocationShared.String(), NeedsDhcp: true, TechnologyProfileID: 64},
		{Name: hsiaServicename, UniTagMatch: 4096, CTag: 4096, CTagAllocation: common.TagAllocationShared.String(), STag: 3101, STagAllocation: common.TagAllocationUnique.String(), TechnologyProfileID: 64},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	olt, pon, onu, uni, stream := getTestOlt(t, ctx, testServices)

	flows := []openolt.Flow{
		{
			AccessIntfId:  int32(pon.ID),
			OnuId:         int32(onu.ID),
			UniId:         int32(uni.ID),
			FlowId:        1,
			FlowType:      flowTypeUpstream,
			AllocId:       allocId,
			NetworkIntfId: nniId,
			GemportId:     gemportId,
			Classifier: &openolt.Classifier{
				OVid:       VID_NETWORK_DPU_MGMT,
				OPbits:     pbitNone,
				EthType:    uint32(layers.EthernetTypeIPv4),
				IpProto:    uint32(layers.IPProtocolUDP),
				SrcPort:    68,
				DstPort:    67,
				PktTagType: flowTagTypeSingle,
			},
			Action: &openolt.Action{
				Cmd: &openolt.ActionCmd{
					TrapToHost: true,
				},
			},
			Priority: 40000,
			PortNo:   uniPortNo,
		},
		{
			AccessIntfId:  -1,
			OnuId:         -1,
			UniId:         -1,
			FlowId:        2,
			FlowType:      flowTypeDownstream,
			AllocId:       -1,
			NetworkIntfId: nniId,
			GemportId:     -1,
			Classifier: &openolt.Classifier{
				OVid:       VID_NETWORK_DPU_MGMT,
				OPbits:     pbitNone,
				EthType:    uint32(layers.EthernetTypeIPv4),
				IpProto:    uint32(layers.IPProtocolUDP),
				SrcPort:    67,
				PktTagType: flowTagTypeDouble,
			},
			Action: &openolt.Action{
				Cmd: &openolt.ActionCmd{
					TrapToHost: true,
				},
			},
			Priority: 40000,
			PortNo:   nniPortNo,
		},
	}

	for _, f := range flows {
		addTestFlow(t, ctx, olt, onu, f)
	}

	//Wait a bit for messages on various channels to be processed
	time.Sleep(time.Second)

	//Check if a DHCP request is sent correctly after trap rules have been added
	assert.True(t, stream.CallCount > 0)
	dhcpCall := stream.Calls[stream.CallCount]

	switch ind := dhcpCall.Data.(type) {
	case *openolt.Indication_PktInd:
		assert.Equal(t, uniPortNo, ind.PktInd.PortNo)
		assert.Equal(t, pon.ID, ind.PktInd.IntfId)
		assert.Equal(t, "pon", ind.PktInd.IntfType)
		assert.Equal(t, uint32(gemportId), ind.PktInd.GemportId)

		packet := gopacket.NewPacket(ind.PktInd.Pkt, layers.LayerTypeEthernet, gopacket.Default)
		_, err := dhcp.GetDhcpLayer(packet)
		assert.Nil(t, err)
	default:
		assert.Fail(t, "Wrong indication type for DHCP request")
	}

	for _, f := range flows {
		removeTestFlow(t, ctx, olt, onu, f)
	}
}
