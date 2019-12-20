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
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"gotest.tools/assert"
	"sync"
	"testing"
	"time"
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
// and the GemPort has already been configured
// it transition to auth_started state
func Test_HandleFlowUpdateEapolWithGem(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, true, false)

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
// no action is taken (this is independent of GemPort status
func Test_HandleFlowUpdateEapolWrongUNI(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, true, false)

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
// and the GemPort has not yet been configured
// it transition to auth_started state
func Test_HandleFlowUpdateEapolWithoutGem(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900, true, false)
	onu.GemPortAdded = false

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

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)

		// emulate the addition of a GemPort
		for _, ch := range onu.GemPortChannels {
			ch <- true
		}

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, onu.InternalState.Current(), "auth_started")
	}(&wg)
	wg.Wait()

}

// validates that when an ONU receives an EAPOL flow for UNI 0
// but the noAuth bit is set no action is taken
func Test_HandleFlowUpdateEapolNoAuth(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, false, false)

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

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 0
// and the GemPort has already been configured
// it transition to dhcp_started state
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
			OPbits:  0,
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

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 255
// and the GemPort has already been configured
// it transition to dhcp_started state
func Test_HandleFlowUpdateDhcpPBit255(t *testing.T) {
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
			OPbits:  255,
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

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has already been configured
// it ignores the message
func Test_HandleFlowUpdateDhcpIgnoreByPbit(t *testing.T) {
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
			OPbits:  1,
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
	assert.Equal(t, onu.DhcpFlowReceived, false)
}

// validates that when an ONU receives a DHCP flow for UNI 0
// but the noDchp bit is set no action is taken
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
	assert.Equal(t, onu.DhcpFlowReceived, false)
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has not already been configured
// it transition to dhcp_started state
func Test_HandleFlowUpdateDhcpWithoutGem(t *testing.T) {
	// NOTE that this feature is required when we do DHCP with no eapol
	// as the DHCP flow will be the first one received
	onu := createMockOnu(1, 1, 900, 900, false, true)

	onu.GemPortAdded = false

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
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
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

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)

		// emulate the addition of a GemPort
		for _, ch := range onu.GemPortChannels {
			ch <- true
		}

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
		assert.Equal(t, onu.DhcpFlowReceived, true)
	}(&wg)
	wg.Wait()
}
