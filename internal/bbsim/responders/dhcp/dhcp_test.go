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

package dhcp

import (
	"errors"
	"net"
	"testing"

	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
)

// MOCKS

var dhcpStateMachine = fsm.NewFSM(
	"dhcp_started",
	fsm.Events{
		{Name: "dhcp_discovery_sent", Src: []string{"dhcp_started"}, Dst: "dhcp_discovery_sent"},
		{Name: "dhcp_request_sent", Src: []string{"dhcp_discovery_sent"}, Dst: "dhcp_request_sent"},
		{Name: "dhcp_ack_received", Src: []string{"dhcp_request_sent"}, Dst: "dhcp_ack_received"},
		{Name: "dhcp_failed", Src: []string{"dhcp_started", "dhcp_discovery_sent", "dhcp_request_sent"}, Dst: "dhcp_failed"},
	},
	fsm.Callbacks{},
)

type mockStreamSuccess struct {
	grpc.ServerStream
	CallCount int
	Calls     map[int]*openolt.PacketIndication
	fail      bool
}

func (s *mockStreamSuccess) Send(ind *openolt.Indication) error {
	s.CallCount++
	if s.fail {
		return errors.New("fake-error")
	}
	s.Calls[s.CallCount] = ind.GetPktInd()
	return nil
}

// TESTS

func TestSendDHCPDiscovery(t *testing.T) {
	dhcpStateMachine.SetState("dhcp_started")

	var onuId uint32 = 1
	var gemPortId uint32 = 1
	var ponPortId uint32 = 0
	var oltId int = 1
	var serialNumber = "BBSM00000001"
	var mac = net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(ponPortId), byte(onuId)}
	var portNo uint32 = 16

	stream := &mockStreamSuccess{
		Calls: make(map[int]*openolt.PacketIndication),
		fail:  false,
	}

	if err := SendDHCPDiscovery(oltId, ponPortId, onuId, 900, gemPortId, serialNumber, portNo, dhcpStateMachine, mac, 7, stream); err != nil {
		t.Errorf("SendDHCPDiscovery returned an error: %v", err)
		t.Fail()
	}

	assert.Equal(t, stream.CallCount, 1)
	assert.Equal(t, stream.Calls[1].PortNo, portNo)
	assert.Equal(t, stream.Calls[1].IntfId, ponPortId)
	assert.Equal(t, stream.Calls[1].IntfType, "pon")
	assert.Equal(t, stream.Calls[1].GemportId, uint32(gemPortId))

	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_discovery_sent")
}

// TODO test dhcp.HandleNextPacket

func TestUpdateDhcpFailed(t *testing.T) {

	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	dhcpStateMachine.SetState("dhcp_started")
	_ = updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_discovery_sent")
	_ = updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_request_sent")
	_ = updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_ack_received")
	err := updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	if err == nil {
		t.Errorf("updateDhcpFailed did not return an error")
		t.Fail()
	}
	assert.Equal(t, err.Error(), "event dhcp_failed inappropriate in current state dhcp_ack_received")

}
