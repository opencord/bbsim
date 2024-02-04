/*
 * Copyright 2018-2024 Open Networking Foundation (ONF) and the ONF Contributors

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
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	log "github.com/sirupsen/logrus"
)

type DHCPServerIf interface {
	HandleServerPacket(pkt gopacket.Packet) (gopacket.Packet, error)
}

type DHCPServer struct {
	DHCPServerMacAddress net.HardwareAddr
}

func NewDHCPServer() *DHCPServer {
	return &DHCPServer{
		// NOTE we may need to make this configurable in case we'll need multiple servers
		DHCPServerMacAddress: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
	}
}

func (s *DHCPServer) getClientMacAddress(pkt gopacket.Packet) (net.HardwareAddr, error) {
	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		return nil, err
	}
	return dhcpLayer.ClientHWAddr, nil
}

func (s *DHCPServer) getTxId(pkt gopacket.Packet) (uint32, error) {
	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		return 0, err
	}
	return dhcpLayer.Xid, nil
}

func (s *DHCPServer) getPacketHostName(pkt gopacket.Packet) ([]byte, error) {
	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		return nil, err
	}
	for _, option := range dhcpLayer.Options {
		if option.Type == layers.DHCPOptHostname {
			return option.Data, nil
		}
	}
	return nil, errors.New("hostname-not-found")
}

func (s *DHCPServer) getOption82(pkt gopacket.Packet) ([]byte, error) {
	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		return nil, err
	}
	for _, option := range dhcpLayer.Options {
		if option.Type == 82 {
			return option.Data, nil
		}
	}
	log.WithFields(log.Fields{
		"pkt": hex.EncodeToString(pkt.Data()),
	}).Debug("option82-not-found")
	return []byte{}, nil
}

func (s *DHCPServer) createDefaultDhcpReply(xid uint32, mac net.HardwareAddr) layers.DHCPv4 {
	clientIp := s.createIpFromMacAddress(mac)
	return layers.DHCPv4{
		Operation:    layers.DHCPOpReply,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  6,
		HardwareOpts: 0,
		Xid:          xid,
		ClientHWAddr: mac,
		ClientIP:     clientIp,
		YourClientIP: clientIp,
	}
}

func (s *DHCPServer) createIpFromMacAddress(mac net.HardwareAddr) net.IP {
	ip := []byte{}
	for i := 2; i < 6; i++ {
		ip = append(ip, mac[i])
	}
	return net.IPv4(10+ip[0], ip[1], ip[2], ip[3])
}

func (s *DHCPServer) serializeServerDHCPPacket(clientMac net.HardwareAddr, dhcpLayer *layers.DHCPv4) (gopacket.Packet, error) {
	buffer := gopacket.NewSerializeBuffer()

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       s.DHCPServerMacAddress,
		DstMAC:       clientMac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipLayer := &layers.IPv4{
		Version:  4,
		TOS:      0x10,
		TTL:      128,
		SrcIP:    []byte{0, 0, 0, 0},
		DstIP:    []byte{255, 255, 255, 255},
		Protocol: layers.IPProtocolUDP,
	}

	udpLayer := &layers.UDP{
		SrcPort: 67,
		DstPort: 68,
	}

	_ = udpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, udpLayer, dhcpLayer); err != nil {
		dhcpLogger.Error("SerializeLayers")
		return nil, err
	}

	return gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default), nil

}

func (s *DHCPServer) getDefaultDhcpServerOptions(hostname []byte, option82 []byte) []layers.DHCPOption {
	defaultOpts := []layers.DHCPOption{}
	defaultOpts = append(defaultOpts, layers.DHCPOption{
		Type:   layers.DHCPOptHostname,
		Data:   hostname,
		Length: uint8(len(hostname)),
	})

	defaultOpts = append(defaultOpts, layers.DHCPOption{
		Type:   82,
		Data:   option82,
		Length: uint8(len(option82)),
	})

	return defaultOpts
}

// get a Discover packet an return a valid Offer
func (s *DHCPServer) handleDiscover(pkt gopacket.Packet) (gopacket.Packet, error) {

	oTag, iTag, err := packetHandlers.GetTagsFromPacket(pkt)
	if err != nil {
		return nil, err
	}

	clientMac, err := s.getClientMacAddress(pkt)
	if err != nil {
		return nil, err
	}

	txId, err := s.getTxId(pkt)
	if err != nil {
		return nil, err
	}

	hostname, err := s.getPacketHostName(pkt)
	if err != nil {
		return nil, err
	}

	option82, err := s.getOption82(pkt)
	if err != nil {
		return nil, err
	}

	dhcpLogger.WithFields(log.Fields{
		"oTag":      oTag,
		"iTag":      iTag,
		"clientMac": clientMac,
		"txId":      txId,
		"hostname":  string(hostname),
		"option82":  string(option82),
	}).Debug("Handling DHCP Discovery packet")

	dhcpLayer := s.createDefaultDhcpReply(txId, clientMac)
	defaultOpts := s.getDefaultDhcpServerOptions(hostname, option82)

	dhcpLayer.Options = append([]layers.DHCPOption{{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeOffer)},
		Length: 1,
	}}, defaultOpts...)

	data := []byte{01}
	data = append(data, clientMac...)
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptClientID,
		Data:   data,
		Length: uint8(len(data)),
	})

	// serialize the packet
	responsePkt, err := s.serializeServerDHCPPacket(clientMac, &dhcpLayer)
	if err != nil {
		return nil, err
	}

	var taggedResponsePkt gopacket.Packet
	if iTag != 0 { //Double tagged
		taggedResponsePkt, err = packetHandlers.PushDoubleTag(int(oTag), int(iTag), responsePkt, 0)
	} else { //Single tagged
		taggedResponsePkt, err = packetHandlers.PushSingleTag(int(oTag), responsePkt, 0)
	}

	if err != nil {
		return nil, err
	}

	return taggedResponsePkt, nil
}

func (s *DHCPServer) handleRequest(pkt gopacket.Packet) (gopacket.Packet, error) {
	oTag, iTag, err := packetHandlers.GetTagsFromPacket(pkt)
	if err != nil {
		return nil, err
	}

	clientMac, err := s.getClientMacAddress(pkt)
	if err != nil {
		return nil, err
	}

	txId, err := s.getTxId(pkt)
	if err != nil {
		return nil, err
	}

	hostname, err := s.getPacketHostName(pkt)
	if err != nil {
		return nil, err
	}

	option82, err := s.getOption82(pkt)
	if err != nil {
		return nil, err
	}

	dhcpLogger.WithFields(log.Fields{
		"oTag":      oTag,
		"iTag":      iTag,
		"clientMac": clientMac,
		"txId":      txId,
		"hostname":  string(hostname),
		"option82":  string(option82),
	}).Debug("Handling DHCP Request packet")

	dhcpLayer := s.createDefaultDhcpReply(txId, clientMac)
	defaultOpts := s.getDefaultDhcpServerOptions(hostname, option82)

	dhcpLayer.Options = append([]layers.DHCPOption{{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeAck)},
		Length: 1,
	}}, defaultOpts...)

	// TODO can we move this in getDefaultDhcpServerOptions?
	data := []byte{01}
	data = append(data, clientMac...)
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptClientID,
		Data:   data,
		Length: uint8(len(data)),
	})

	// serialize the packet
	responsePkt, err := s.serializeServerDHCPPacket(clientMac, &dhcpLayer)
	if err != nil {
		return nil, err
	}

	var taggedResponsePkt gopacket.Packet
	if iTag != 0 { //Double tagged
		taggedResponsePkt, err = packetHandlers.PushDoubleTag(int(oTag), int(iTag), responsePkt, 0)
	} else { //Single tagged
		taggedResponsePkt, err = packetHandlers.PushSingleTag(int(oTag), responsePkt, 0)
	}

	if err != nil {
		return nil, err
	}

	return taggedResponsePkt, nil
}

// HandleServerPacket is a very simple implementation of a DHCP server
// that only replies to DHCPDiscover and DHCPRequest packets
func (s DHCPServer) HandleServerPacket(pkt gopacket.Packet) (gopacket.Packet, error) {
	dhcpLayer, _ := GetDhcpLayer(pkt)

	if dhcpLayer.Operation == layers.DHCPOpReply {
		dhcpLogger.WithFields(log.Fields{
			"pkt": hex.EncodeToString(pkt.Data()),
		}).Error("Received DHCP Reply on the server. Ignoring the packet but this is a serious error.")
	}

	dhcpMessageType, _ := GetDhcpMessageType(dhcpLayer)

	switch dhcpMessageType {
	case layers.DHCPMsgTypeDiscover:
		dhcpLogger.Info("Received DHCP Discover")
		return s.handleDiscover(pkt)
	case layers.DHCPMsgTypeRequest:
		dhcpLogger.Info("Received DHCP Request")
		return s.handleRequest(pkt)
	}
	return nil, fmt.Errorf("cannot-handle-dhcp-packet-of-type-%s", dhcpMessageType.String())
}
