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
	"context"
	"net"
	"testing"
	"time"

	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	Name                   string
	HandleAuthCallCount    int
	HandleDhcpCallCount    int
	HandlePacketsCallCount int
}

func (s *mockService) HandleAuth() {
	s.HandleAuthCallCount = s.HandleAuthCallCount + 1
}

func (s *mockService) HandleDhcp(pbit uint8, cTag int) {
	s.HandleDhcpCallCount = s.HandleDhcpCallCount + 1
}

func (s *mockService) HandlePackets() {
	s.HandlePacketsCallCount = s.HandlePacketsCallCount + 1
}

func (s *mockService) Initialize(stream types.Stream)   {}
func (s *mockService) UpdateStream(stream types.Stream) {}
func (s *mockService) Disable()                         {}

func createTestService(needsEapol bool, needsDchp bool) (*Service, error) {

	enableContext := context.TODO()

	mac := net.HardwareAddr{0x2e, byte(1), byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	onu.PonPort = &PonPort{}
	onu.PonPort.Olt = &OltDevice{}
	onu.PonPort.Olt.enableContext = enableContext

	uni := UniPort{ID: 1, Onu: onu}
	return NewService(0, "testService", mac, &uni, 900, 900,
		needsEapol, needsDchp, false, false, 64, 0, false, false,
		7, 7, 7, 7)
}

// test the internalState transitions
func TestService_InternalState(t *testing.T) {
	s, err := createTestService(false, false)

	assert.Nil(t, err)

	assert.Empty(t, s.PacketCh)
	s.Initialize(&mockStream{})

	// check that channels have been created
	assert.NotNil(t, s.PacketCh)
	assert.NotNil(t, s.Channel)

	// set EAPOL and DHCP states to something else
	s.EapolState.SetState(eapol.StateResponseSuccessReceived)
	s.DHCPState.SetState("dhcp_ack_received")

	s.Disable()
	// make sure the EAPOL and DHCP states have been reset after disable
	assert.Equal(t, eapol.StateCreated, s.EapolState.Current())
	assert.Equal(t, "created", s.DHCPState.Current())

	// make sure the channel have been closed
	assert.Nil(t, s.Channel)
	assert.Nil(t, s.PacketCh)
}

// make sure that if the service does not need EAPOL we're not sending any packet
func TestService_HandleAuth_noEapol(t *testing.T) {
	s, err := createTestService(false, false)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls:   make(map[int]*openolt.Indication),
		channel: make(chan int, 10),
	}
	s.Initialize(stream)

	s.HandleAuth()
	time.Sleep(1 * time.Second)

	// if the service does not need EAPOL we don't expect any packet to be generated
	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.EapolState.Current(), eapol.StateCreated)
}

// make sure that if the service does need EAPOL we're sending any packet
func TestService_HandleAuth_withEapol(t *testing.T) {
	s, err := createTestService(true, false)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	s.HandleAuth()
	time.Sleep(1 * time.Second)

	// if the service does not need EAPOL we don't expect any packet to be generated
	assert.Equal(t, stream.CallCount, 1)

	// state should not change
	assert.Equal(t, s.EapolState.Current(), eapol.StateStartSent)
}

// make sure that if the service does not need DHCP we're not sending any packet
func TestService_HandleDhcp_not_needed(t *testing.T) {
	s, err := createTestService(false, false)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	s.HandleDhcp(7, 900)
	time.Sleep(1 * time.Second)

	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "created")
}

// when we receive a DHCP flow we call HandleDhcp an all the ONU Services
// each service device whether the tag matches it's own configuration
func TestService_HandleDhcp_different_c_Tag(t *testing.T) {
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// NOTE that the c_tag is different from the one configured in the service
	s.HandleDhcp(7, 800)
	time.Sleep(1 * time.Second)

	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "created")
}

// when we receive a DHCP flow we call HandleDhcp an all the ONU Services
// each service device whether the tag matches it's own configuration
func TestService_HandleDhcp_different_pbit(t *testing.T) {
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// NOTE that the c_tag is different from the one configured in the service
	s.HandleDhcp(5, 900)
	time.Sleep(1 * time.Second)

	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "created")
}

