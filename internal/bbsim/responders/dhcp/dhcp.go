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
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v3/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"strconv"
)

var dhcpLogger = log.WithFields(log.Fields{
	"module": "DHCP",
})

var defaultParamsRequestList = []layers.DHCPOpt{
	layers.DHCPOptSubnetMask,
	layers.DHCPOptBroadcastAddr,
	layers.DHCPOptTimeOffset,
	layers.DHCPOptRouter,
	layers.DHCPOptDomainName,
	layers.DHCPOptDNS,
	layers.DHCPOptDomainSearch,
	layers.DHCPOptHostname,
	layers.DHCPOptNetBIOSTCPNS,
	layers.DHCPOptNetBIOSTCPScope,
	layers.DHCPOptInterfaceMTU,
	layers.DHCPOptClasslessStaticRoute,
	layers.DHCPOptNTPServers,
}

func macAddressToTxId(mac net.HardwareAddr) uint32 {

	// NOTE we want to generate a unique XID,
	// the easiest way is to transform the macAddress (already unique) into an integer
	str := ""
	for _, i := range mac {
		str = str + fmt.Sprintf("%d", i)
	}
	xid, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal("Can't generate unique XID for ONU")
	}

	return uint32(xid)
}

func createDefaultDHCPReq(mac net.HardwareAddr) layers.DHCPv4 {
	xid := macAddressToTxId(mac)

	return layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  6,
		HardwareOpts: 0,
		Xid:          uint32(xid),
		ClientHWAddr: mac,
	}
}

func createDefaultOpts(gemPortId uint32, intfId uint32, onuId uint32) []layers.DHCPOption {
	hostname := []byte(fmt.Sprintf("%d.%d.%d.bbsim.onf.org", gemPortId, intfId, onuId))
	opts := []layers.DHCPOption{}
	opts = append(opts, layers.DHCPOption{
		Type:   layers.DHCPOptHostname,
		Data:   hostname,
		Length: uint8(len(hostname)),
	})

	bytes := []byte{}
	for _, option := range defaultParamsRequestList {
		bytes = append(bytes, byte(option))
	}

	opts = append(opts, layers.DHCPOption{
		Type:   layers.DHCPOptParamsRequest,
		Data:   bytes,
		Length: uint8(len(bytes)),
	})
	return opts
}

func createDHCPDisc(intfId uint32, onuId uint32, gemPort uint32, macAddress net.HardwareAddr) *layers.DHCPv4 {
	dhcpLayer := createDefaultDHCPReq(macAddress)
	defaultOpts := createDefaultOpts(gemPort, intfId, onuId)
	dhcpLayer.Options = append([]layers.DHCPOption{{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeDiscover)},
		Length: 1,
	}}, defaultOpts...)

	data := []byte{01}
	data = append(data, macAddress...)
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptClientID,
		Data:   data,
		Length: uint8(len(data)),
	})

	return &dhcpLayer
}

func createDHCPReq(intfId uint32, onuId uint32, macAddress net.HardwareAddr, offeredIp net.IP, gemPortId uint32) *layers.DHCPv4 {
	dhcpLayer := createDefaultDHCPReq(macAddress)
	defaultOpts := createDefaultOpts(gemPortId, intfId, onuId)

	dhcpLayer.Options = append(defaultOpts, layers.DHCPOption{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeRequest)},
		Length: 1,
	})

	data := []byte{182, 21, 0, 128}
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptServerID,
		Data:   data,
		Length: uint8(len(data)),
	})

	data = []byte{0xcd, 0x28, 0xcb, 0xcc, 0x00, 0x01, 0x00, 0x01,
		0x23, 0xed, 0x11, 0xec, 0x4e, 0xfc, 0xcd, 0x28, byte(intfId), byte(onuId)}
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptClientID,
		Data:   data,
		Length: uint8(len(data)),
	})

	// NOTE we should not request a specific IP, or we should request the one that has been offered
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptRequestIP,
		Data:   offeredIp,
		Length: uint8(len(offeredIp)),
	})
	return &dhcpLayer
}

