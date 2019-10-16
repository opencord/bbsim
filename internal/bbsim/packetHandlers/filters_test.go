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

package packetHandlers_test

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"gotest.tools/assert"
	"net"
	"testing"
)

func Test_IsDhcpPacket_True(t *testing.T) {
	dhcp := &layers.DHCPv4{
		Operation: layers.DHCPOpReply,
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, dhcp)
	if err != nil {
		t.Fatal(err)
	}

	dhcpPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeDHCPv4, gopacket.DecodeOptions{})

	res := packetHandlers.IsDhcpPacket(dhcpPkt)
	assert.Equal(t, res, true)
}

func Test_IsDhcpPacket_False(t *testing.T) {
	eth := &layers.Ethernet{
		DstMAC: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x15, 0x16},
		SrcMAC: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x15, 0x17},
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.DecodeOptions{})

	res := packetHandlers.IsDhcpPacket(ethernetPkt)
	assert.Equal(t, res, false)
}

func Test_IsIncomingPacket_True(t *testing.T) {
	eth := &layers.IPv4{
		SrcIP: net.ParseIP("192.168.254.1"),
		DstIP: net.ParseIP("182.21.0.122"),
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeIPv4, gopacket.DecodeOptions{})

	res := packetHandlers.IsIncomingPacket(ethernetPkt)
	assert.Equal(t, res, true)
}

func Test_IsIncomingPacket_False(t *testing.T) {
	eth := &layers.IPv4{
		SrcIP: net.ParseIP("182.21.0.122"),
		DstIP: net.ParseIP("192.168.254.1"),
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeIPv4, gopacket.DecodeOptions{})

	res := packetHandlers.IsIncomingPacket(ethernetPkt)
	assert.Equal(t, res, false)
}

func Test_GetDstMacAddressFromPacket(t *testing.T) {
	dstMac := net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x15, 0x16}
	eth := &layers.Ethernet{
		DstMAC: dstMac,
		SrcMAC: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x15, 0x17},
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.DecodeOptions{})

	res, err := packetHandlers.GetDstMacAddressFromPacket(ethernetPkt)
	assert.Equal(t, err, nil)
	assert.Equal(t, res.String(), dstMac.String())
}
