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

package eapol

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
)

// MOCKS

var eapolStateMachine = fsm.NewFSM(
	StateAuthStarted,
	fsm.Events{
		{Name: EventStartSent, Src: []string{StateAuthStarted}, Dst: StateStartSent},
		{Name: EventResponseIdentitySent, Src: []string{StateStartSent}, Dst: StateResponseIdentitySent},
		{Name: EventResponseChallengeSent, Src: []string{StateResponseIdentitySent}, Dst: StateResponseChallengeSent},
		{Name: EventResponseSuccessReceived, Src: []string{StateResponseChallengeSent}, Dst: StateResponseSuccessReceived},
		{Name: EventAuthFailed, Src: []string{StateAuthStarted, StateStartSent, StateResponseIdentitySent, StateResponseChallengeSent}, Dst: StateAuthFailed},
	},
	fsm.Callbacks{},
)

// params for the function under test
var onuId uint32 = 1
var gemPortId uint32 = 1
var ponPortId uint32 = 0
var uniId uint32 = 0
var serialNumber string = "BBSM00000001"
var macAddress = net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03}
var portNo uint32 = 16
var serviceId uint32 = 0
var oltId int = 0

type mockStream struct {
	grpc.ServerStream
	CallCount int
	Calls     map[int]*openolt.PacketIndication
	fail      bool
}

func (s *mockStream) Send(ind *openolt.Indication) error {
	s.CallCount++
	if s.fail {
		return errors.New("fake-error")
	}
	s.Calls[s.CallCount] = ind.GetPktInd()
	return nil
}

// TESTS

func TestSendEapStartSuccess(t *testing.T) {
	eapolStateMachine.SetState(StateAuthStarted)

	stream := &mockStream{
		Calls: make(map[int]*openolt.PacketIndication),
		fail:  false,
	}

	if err := SendEapStart(onuId, ponPortId, serialNumber, portNo, macAddress, gemPortId, uniId, eapolStateMachine, stream); err != nil {
		t.Errorf("SendEapStart returned an error: %v", err)
		t.Fail()
	}

	assert.Equal(t, stream.CallCount, 1)
	assert.Equal(t, stream.Calls[1].PortNo, portNo)
	assert.Equal(t, stream.Calls[1].IntfId, ponPortId)
	assert.Equal(t, stream.Calls[1].IntfType, "pon")
	assert.Equal(t, stream.Calls[1].GemportId, uint32(gemPortId))

	assert.Equal(t, eapolStateMachine.Current(), StateStartSent)

}

func TestSendEapStartFailStreamError(t *testing.T) {

	eapolStateMachine.SetState(StateAuthStarted)

	stream := &mockStream{
		Calls: make(map[int]*openolt.PacketIndication),
		fail:  true,
	}

	err := SendEapStart(onuId, ponPortId, serialNumber, portNo, macAddress, gemPortId, uniId, eapolStateMachine, stream)
	if err == nil {
		t.Errorf("SendEapStart did not return an error")
		t.Fail()
	}

	assert.Equal(t, err.Error(), "fake-error")

	assert.Equal(t, eapolStateMachine.Current(), StateAuthFailed)
}

// TODO test eapol.HandleNextPacket

func TestUpdateAuthFailed(t *testing.T) {

	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	eapolStateMachine.SetState(StateAuthStarted)
	_ = updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState(StateStartSent)
	_ = updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState(StateResponseIdentitySent)
	_ = updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState(StateResponseChallengeSent)
	_ = updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState(StateResponseSuccessReceived)
	err := updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	if err == nil {
		t.Errorf("updateAuthFailed did not return an error")
		t.Fail()
	}
	assert.Equal(t, err.Error(), fmt.Sprintf("event %s inappropriate in current state %s", EventAuthFailed, StateResponseSuccessReceived))

}

func createTestEAPOLPkt(eap *layers.EAP) gopacket.Packet {
	bytes := createEAPOLPkt(eap, serviceId, uniId, onuId, ponPortId, oltId)
	return gopacket.NewPacket(bytes, layers.LayerTypeEthernet, gopacket.Default)
}

func handleTestEAPOLPkt(pkt gopacket.Packet, stream types.Stream) {
	HandleNextPacket(onuId, ponPortId, gemPortId, serialNumber, portNo, uniId, serviceId, oltId, eapolStateMachine, pkt, stream, nil)
}

func TestDropUnexpectedPackets(t *testing.T) {
	stream := &mockStream{
		Calls: make(map[int]*openolt.PacketIndication),
	}

	const eapId uint8 = 1

	//Create test packets
	identityRequest := createEAPIdentityRequest(eapId)
	challangeRequest := createEAPChallengeRequest(eapId, []byte{0x10})
	success := createEAPSuccess(eapId)

	identityPkt := createTestEAPOLPkt(identityRequest)
	challangePkt := createTestEAPOLPkt(challangeRequest)
	successPkt := createTestEAPOLPkt(success)

	testStates := map[string]struct {
		packets          []gopacket.Packet
		destinationState string
	}{
		//All packet should be dropped in state auth_started
		StateAuthStarted: {[]gopacket.Packet{identityPkt, challangePkt, successPkt}, StateAuthFailed},
		//Only the identity request packet should be handled in state eap_start_sent
		StateStartSent: {[]gopacket.Packet{challangePkt, successPkt}, StateAuthFailed},
		//Only the challange request packet should be handled in state eap_response_identity_sent
		StateResponseIdentitySent: {[]gopacket.Packet{identityPkt, successPkt}, StateAuthFailed},
		//Only the success packet should be handled in state eap_response_challenge_sent
		StateResponseChallengeSent: {[]gopacket.Packet{identityPkt, challangePkt}, StateAuthFailed},
		//All packet should be dropped in state eap_response_success_received
		StateResponseSuccessReceived: {[]gopacket.Packet{identityPkt, challangePkt, successPkt}, StateResponseSuccessReceived},
		//All packet should be dropped in state auth_failed
		StateAuthFailed: {[]gopacket.Packet{identityPkt, challangePkt}, StateAuthFailed},
	}

	for s, info := range testStates {
		for _, p := range info.packets {
			eapolStateMachine.SetState(s)
			handleTestEAPOLPkt(p, stream)

			//No response should be sent
			assert.Equal(t, stream.CallCount, 0)
			//The state machine should transition to the failed state
			assert.Equal(t, eapolStateMachine.Current(), info.destinationState)
		}
	}
}
