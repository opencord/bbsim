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

func IsIgmpPacket(pkt gopacket.Packet) bool {
	if igmpLayer := pkt.Layer(layers.LayerTypeIGMP); igmpLayer != nil {
		return true
	}
	return false
}

func IsLldpPacket(pkt gopacket.Packet) bool {
	if layer := pkt.Layer(layers.LayerTypeLinkLayerDiscovery); layer != nil {
		return true
	}
	return false
}

func IsIcmpPacket(pkt gopacket.Packet) bool {
	if layer := pkt.Layer(layers.LayerTypeICMPv6); layer != nil {
		return true
	}
	return false
}

// return true if the packet is coming in the OLT from the DHCP Server
// given that we only check DHCP packets we can use the Operation
// Request are outgoing (toward the server)
// Replies are incoming (toward the OLT)
func IsIncomingPacket(packet gopacket.Packet) bool {
	layerDHCP := packet.Layer(layers.LayerTypeDHCPv4)
	if dhcp, ok := layerDHCP.(*layers.DHCPv4); ok {
		if dhcp.Operation == layers.DHCPOpReply {
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
func GetPktType(pkt gopacket.Packet) (PacketType, error) {
	if pkt.Layer(layers.LayerTypeEAP) != nil || pkt.Layer(layers.LayerTypeEAPOL) != nil {
		return EAPOL, nil
	} else if IsDhcpPacket(pkt) {
		return DHCP, nil
	} else if IsIgmpPacket(pkt) {
		return IGMP, nil
	} else if IsIcmpPacket(pkt) {
		return ICMP, nil
	}
	return UNKNOWN, errors.New("unknown-packet-type")
}