func serializeDHCPPacket(intfId uint32, onuId uint32, cTag int, srcMac net.HardwareAddr, dhcp *layers.DHCPv4, pbit uint8) ([]byte, error) {
	buffer := gopacket.NewSerializeBuffer()

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       srcMac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
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
		SrcPort: 68,
		DstPort: 67,
	}

	_ = udpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, udpLayer, dhcp); err != nil {
		dhcpLogger.Error("SerializeLayers")
		return nil, err
	}

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	taggedPkt, err := packetHandlers.PushSingleTag(cTag, untaggedPkt, pbit)
	if err != nil {
		dhcpLogger.Error("TagPacket")
		return nil, err
	}

	return gopacket.Payload(taggedPkt.Data()), nil
}

func GetDhcpLayer(pkt gopacket.Packet) (*layers.DHCPv4, error) {
	layerDHCP := pkt.Layer(layers.LayerTypeDHCPv4)
	dhcp, _ := layerDHCP.(*layers.DHCPv4)
	if dhcp == nil {
		return nil, errors.New("Failed-to-extract-DHCP-layer")
	}
	return dhcp, nil
}

func GetDhcpMessageType(dhcp *layers.DHCPv4) (layers.DHCPMsgType, error) {
	for _, option := range dhcp.Options {
		if option.Type == layers.DHCPOptMessageType {
			if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeDiscover)}) {
				return layers.DHCPMsgTypeDiscover, nil
			} else if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeOffer)}) {
				return layers.DHCPMsgTypeOffer, nil
			} else if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeRequest)}) {
				return layers.DHCPMsgTypeRequest, nil
			} else if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeAck)}) {
				return layers.DHCPMsgTypeAck, nil
			} else if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeRelease)}) {
				return layers.DHCPMsgTypeRelease, nil
			} else {
				msg := fmt.Sprintf("This type %x is not supported", option.Data)
				return 0, errors.New(msg)
			}
		}
	}
	return 0, errors.New("Failed to extract MsgType from dhcp")
}

// returns the DHCP Layer type or error if it's not a DHCP Packet
func GetDhcpPacketType(pkt gopacket.Packet) (string, error) {
	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		return "", err
	}
	dhcpMessageType, err := GetDhcpMessageType(dhcpLayer)
	if err != nil {
		return "", err
	}

	return dhcpMessageType.String(), nil
}

func sendDHCPPktIn(msg bbsim.ByteMsg, portNo uint32, gemPortId uint32, stream bbsim.Stream) error {

	dhcpLogger.WithFields(log.Fields{
		"OnuId":   msg.OnuId,
		"IntfId":  msg.IntfId,
		"GemPort": gemPortId,
		"Type":    "DHCP",
		"Pkt":     hex.EncodeToString(msg.Bytes),
	}).Trace("Sending DHCP packet in")

	data := &openolt.Indication_PktInd{PktInd: &openolt.PacketIndication{
		IntfType:  "pon",
		IntfId:    msg.IntfId,
		GemportId: gemPortId,
		Pkt:       msg.Bytes,
		PortNo:    portNo,
	}}

	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		dhcpLogger.Errorf("Fail to send DHCP PktInd indication. %v", err)
		return err
	}
	return nil
}

