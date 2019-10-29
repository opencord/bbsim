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
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"gotest.tools/assert"
	"testing"
)

func Test_Onu_SendEapolFlow(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, false, false)

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
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowId, onu.ID)
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowType, "downstream")
	assert.Equal(t, client.FlowAddSpy.Calls[1].PortNo, onu.ID)
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// it transition to auth_started state
func Test_HandleFlowUpdateEapolFromGem(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, true, false)

	onu.InternalState = fsm.NewFSM(
		"gem_port_added",
		fsm.Events{
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken
func Test_HandleFlowUpdateEapolFromGemIgnore(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, false, false)

	onu.InternalState = fsm.NewFSM(
		"gem_port_added",
		fsm.Events{
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(1),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "gem_port_added")
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// it transition to auth_started state
func Test_HandleFlowUpdateEapolFromEnabled(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, false, false)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "eapol_flow_received")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken
func Test_HandleFlowUpdateEapolFromEnabledIgnore(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, false, false)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(1),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// but the noAuth bit is set no action is taken
func Test_HandleFlowUpdateEapolNoAuth(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, false, false)

	onu.InternalState = fsm.NewFSM(
		"gem_port_added",
		fsm.Events{
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "gem_port_added")
}

func Test_HandleFlowUpdateDhcp(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, false, true)

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
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	assert.Equal(t, onu.DhcpFlowReceived, true)
}

func Test_HandleFlowUpdateDhcpNoDhcp(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, false, false)

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
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, onu.DhcpFlowReceived, true)
}
