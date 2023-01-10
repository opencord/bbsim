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

package igmp

import (
	"encoding/hex"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type mockStream struct {
	CallCount int
	Calls     map[int]*openolt.Indication
	grpc.ServerStream
}

func (s *mockStream) Send(ind *openolt.Indication) error {
	s.CallCount++
	s.Calls[s.CallCount] = ind
	return nil
}

func TestHandleNextPacket(t *testing.T) {

	t.Skip("Need to find how to serialize an IGMP packet")

	stream := &mockStream{
		CallCount: 0,
		Calls:     make(map[int]*openolt.Indication),
	}

	mac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x15, 0x16}

	packetData := []byte{
		1, 0, 94, 0, 0, 22, 222, 173, 190, 239, 186, 17, 8, 0, 70, 0, 0, 32, 0, 0, 0, 0, 120, 2, 191,
		215, 10, 244, 2, 246, 224, 0, 0, 22, 148, 4, 0, 0, 17, 10, 14, 223, 224, 0, 0, 22, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	fmt.Println(hex.EncodeToString(packetData))

	packet := gopacket.NewPacket(packetData, layers.LayerTypeIPv4, gopacket.Default)

	fmt.Println(hex.EncodeToString(packet.Data()))

	fmt.Println(packet.Layers())

	err := HandleNextPacket(0, 0, "FOO", 1, 1024, mac, packet, 55, 5, stream)
	assert.Nil(t, err)

	assert.Equal(t, 1, stream.CallCount)
}