func sendDHCPRequest(ponPortId uint32, onuId uint32, serviceName string, serialNumber string, portNo uint32,
	cTag int, gemPortId uint32, onuStateMachine *fsm.FSM, onuHwAddress net.HardwareAddr,
	offeredIp net.IP, pbit uint8, stream bbsim.Stream) error {
	dhcp := createDHCPReq(ponPortId, onuId, onuHwAddress, offeredIp, gemPortId)
	pkt, err := serializeDHCPPacket(ponPortId, onuId, cTag, onuHwAddress, dhcp, pbit)

	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":     onuId,
			"IntfId":    ponPortId,
			"OnuSn":     serialNumber,
			"OfferedIp": offeredIp.String(),
		}).Errorf("Cannot serializeDHCPPacket: %s", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}
	// NOTE I don't think we need to tag the packet
	//taggedPkt, err := packetHandlers.PushSingleTag(cTag, pkt)

	msg := bbsim.ByteMsg{
		IntfId: ponPortId,
		OnuId:  onuId,
		Bytes:  pkt,
	}

	if err := sendDHCPPktIn(msg, portNo, gemPortId, stream); err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Cannot sendDHCPPktIn: %s", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}

	dhcpLogger.WithFields(log.Fields{
		"OnuId":       onuId,
		"IntfId":      ponPortId,
		"OnuSn":       serialNumber,
		"OfferedIp":   offeredIp.String(),
		"ServiceName": serviceName,
	}).Infof("DHCPRequest Sent")
	return nil
}

func updateDhcpFailed(onuId uint32, ponPortId uint32, serialNumber string, onuStateMachine *fsm.FSM) error {
	if err := onuStateMachine.Event("dhcp_failed"); err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Error while transitioning ONU State %v", err)
		return err
	}
	return nil
}

func SendDHCPDiscovery(ponPortId uint32, onuId uint32, serviceName string, cTag int, gemPortId uint32,
	serialNumber string, portNo uint32, stateMachine *fsm.FSM, onuHwAddress net.HardwareAddr,
	pbit uint8, stream bbsim.Stream) error {

	dhcp := createDHCPDisc(ponPortId, onuId, gemPortId, onuHwAddress)
	pkt, err := serializeDHCPPacket(ponPortId, onuId, cTag, onuHwAddress, dhcp, pbit)

	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Errorf("Cannot serializeDHCPPacket: %s", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, stateMachine); err != nil {
			return err
		}
		return err
	}
	// NOTE I don't think we need to tag the packet
	//taggedPkt, err := packetHandlers.PushSingleTag(cTag, pkt)

	msg := bbsim.ByteMsg{
		IntfId: ponPortId,
		OnuId:  onuId,
		Bytes:  pkt,
	}

	if err := sendDHCPPktIn(msg, portNo, gemPortId, stream); err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Errorf("Cannot sendDHCPPktIn: %s", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, stateMachine); err != nil {
			return err
		}
		return err
	}
	dhcpLogger.WithFields(log.Fields{
		"OnuId":       onuId,
		"IntfId":      ponPortId,
		"OnuSn":       serialNumber,
		"ServiceName": serviceName,
	}).Info("DHCPDiscovery Sent")

	if err := stateMachine.Event("dhcp_discovery_sent"); err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Errorf("Error while transitioning ONU State %v", err)
		return err
	}
	return nil
}

func HandleNextPacket(onuId uint32, ponPortId uint32, serviceName string, serialNumber string, portNo uint32,
	cTag int, gemPortId uint32, onuHwAddress net.HardwareAddr, onuStateMachine *fsm.FSM,
	pkt gopacket.Packet, pbit uint8, stream bbsim.Stream) error {

	dhcpLayer, err := GetDhcpLayer(pkt)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Errorf("Can't get DHCP Layer from Packet: %v", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}
	dhcpMessageType, err := GetDhcpMessageType(dhcpLayer)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Errorf("Can't get DHCP Message Type from DHCP Layer: %v", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}

	if dhcpLayer.Operation == layers.DHCPOpReply {
		if dhcpMessageType == layers.DHCPMsgTypeOffer {
			offeredIp := dhcpLayer.YourClientIP
			if err := sendDHCPRequest(ponPortId, onuId, serviceName, serialNumber, portNo, cTag, gemPortId, onuStateMachine, onuHwAddress, offeredIp, pbit, stream); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":       onuId,
					"IntfId":      ponPortId,
					"OnuSn":       serialNumber,
					"ServiceName": serviceName,
				}).Errorf("Can't send DHCP Request: %s", err)
				if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
					return err
				}
				return err
			}
			if err := onuStateMachine.Event("dhcp_request_sent"); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":       onuId,
					"IntfId":      ponPortId,
					"OnuSn":       serialNumber,
					"ServiceName": serviceName,
				}).Errorf("Error while transitioning State %v", err)
			}

		} else if dhcpMessageType == layers.DHCPMsgTypeAck {
			// NOTE once the ack is received we don't need to do anything but change the state
			if err := onuStateMachine.Event("dhcp_ack_received"); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":       onuId,
					"IntfId":      ponPortId,
					"OnuSn":       serialNumber,
					"ServiceName": serviceName,
				}).Errorf("Error while transitioning State %v", err)
			}
			dhcpLogger.WithFields(log.Fields{
				"OnuId":       onuId,
				"IntfId":      ponPortId,
				"OnuSn":       serialNumber,
				"ServiceName": serviceName,
			}).Infof("DHCP State machine completed")
		}
		// NOTE do we need to care about DHCPMsgTypeRelease??
	} else {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":       onuId,
			"IntfId":      ponPortId,
			"OnuSn":       serialNumber,
			"ServiceName": serviceName,
		}).Warnf("Unsupported DHCP Operation: %s", dhcpLayer.Operation.String())
	}

	return nil
}

