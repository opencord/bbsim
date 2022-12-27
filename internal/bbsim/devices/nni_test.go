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

//https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking
package devices

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNNI(t *testing.T) {
	olt := OltDevice{
		ID: 0,
	}
	nni, err := CreateNNI(&olt)

	assert.Nil(t, err)
	assert.Equal(t, "nni", nni.Type)
	assert.Equal(t, uint32(0), nni.ID)
	assert.Equal(t, "down", nni.OperState.Current())
}

func TestSendNniPacket(t *testing.T) {

	stream := &mockStream{
		CallCount: 0,
		Calls:     make(map[int]*openolt.Indication),
		fail:      false,
		channel:   make(chan int, 10),
	}

	dhcpServer := &mockDhcpServer{
		callCount: 0,
		fail:      false,
	}

	nni := NniPort{
		Olt: &OltDevice{
			OpenoltStream: stream,
			dhcpServer:    dhcpServer,
		},
		ID: 12,
	}

	// the DHCP server is mocked, so we don't really care about the packet we send in
	pkt := createTestDhcpPacket(t)
	err := nni.handleNniPacket(pkt)
	assert.Nil(t, err)
	assert.Equal(t, stream.CallCount, 1)
	indication := stream.Calls[1].GetPktInd()
	assert.Equal(t, "nni", indication.IntfType)
	assert.Equal(t, nni.ID, indication.IntfId)
	assert.Equal(t, pkt.Data(), indication.Pkt)
}

type mockDhcpServer struct {
	callCount int
	fail      bool
}

// being a Mock I just return the same packet I got
func (s mockDhcpServer) HandleServerPacket(pkt gopacket.Packet) (gopacket.Packet, error) {
	if s.fail {
		return nil, errors.New("mocked-error")
	}
	return pkt, nil
}

func createTestDhcpPacket(t *testing.T) gopacket.Packet {
	dhcp := &layers.DHCPv4{
		Operation: layers.DHCPOpRequest,
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, dhcp)
	if err != nil {
		t.Fatal(err)
	}

	return gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeDHCPv4, gopacket.DecodeOptions{})
}
