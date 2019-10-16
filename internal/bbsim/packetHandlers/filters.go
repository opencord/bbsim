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

package packetHandlers

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

func IsDhcpPacket(pkt gopacket.Packet) bool {
	if layerDHCP := pkt.Layer(layers.LayerTypeDHCPv4); layerDHCP != nil {
		return true
	}
	return false
}

// return true if the packet is coming in the OLT from the NNI port
// it uses the ack to check if the source is the one we assigned to the
// dhcp server
func IsIncomingPacket(packet gopacket.Packet) bool {
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {

		ip, _ := ipLayer.(*layers.IPv4)

		// FIXME find a better way to filter outgoing packets
		if ip.SrcIP.Equal(net.ParseIP("192.168.254.1")) {
			return true
		}
	}
	return false
}

// returns the Destination Mac Address contained in the packet
func GetDstMacAddressFromPacket(packet gopacket.Packet) (net.HardwareAddr, error) {
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)

		if eth.DstMAC != nil {
			return eth.DstMAC, nil
		}
	}
	return nil, errors.New("cant-find-mac-address")
}

// returns the Source Mac Address contained in the packet
func GetSrcMacAddressFromPacket(packet gopacket.Packet) (net.HardwareAddr, error) {
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)

		if eth.DstMAC != nil {
			return eth.SrcMAC, nil
		}
	}
	return nil, errors.New("cant-find-mac-address")
}

// returns wether it's an EAPOL or DHCP packet, error if it's none
func IsEapolOrDhcp(pkt gopacket.Packet) (PacketType, error) {
	if pkt.Layer(layers.LayerTypeEAP) != nil || pkt.Layer(layers.LayerTypeEAPOL) != nil {
		return EAPOL, nil
	} else if IsDhcpPacket(pkt) {
		return DHCP, nil
	}
	return UNKNOWN, errors.New("packet-is-neither-eapol-or-dhcp")
}