// This method handle the BBR DHCP Packets
// BBR does not need to do anything but forward the packets in the correct direction
func HandleNextBbrPacket(onuId uint32, ponPortId uint32, serialNumber string, doneChannel chan bool, pkt gopacket.Packet, client openolt.OpenoltClient) error {

	// check if the packet is going:
	// - outgouing: toward the DHCP
	// - incoming: toward the ONU
	isIncoming := packetHandlers.IsIncomingPacket(pkt)
	dhcpLogger.Tracef("Is Incoming: %t", isIncoming)

	pkt, err := packetHandlers.PopSingleTag(pkt)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"error":  err,
		}).Fatalf("Can't untag packet")
	}

	dhcpType, err := GetDhcpPacketType(pkt)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"error":  err,
		}).Fatalf("Can't find DHCP type for packet")
	}

	srcMac, _ := packetHandlers.GetSrcMacAddressFromPacket(pkt)
	dstMac, _ := packetHandlers.GetDstMacAddressFromPacket(pkt)

	if isIncoming {

		onuPacket := openolt.OnuPacket{
			IntfId:    ponPortId,
			OnuId:     onuId,
			PortNo:    onuId,
			GemportId: 1,
			Pkt:       pkt.Data(),
		}

		if _, err := client.OnuPacketOut(context.Background(), &onuPacket); err != nil {
			dhcpLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"Type":   dhcpType,
				"error":  err,
			}).Error("Failed to send DHCP packet to the ONU")
		}

		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"Type":   dhcpType,
			"DstMac": dstMac,
			"SrcMac": srcMac,
		}).Infof("Sent DHCP packet to the ONU")

		dhcpLayer, _ := GetDhcpLayer(pkt)
		dhcpMessageType, _ := GetDhcpMessageType(dhcpLayer)
		if dhcpMessageType == layers.DHCPMsgTypeAck {
			doneChannel <- true
		}

	} else {
		// double tag the packet and send it to the NNI
		// we don't really care about the tags as they are stripped before
		// the packet is sent to the DHCP server
		// in BBR we don't care about the pBit
		doubleTaggedPkt, err := packetHandlers.PushDoubleTag(900, 900, pkt, 0)
		if err != nil {
			dhcpLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"Type":   dhcpType,
				"error":  err,
			}).Error("Failed to add double tag to packet")
		}

		pkt := openolt.UplinkPacket{
			IntfId: 0, // BBSim does not care about which NNI, it has only one
			Pkt:    doubleTaggedPkt.Data(),
		}
		if _, err := client.UplinkPacketOut(context.Background(), &pkt); err != nil {
			dhcpLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"Type":   dhcpType,
				"error":  err,
			}).Error("Failed to send DHCP packet out of the NNI Port")
		}
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"Type":   dhcpType,
			"DstMac": dstMac,
			"SrcMac": srcMac,
		}).Infof("Sent DHCP packet out of the NNI Port")
	}
	return nil
}
