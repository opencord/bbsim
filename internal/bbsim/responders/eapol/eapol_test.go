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
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	originalSendEapStart func(onuId uint32, ponPortId uint32, serialNumber *openolt.SerialNumber, stream openolt.Openolt_EnableIndicationServer) error
)

type fakeStream struct {
	calledSend int
	grpc.ServerStream
}

func (s fakeStream) Send(flow *openolt.Indication) error {
	s.calledSend++
	return nil
}

func setUp()  {
	originalSendEapStart = sendEapStart
}


func tearDown()  {
	sendEapStart = originalSendEapStart
}

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestCreateWPASupplicant(t *testing.T) {

	// mocks
	mockSendEapStartCalled := 0
	mockSendEapStartArgs := struct {
		onuId uint32
		ponPortId uint32
		serialNumber *openolt.SerialNumber
		stream openolt.Openolt_EnableIndicationServer
	}{}
	mockSendEapStart := func(onuId uint32, ponPortId uint32, serialNumber *openolt.SerialNumber, stream openolt.Openolt_EnableIndicationServer) error {
		mockSendEapStartCalled++
		mockSendEapStartArgs.onuId = onuId
		mockSendEapStartArgs.ponPortId = ponPortId
		return nil
	}
	sendEapStart = mockSendEapStart

	// params for the function under test
	var onuId uint32 = 1
	var ponPortId uint32 = 0
	var serialNumber = new(openolt.SerialNumber)

	serialNumber.VendorId = []byte("BBSM")
	serialNumber.VendorSpecific = []byte{0, byte(0 % 256), byte(ponPortId), byte(onuId)}

	eapolStateMachine := fsm.NewFSM(
	"auth_started",
	fsm.Events{
		{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
		{Name: "eap_resonse_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_resonse_identity_sent"},
		{Name: "eap_resonse_challenge_sent", Src: []string{"eap_resonse_identity_sent"}, Dst: "eap_resonse_challenge_sent"},
		{Name: "eap_resonse_success_received", Src: []string{"eap_resonse_challenge_sent"}, Dst: "eap_resonse_success_received"},
	},
	fsm.Callbacks{},
	)

	pktOutCh := make(chan *bbsim.ByteMsg, 1024)

	stream := fakeStream{}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go CreateWPASupplicant(onuId, ponPortId, serialNumber, eapolStateMachine, stream, pktOutCh)
	go func(){
		time.Sleep(1 * time.Second)
		close(pktOutCh)
		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, mockSendEapStartCalled, 1)
	assert.Equal(t, mockSendEapStartArgs.onuId, onuId)
	assert.Equal(t, mockSendEapStartArgs.ponPortId, ponPortId)
}