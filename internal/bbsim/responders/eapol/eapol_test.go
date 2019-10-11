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

package eapol

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

var eapolStateMachine = fsm.NewFSM(
	"auth_started",
	fsm.Events{
		{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
		{Name: "eap_response_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_response_identity_sent"},
		{Name: "eap_response_challenge_sent", Src: []string{"eap_response_identity_sent"}, Dst: "eap_response_challenge_sent"},
		{Name: "eap_response_success_received", Src: []string{"eap_response_challenge_sent"}, Dst: "eap_response_success_received"},
		{Name: "auth_failed", Src: []string{"auth_started", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent"}, Dst: "auth_failed"},
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

func TestSendEapStartSuccess(t *testing.T) {
	calledSend = 0
	eapolStateMachine.SetState("auth_started")

	// Save current function and restore at the end:
	old := GetGemPortId
	defer func() { GetGemPortId = old }()

	GetGemPortId = func(intfId uint32, onuId uint32) (uint16, error) {
		return 1, nil
	}

	// params for the function under test
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	var macAddress = net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03}

	stream := mockStreamSuccess{}

	if err := SendEapStart(onuId, ponPortId, serialNumber, macAddress, eapolStateMachine, stream); err != nil {
		t.Errorf("SendEapStart returned an error: %v", err)
		t.Fail()
	}

	assert.Equal(t, calledSend, 1)

	assert.Equal(t, eapolStateMachine.Current(), "eap_start_sent")

}

func TestSendEapStartFailNoGemPort(t *testing.T) {
	calledSend = 0
	eapolStateMachine.SetState("auth_started")

	// Save current function and restore at the end:
	old := GetGemPortId
	defer func() { GetGemPortId = old }()

	GetGemPortId = func(intfId uint32, onuId uint32) (uint16, error) {
		return 0, errors.New("no-gem-port")
	}

	// params for the function under test
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	var macAddress = net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03}

	stream := mockStreamSuccess{}

	err := SendEapStart(onuId, ponPortId, serialNumber, macAddress, eapolStateMachine, stream)
	if err == nil {
		t.Errorf("SendEapStart did not return an error")
		t.Fail()
	}

	assert.Equal(t, err.Error(), "no-gem-port")

	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")
}

func TestSendEapStartFailStreamError(t *testing.T) {
	calledSend = 0
	eapolStateMachine.SetState("auth_started")

	// Save current function and restore at the end:
	old := GetGemPortId
	defer func() { GetGemPortId = old }()

	GetGemPortId = func(intfId uint32, onuId uint32) (uint16, error) {
		return 1, nil
	}

	// params for the function under test
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber = "BBSM00000001"
	var macAddress = net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03}

	stream := mockStreamError{}

	err := SendEapStart(onuId, ponPortId, serialNumber, macAddress, eapolStateMachine, stream)
	if err == nil {
		t.Errorf("SendEapStart did not return an error")
		t.Fail()
	}

	assert.Equal(t, err.Error(), "stream-error")

	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")
}

// TODO test eapol.HandleNextPacket

func TestUpdateAuthFailed(t *testing.T) {

	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	eapolStateMachine.SetState("auth_started")
	updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState("eap_start_sent")
	updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState("eap_response_identity_sent")
	updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState("eap_response_challenge_sent")
	updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	assert.Equal(t, eapolStateMachine.Current(), "auth_failed")

	eapolStateMachine.SetState("eap_response_success_received")
	err := updateAuthFailed(onuId, ponPortId, serialNumber, eapolStateMachine)
	if err == nil {
		t.Errorf("updateAuthFailed did not return an error")
		t.Fail()
	}
	assert.Equal(t, err.Error(), "event auth_failed inappropriate in current state eap_response_success_received")

}