// if PBIT is 255 it means all of them, so start DHCP if the C_TAG matches
func TestService_HandleDhcp_pbit_255(t *testing.T) {
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// NOTE that the c_tag is different from the one configured in the service
	s.HandleDhcp(255, 900)
	time.Sleep(1 * time.Second)

	assert.Equal(t, 1, stream.CallCount)
	assert.Equal(t, s.DHCPState.Current(), "dhcp_discovery_sent")
}

// make sure that if the service does need DHCP we're sending any packet
func TestService_HandleDhcp_needed(t *testing.T) {
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	s.HandleDhcp(7, 900)
	time.Sleep(1 * time.Second)

	assert.Equal(t, 1, stream.CallCount)
	assert.Equal(t, "dhcp_discovery_sent", s.DHCPState.Current())
}

// Test that if the EAPOL state machine doesn't complete in 30 seconds we
// move it to EAPOL failed
func TestService_EAPOLFailed(t *testing.T) {

	common.Config = &common.GlobalConfig{
		BBSim: common.BBSimConfig{
			AuthRetry: false,
		},
	}

	// override the default wait time
	eapolWaitTime = 500 * time.Millisecond
	s, err := createTestService(true, false)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// set to failed if timeout occurs
	_ = s.EapolState.Event(eapol.EventStartAuth)
	time.Sleep(1 * time.Second)
	assert.Equal(t, eapol.StateAuthFailed, s.EapolState.Current())

	// do not set to failed if succeeded
	s.EapolState.SetState(eapol.StateCreated)
	_ = s.EapolState.Event(eapol.EventStartAuth)
	s.EapolState.SetState(eapol.StateResponseSuccessReceived)
	time.Sleep(1 * time.Second)
	assert.Equal(t, eapol.StateResponseSuccessReceived, s.EapolState.Current())

}

func TestService_EAPOLRestart(t *testing.T) {

	common.Config = &common.GlobalConfig{
		BBSim: common.BBSimConfig{
			AuthRetry: true,
		},
	}

	eapolWaitTime = 500 * time.Millisecond
	s, err := createTestService(true, false)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// set to failed if timeout occurs
	_ = s.EapolState.Event(eapol.EventStartAuth)

	// after a second EAPOL should have failed and restarted
	time.Sleep(1 * time.Second)
	assert.Equal(t, eapol.StateStartSent, s.EapolState.Current())
}

// Test that if the DHCP state machine doesn't complete in 30 seconds we
// move it to DHCP failed
func TestService_DHCPFailed(t *testing.T) {

	common.Config = &common.GlobalConfig{
		BBSim: common.BBSimConfig{
			DhcpRetry: false,
		},
	}

	// override the default wait time
	dhcpWaitTime = 100 * time.Millisecond
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// set to failed if timeout occurs
	_ = s.DHCPState.Event("start_dhcp")
	time.Sleep(1 * time.Second)
	assert.Equal(t, "dhcp_failed", s.DHCPState.Current())

	// do not set to failed if succeeded
	s.DHCPState.SetState("created")
	_ = s.DHCPState.Event("start_dhcp")
	s.DHCPState.SetState("dhcp_ack_received")
	time.Sleep(1 * time.Second)
	assert.Equal(t, "dhcp_ack_received", s.DHCPState.Current())
}

func TestService_DHCPRestart(t *testing.T) {
	common.Config = &common.GlobalConfig{
		BBSim: common.BBSimConfig{
			DhcpRetry: true,
		},
	}

	// override the default wait time
	dhcpWaitTime = 100 * time.Millisecond
	s, err := createTestService(false, true)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}
	s.Initialize(stream)

	// set to failed if timeout occurs
	_ = s.DHCPState.Event("start_dhcp")
	time.Sleep(1 * time.Second)
	assert.Equal(t, "dhcp_discovery_sent", s.DHCPState.Current())
}
