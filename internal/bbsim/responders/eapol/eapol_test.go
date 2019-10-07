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
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"testing"
)

type fakeStream struct {
	calledSend int
	grpc.ServerStream
}

func (s fakeStream) Send(flow *openolt.Indication) error {
	s.calledSend++
	return nil
}

func TestSendEapStart(t *testing.T) {

	// params for the function under test
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber string = "BBSM00000001"

	eapolStateMachine := fsm.NewFSM(
		"auth_started",
		fsm.Events{
			{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
			{Name: "eap_response_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_response_identity_sent"},
			{Name: "eap_response_challenge_sent", Src: []string{"eap_response_identity_sent"}, Dst: "eap_response_challenge_sent"},
			{Name: "eap_response_success_received", Src: []string{"eap_response_challenge_sent"}, Dst: "eap_response_success_received"},
		},
		fsm.Callbacks{},
	)

	stream := fakeStream{}

	if err := SendEapStart(onuId, ponPortId, serialNumber, eapolStateMachine, stream); err != nil {
		t.Errorf("SendEapStart returned an error: %v", err)
		t.Fail()
	}

	assert.Equal(t, stream.calledSend, 1)

}
