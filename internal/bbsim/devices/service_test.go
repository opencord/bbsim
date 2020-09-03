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
	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

type mockService struct {
	Name                   string
	HandleAuthCallCount    int
	HandleDhcpCallCount    int
	HandlePacketsCallCount int
}

func (s *mockService) HandleAuth(stream types.Stream) {
	s.HandleAuthCallCount = s.HandleAuthCallCount + 1
}

func (s *mockService) HandleDhcp(stream types.Stream, cTag int) {
	s.HandleDhcpCallCount = s.HandleDhcpCallCount + 1
}

func (s *mockService) HandlePackets(stream types.Stream) {
	s.HandlePacketsCallCount = s.HandlePacketsCallCount + 1
}

func (s *mockService) Initialize() {}
func (s *mockService) Disable()    {}

// test the internalState transitions
func TestService_InternalState(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	assert.Empty(t, s.PacketCh)
	s.Initialize()

	assert.NotNil(t, s.PacketCh)

	s.Disable()
	assert.Equal(t, "created", s.EapolState.Current())
	assert.Equal(t, "created", s.DHCPState.Current())
}

// make sure that if the service does not need EAPOL we're not sending any packet
func TestService_HandleAuth_noEapol(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls:   make(map[int]*openolt.Indication),
		channel: make(chan int, 10),
	}

	s.HandleAuth(stream)

	// if the service does not need EAPOL we don't expect any packet to be generated
	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.EapolState.Current(), "created")
}

// make sure that if the service does need EAPOL we're sending any packet
func TestService_HandleAuth_withEapol(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		true, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	s.HandleAuth(stream)

	// if the service does not need EAPOL we don't expect any packet to be generated
	assert.Equal(t, stream.CallCount, 1)

	// state should not change
	assert.Equal(t, s.EapolState.Current(), "eap_start_sent")
}

// make sure that if the service does not need DHCP we're not sending any packet
func TestService_HandleDhcp_not_needed(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	s.HandleDhcp(stream, 900)

	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "created")
}

// when we receive a DHCP flow we call HandleDhcp an all the ONU Services
// each service device whether the tag matches it's own configuration
func TestService_HandleDhcp_different_c_Tag(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// NOTE that the c_tag is different from the one configured in the service
	s.HandleDhcp(stream, 800)

	assert.Equal(t, stream.CallCount, 0)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "created")
}

// make sure that if the service does need DHCP we're sending any packet
func TestService_HandleDhcp_needed(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, true, false, 64, 0, false,
		0, 0, 0, 0)

	assert.Nil(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	s.HandleDhcp(stream, 900)

	assert.Equal(t, stream.CallCount, 1)

	// state should not change
	assert.Equal(t, s.DHCPState.Current(), "dhcp_discovery_sent")
}
