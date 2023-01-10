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

	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"gotest.tools/assert"
)

// test that BBR correctly sends the EAPOL Flow
func Test_Onu_SendEapolFlow(t *testing.T) {
	onu := createMockOnu(1, 1)

	client := &mockClient{
		FlowAddSpy: FlowAddSpy{
			Calls: make(map[int]*openolt.Flow),
		},
		fail: false,
	}

	onu.sendEapolFlow(client)
	assert.Equal(t, client.FlowAddSpy.CallCount, 1)

	assert.Equal(t, client.FlowAddSpy.Calls[1].AccessIntfId, int32(onu.PonPortID))
	assert.Equal(t, client.FlowAddSpy.Calls[1].OnuId, int32(onu.ID))
	assert.Equal(t, client.FlowAddSpy.Calls[1].UniId, int32(0))
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowId, uint64(onu.ID))
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowType, flowTypeDownstream)
	assert.Equal(t, client.FlowAddSpy.Calls[1].PortNo, onu.ID)
}

// checks that the FlowId is added to the list
func Test_HandleFlowAddFlowId(t *testing.T) {
	onu := createMockOnu(1, 1)

	// add flow on the first UNI
	flow1 := openolt.Flow{
		FlowId:     64,
		Classifier: &openolt.Classifier{},
		PortNo:     onu.UniPorts[0].(*UniPort).PortNo,
	}
	msg1 := types.OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow1,
	}
	onu.handleFlowAdd(msg1)
	assert.Equal(t, len(onu.FlowIds), 1)
	assert.Equal(t, onu.FlowIds[0], uint64(64))

	// add flow on the second UNI
	flow2 := openolt.Flow{
		FlowId:     65,
		Classifier: &openolt.Classifier{},
		PortNo:     onu.UniPorts[1].(*UniPort).PortNo,
	}
	msg2 := types.OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow2,
	}
	onu.handleFlowAdd(msg2)
	assert.Equal(t, len(onu.FlowIds), 2)
	assert.Equal(t, onu.FlowIds[1], uint64(65))
}

// checks that we only remove the correct flow
func Test_HandleFlowRemoveFlowId(t *testing.T) {
	onu := createMockOnu(1, 1)

	onu.FlowIds = []uint64{1, 2, 34, 64, 92}

	flow := openolt.Flow{
		FlowId:     64,
		Classifier: &openolt.Classifier{},
	}
	msg := types.OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow,
	}
	onu.handleFlowRemove(msg)
	assert.Equal(t, len(onu.FlowIds), 4)
	assert.Equal(t, onu.FlowIds[0], uint64(1))
	assert.Equal(t, onu.FlowIds[1], uint64(2))
	assert.Equal(t, onu.FlowIds[2], uint64(34))
	assert.Equal(t, onu.FlowIds[3], uint64(92))
}

// checks that when the last flow is removed we reset the stored flags in the ONU
func Test_HandleFlowRemoveFlowId_LastFlow(t *testing.T) {
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{},
	)

	onu.FlowIds = []uint64{64}

	flow := openolt.Flow{
		FlowId:     64,
		Classifier: &openolt.Classifier{},
	}
	msg := types.OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow,
	}
	onu.handleFlowRemove(msg)
	assert.Equal(t, len(onu.FlowIds), 0)
}

func TestOnu_HhandleEAPOLStart(t *testing.T) {
	// FIXME
	t.Skip("move in the UNI")
	onu := createMockOnu(1, 1)
	hsia := mockService{Name: "hsia"}
	voip := mockService{Name: "voip"}

	//onu.Services = []ServiceIf{&hsia, &voip}

	stream := mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	onu.PonPort.Olt.OpenoltStream = &stream

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)

	// check that we call HandleAuth on all the services
	assert.Equal(t, hsia.HandleAuthCallCount, 1)
	assert.Equal(t, voip.HandleAuthCallCount, 1)
}

// TODO all the following tests needs to be moved in the Service model

// validates that when an ONU receives an EAPOL flow for UNI 0
// and the GemPort has already been configured
// it transition to auth_started state
func Test_HandleFlowAddEapolWithGem(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "start_auth", Src: []string{"enabled"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken (this is independent of GemPort status
func Test_HandleFlowAddEapolWrongUNI(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "start_auth", Src: []string{"enabled"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(1),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 0
// and the GemPort has already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcp(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"eap_response_success_received",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"eap_response_success_received"}, Dst: "dhcp_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeIPv4),
			SrcPort: uint32(68),
			DstPort: uint32(67),
			OPbits:  0,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 255
// and the GemPort has already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcpPBit255(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"eap_response_success_received",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"eap_response_success_received"}, Dst: "dhcp_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeIPv4),
			SrcPort: uint32(68),
			DstPort: uint32(67),
			OPbits:  255,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has already been configured
// it ignores the message
func Test_HandleFlowAddDhcpIgnoreByPbit(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"eap_response_success_received",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"eap_response_success_received"}, Dst: "dhcp_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeIPv4),
			SrcPort: uint32(68),
			DstPort: uint32(67),
			OPbits:  1,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
}

// validates that when an ONU receives a DHCP flow for UNI 0
// but the noDchp bit is set no action is taken
func Test_HandleFlowAddDhcpNoDhcp(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"eap_response_success_received",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"eap_response_success_received"}, Dst: "dhcp_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeIPv4),
			SrcPort: uint32(68),
			DstPort: uint32(67),
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has not already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcpWithoutGem(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	// NOTE that this feature is required as there is no guarantee that the gemport is the same
	// one we received with the EAPOL flow
	onu := createMockOnu(1, 1)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"enabled"}, Dst: "dhcp_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint64(onu.ID),
		FlowType:      flowTypeDownstream,
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeIPv4),
			SrcPort: uint32(68),
			DstPort: uint32(67),
			OPbits:  0,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := types.OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowAdd(msg)
}
