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

package packetHandlers_test

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/igmp"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
	"gotest.tools/assert"
	"net"
	"testing"
)

func init() {
	common.SetLogLevel(log.StandardLogger(), "error", false)
}

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

func Test_IsIgmpPacket(t *testing.T) {
	igmp := &igmp.IGMP{
		Type:         layers.IGMPMembershipReportV2, //IGMPV2 Membership Report
		Checksum:     0,
		GroupAddress: net.IPv4(224, 0, 0, 22),
		Version:      2,
	}
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	if err := gopacket.SerializeLayers(buffer, options, igmp); err != nil {
		t.Fatal(err)
	}

	pkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeIGMP, gopacket.DecodeOptions{})

	res := packetHandlers.IsIgmpPacket(pkt)
	assert.Equal(t, res, true)
}

func Test_IsLldpPacket_True(t *testing.T) {
	layer := &layers.LinkLayerDiscovery{
		PortID: layers.LLDPPortID{
			ID:      []byte{1, 2, 3, 4},
			Subtype: layers.LLDPPortIDSubtypeMACAddr,
		},
		ChassisID: layers.LLDPChassisID{
			ID:      []byte{1, 2, 3, 4},
			Subtype: layers.LLDPChassisIDSubTypeMACAddr,
		},
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, layer)
	if err != nil {
		t.Fatal(err)
	}

	packet := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeLinkLayerDiscovery, gopacket.DecodeOptions{})

	res := packetHandlers.IsLldpPacket(packet)
	assert.Equal(t, res, true)
}

func Test_IsLldpPacket_False(t *testing.T) {
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

	res := packetHandlers.IsLldpPacket(ethernetPkt)
	assert.Equal(t, res, false)
}

func Test_IsIncomingPacket_True(t *testing.T) {
	eth := &layers.DHCPv4{
		Operation: layers.DHCPOpReply,
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeDHCPv4, gopacket.DecodeOptions{})

	res := packetHandlers.IsIncomingPacket(ethernetPkt)
	assert.Equal(t, res, true)
}

func Test_IsIncomingPacket_False(t *testing.T) {
	eth := &layers.DHCPv4{
		Operation: layers.DHCPOpRequest,
	}

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buffer, opts, eth)
	if err != nil {
		t.Fatal(err)
	}

	ethernetPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeDHCPv4, gopacket.DecodeOptions{})

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
