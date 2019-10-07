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
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"net"
	"testing"
)

// MOCKS
var calledSend = 0

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
}

func (s mockStreamSuccess) Send(ind *openolt.Indication) error {
	calledSend++
	return nil
}

type mockStreamError struct {
	grpc.ServerStream
}

func (s mockStreamError) Send(ind *openolt.Indication) error {
	calledSend++
	return errors.New("stream-error")
}

// TESTS

func TestSendDHCPDiscovery(t *testing.T) {
	calledSend = 0
	dhcpStateMachine.SetState("dhcp_started")

	// Save current function and restore at the end:
	old := GetGemPortId
	defer func() { GetGemPortId = old }()

	GetGemPortId = func(intfId uint32, onuId uint32) (uint16, error) {
		return 1, nil
	}

	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"
	var mac = net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(ponPortId), byte(onuId)}

	stream := mockStreamSuccess{}

	if err := SendDHCPDiscovery(ponPortId, onuId, serialNumber, dhcpStateMachine, mac, 1, stream); err != nil {
		t.Errorf("SendDHCPDiscovery returned an error: %v", err)
		t.Fail()
	}

	assert.Equal(t, calledSend, 1)

	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_discovery_sent")
}

// TODO test dhcp.HandleNextPacket

func TestUpdateDhcpFailed(t *testing.T) {

	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	dhcpStateMachine.SetState("dhcp_started")
	updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_discovery_sent")
	updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_request_sent")
	updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	assert.Equal(t, dhcpStateMachine.Current(), "dhcp_failed")

	dhcpStateMachine.SetState("dhcp_ack_received")
	err := updateDhcpFailed(onuId, ponPortId, serialNumber, dhcpStateMachine)
	if err == nil {
		t.Errorf("updateDhcpFailed did not return an error")
		t.Fail()
	}
	assert.Equal(t, err.Error(), "event dhcp_failed inappropriate in current state dhcp_ack_received")

}
