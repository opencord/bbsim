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
	"gotest.tools/assert"
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

func TestService_HandleAuth_noEapol(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		false, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.NilError(t, err)

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

func TestService_HandleAuth_withEapol(t *testing.T) {
	mac := net.HardwareAddr{0x2e, 0x60, byte(1), byte(1), byte(1), byte(1)}
	onu := createMockOnu(1, 1)
	s, err := NewService("testService", mac, onu, 900, 900,
		true, false, false, 64, 0, false,
		0, 0, 0, 0)

	assert.NilError(t, err)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	s.HandleAuth(stream)

	// if the service does not need EAPOL we don't expect any packet to be generated
	assert.Equal(t, stream.CallCount, 1)

	// state should not change
	assert.Equal(t, s.EapolState.Current(), "eap_start_sent")
}
