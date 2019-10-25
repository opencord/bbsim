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

package dhcp

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	log "github.com/sirupsen/logrus"
	"net"
)

var dhcpServerLogger = log.WithFields(log.Fields{
	"module": "DHCP_SERVER",
})

type AddressPool struct {
	usedAddresses      []string
	availableAddresses []string
}

func (ap *AddressPool) getNextAvailableIp() (net.IP, error) {
	ip := ap.availableAddresses[0]
	ap.availableAddresses = ap.availableAddresses[1:]
	ap.usedAddresses = append(ap.usedAddresses, ip)
	return net.ParseIP(ip), nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func createAddressPool(cidr string) (*AddressPool, error) {
	ap := AddressPool{}
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ap.availableAddresses = append(ap.availableAddresses, ip.String())
	}
	return &ap, nil
}

type DhcpServer struct {
	AddressPool *AddressPool
	macAddress  net.HardwareAddr
	ipAddress   net.IP
}

func CreateDhcpServer(cidr string) (*DhcpServer, error) {
	ap, _ := createAddressPool(cidr)
	return &DhcpServer{
		AddressPool: ap,
		macAddress:  net.HardwareAddr{0x31, 0xf5, 0x06, 0xfc, 0xd5, 0x62},
		ipAddress:   net.IP{192, 168, 254, 1},
	}, nil
}

// FIXME this packet fails in BBSim with:
// ERRO[0001] Can't find Dst MacAddress in packet           IntfId=0 IntfType=nni
// Pkt=2e607013000131f506fcd562080045100134000000008011ba56c0a8fe01c0a80000004300440120979201010600000000010000000000000000c0a8000000000000000000002e607013000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000638253630c0d626273696d2e6f6e662e6f7267370d011c02030f06770c2c2f1a792a3501023604c0a8fe01ff
// apparently because of:
// https://github.com/google/gopacket/blob/master/layers/dhcpv4.go#L588
func (d *DhcpServer) createDHCPOffer(packet gopacket.Packet) (gopacket.Packet, error) {
	macAddress, err := packetHandlers.GetSrcMacAddressFromPacket(packet)
	if err != nil {
		dhcpServerLogger.Error("Can't find mac-address in packet")
	}

	rcvDhcpLayer, err := GetDhcpLayer(packet)
	if err != nil {
		dhcpServerLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Can't get DHCP Layer from Packet: %v", err)
		return nil, err
	}

	offeredIp, _ := d.AddressPool.getNextAvailableIp()

	dhcpLayer := &layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  6,
		HardwareOpts: 0,
		Xid:          rcvDhcpLayer.Xid,
		ClientHWAddr: macAddress,
		YourClientIP: offeredIp,
	}

	defaultOpts := createDefaultOpts()

	dhcpLayer.Options = append(defaultOpts, layers.DHCPOption{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeOffer)},
		Length: 1,
	})

	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptServerID,
		Data:   d.ipAddress,
		Length: uint8(len(d.ipAddress)),
	})

	// FIXME do we need this?
	//data = []byte{0xcd, 0x28, 0xcb, 0xcc, 0x00, 0x01, 0x00, 0x01,
	//	0x23, 0xed, 0x11, 0xec, 0x4e, 0xfc, 0xcd, 0x28, byte(intfId), byte(onuId)}
	//dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
	//	Type:   layers.DHCPOptClientID,
	//	Data:   data,
	//	Length: uint8(len(data)),
	//})

	//dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
	//	Type:   layers.DHCPOptRequestIP,
	//	Data:   offeredIp,
	//	Length: uint8(len(offeredIp)),
	//})

	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       d.macAddress,
		DstMAC:       macAddress,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipLayer := &layers.IPv4{
		Version:  4,
		TOS:      0x10,
		TTL:      128,
		SrcIP:    d.ipAddress,
		DstIP:    offeredIp,
		Protocol: layers.IPProtocolUDP,
	}

	udpLayer := &layers.UDP{
		SrcPort: 67,
		DstPort: 68,
	}

	udpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, udpLayer, dhcpLayer); err != nil {
		log.Errorf("Cannot serialize DHCP response: %v", err)
	}

	pkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeDHCPv4, gopacket.DecodeOptions{})
	return pkt, nil
}

func (d *DhcpServer) HandleDhcpReq(packet gopacket.Packet) (gopacket.Packet, error) {
	dhcpLayer, err := GetDhcpLayer(packet)
	if err != nil {
		dhcpServerLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Can't get DHCP Layer from Packet: %v", err)
		return nil, err
	}
	dhcpMessageType, err := GetDhcpMessageType(dhcpLayer)
	if err != nil {
		dhcpServerLogger.WithFields(log.Fields{
			"packet": packet,
		}).Errorf("Can't get DHCP Message Type from DHCP Layer: %v", err)
		return nil, err
	}

	if dhcpMessageType == layers.DHCPMsgTypeDiscover {
		dhcpServerLogger.Warnf("Asking for offer")
		return d.createDHCPOffer(packet)

	}
	if dhcpMessageType == layers.DHCPMsgTypeRequest {
		dhcpServerLogger.Warnf("Asking for ACK")
	}
	return packet, nil
}
