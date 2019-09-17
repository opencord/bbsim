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
