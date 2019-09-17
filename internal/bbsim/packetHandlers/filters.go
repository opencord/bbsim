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

func IsIncomingPacket(packet gopacket.Packet) bool {
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {

		ip, _ := ipLayer.(*layers.IPv4)

		// FIXME find a better way to filter outgoing packets
		if ip.SrcIP.Equal(net.ParseIP("182.21.0.128")) {
			return true
		}
	}
	return false
}
