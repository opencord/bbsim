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

// checks that the FlowId is added to the list
func Test_HandleFlowAddFlowId(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, true, false)

	flow := openolt.Flow{
		FlowId: 64,
		Classifier: &openolt.Classifier{},
	}
	msg := OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow,
	}
	onu.handleFlowAdd(msg)
	assert.Equal(t, len(onu.FlowIds), 1)
	assert.Equal(t, onu.FlowIds[0], uint32(64))
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// and the GemPort has already been configured
// it transition to auth_started state
func Test_HandleFlowAddEapolWithGem(t *testing.T) {

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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken (this is independent of GemPort status
func Test_HandleFlowAddEapolWrongUNI(t *testing.T) {

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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// and the GemPort has not yet been configured
// it transition to auth_started state
func Test_HandleFlowAddEapolWithoutGem(t *testing.T) {

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

	onu.handleFlowAdd(msg)

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
func Test_HandleFlowAddEapolNoAuth(t *testing.T) {
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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 0
// and the GemPort has already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcp(t *testing.T) {
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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	assert.Equal(t, onu.DhcpFlowReceived, true)
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit 255
// and the GemPort has already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcpPBit255(t *testing.T) {
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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	assert.Equal(t, onu.DhcpFlowReceived, true)
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has already been configured
// it ignores the message
func Test_HandleFlowAddDhcpIgnoreByPbit(t *testing.T) {
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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, onu.DhcpFlowReceived, false)
}

// validates that when an ONU receives a DHCP flow for UNI 0
// but the noDchp bit is set no action is taken
func Test_HandleFlowAddDhcpNoDhcp(t *testing.T) {
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

	onu.handleFlowAdd(msg)
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, onu.DhcpFlowReceived, false)
}

// validates that when an ONU receives a DHCP flow for UNI 0 and pbit not 0 or 255
// and the GemPort has not already been configured
// it transition to dhcp_started state
func Test_HandleFlowAddDhcpWithoutGem(t *testing.T) {
	// NOTE that this feature is required as there is no guarantee that the gemport is the same
	// one we received with the EAPOL flow
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

	onu.handleFlowAdd(msg)

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

// checks that we only remove the correct flow
func Test_HandleFlowRemoveFlowId(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, true, false)

	onu.FlowIds = []uint32{1, 2, 34, 64, 92}

	flow := openolt.Flow{
		FlowId: 64,
		Classifier: &openolt.Classifier{},
	}
	msg := OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow,
	}
	onu.handleFlowRemove(msg)
	assert.Equal(t, len(onu.FlowIds), 4)
	assert.Equal(t, onu.FlowIds[0], uint32(1))
	assert.Equal(t, onu.FlowIds[1], uint32(2))
	assert.Equal(t, onu.FlowIds[2], uint32(34))
	assert.Equal(t, onu.FlowIds[3], uint32(92))
}

// checks that when the last flow is removed we reset the stored flags in the ONU
func Test_HandleFlowRemoveFlowId_LastFlow(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900, true, false)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{},
	)

	onu.GemPortAdded = true
	onu.DhcpFlowReceived = true
	onu.EapolFlowReceived = true

	onu.FlowIds = []uint32{64}

	flow := openolt.Flow{
		FlowId: 64,
		Classifier: &openolt.Classifier{},
	}
	msg := OnuFlowUpdateMessage{
		OnuID:     onu.ID,
		PonPortID: onu.PonPortID,
		Flow:      &flow,
	}
	onu.handleFlowRemove(msg)
	assert.Equal(t, len(onu.FlowIds), 0)
	assert.Equal(t, onu.GemPortAdded, false)
	assert.Equal(t, onu.DhcpFlowReceived, false)
	assert.Equal(t, onu.EapolFlowReceived, false)
}
