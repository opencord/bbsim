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
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
)

var GetGemPortId = omci.GetGemPortId

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

func createDefaultDHCPReq(intfId uint32, onuId uint32) layers.DHCPv4 {
	return layers.DHCPv4{
		Operation:    layers.DHCPOpRequest,
		HardwareType: layers.LinkTypeEthernet,
		HardwareLen:  6,
		HardwareOpts: 0,
		Xid:          onuId,
		ClientHWAddr: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(intfId), byte(onuId)},
	}
}

func createDefaultOpts() []layers.DHCPOption {
	hostname := []byte("bbsim.onf.org")
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

func createDHCPDisc(intfId uint32, onuId uint32) *layers.DHCPv4 {
	dhcpLayer := createDefaultDHCPReq(intfId, onuId)
	defaultOpts := createDefaultOpts()
	dhcpLayer.Options = append([]layers.DHCPOption{layers.DHCPOption{
		Type:   layers.DHCPOptMessageType,
		Data:   []byte{byte(layers.DHCPMsgTypeDiscover)},
		Length: 1,
	}}, defaultOpts...)

	return &dhcpLayer
}

func createDHCPReq(intfId uint32, onuId uint32) *layers.DHCPv4 {
	dhcpLayer := createDefaultDHCPReq(intfId, onuId)
	defaultOpts := createDefaultOpts()

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
		0x23, 0xed, 0x11, 0xec, 0x4e, 0xfc, 0xcd, 0x28, 0xcb, 0xcc}
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptClientID,
		Data:   data,
		Length: uint8(len(data)),
	})

	data = []byte{182, 21, 0, byte(onuId)}
	dhcpLayer.Options = append(dhcpLayer.Options, layers.DHCPOption{
		Type:   layers.DHCPOptRequestIP,
		Data:   data,
		Length: uint8(len(data)),
	})
	return &dhcpLayer
}

func serializeDHCPPacket(intfId uint32, onuId uint32, srcMac net.HardwareAddr, dhcp *layers.DHCPv4) ([]byte, error) {
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

	udpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, udpLayer, dhcp); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return bytes, nil
}

func getDhcpLayer(pkt gopacket.Packet) (*layers.DHCPv4, error) {
	layerDHCP := pkt.Layer(layers.LayerTypeDHCPv4)
	dhcp, _ := layerDHCP.(*layers.DHCPv4)
	if dhcp == nil {
		return nil, errors.New("Failed-to-extract-DHCP-layer")
	}
	return dhcp, nil
}

func getDhcpMessageType(dhcp *layers.DHCPv4) (layers.DHCPMsgType, error) {
	for _, option := range dhcp.Options {
		if option.Type == layers.DHCPOptMessageType {
			if reflect.DeepEqual(option.Data, []byte{byte(layers.DHCPMsgTypeOffer)}) {
				return layers.DHCPMsgTypeOffer, nil
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

func sendDHCPPktIn(msg bbsim.ByteMsg, portNo uint32, stream bbsim.Stream) error {
	// FIXME unify sendDHCPPktIn and sendEapolPktIn methods
	gemid, err := GetGemPortId(msg.IntfId, msg.OnuId)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  msg.OnuId,
			"IntfId": msg.IntfId,
		}).Errorf("Can't retrieve GemPortId: %s", err)
		return err
	}
	data := &openolt.Indication_PktInd{PktInd: &openolt.PacketIndication{
		IntfType:  "pon",
		IntfId:    msg.IntfId,
		GemportId: uint32(gemid),
		Pkt:       msg.Bytes,
		PortNo:    portNo,
	}}

	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		dhcpLogger.Errorf("Fail to send DHCP PktInd indication. %v", err)
		return err
	}
	return nil
}

func sendDHCPRequest(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32, onuHwAddress net.HardwareAddr, cTag int, stream openolt.Openolt_EnableIndicationServer) error {
	dhcp := createDHCPReq(ponPortId, onuId)
	pkt, err := serializeDHCPPacket(ponPortId, onuId, onuHwAddress, dhcp)

	if err != nil {
		dhcpLogger.Errorf("Cannot serializeDHCPPacket: %s", err)
		return err
	}
	// NOTE I don't think we need to tag the packet
	//taggedPkt, err := packetHandlers.PushSingleTag(cTag, pkt)

	msg := bbsim.ByteMsg{
		IntfId: ponPortId,
		OnuId:  onuId,
		Bytes:  pkt,
	}

	if err := sendDHCPPktIn(msg, portNo, stream); err != nil {
		return err
	}
	dhcpLogger.WithFields(log.Fields{
		"OnuId":  onuId,
		"IntfId": ponPortId,
		"OnuSn":  serialNumber,
	}).Infof("DHCPDiscovery Sent")
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

func SendDHCPDiscovery(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32, onuStateMachine *fsm.FSM, onuHwAddress net.HardwareAddr, cTag int, stream bbsim.Stream) error {
	dhcp := createDHCPDisc(ponPortId, onuId)
	pkt, err := serializeDHCPPacket(ponPortId, onuId, onuHwAddress, dhcp)
	if err != nil {
		dhcpLogger.Errorf("Cannot serializeDHCPPacket: %s", err)
		return err
	}
	// NOTE I don't think we need to tag the packet
	//taggedPkt, err := packetHandlers.PushSingleTag(cTag, pkt)

	msg := bbsim.ByteMsg{
		IntfId: ponPortId,
		OnuId:  onuId,
		Bytes:  pkt,
	}

	if err := sendDHCPPktIn(msg, portNo, stream); err != nil {
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}
	dhcpLogger.WithFields(log.Fields{
		"OnuId":  onuId,
		"IntfId": ponPortId,
		"OnuSn":  serialNumber,
	}).Infof("DHCPDiscovery Sent")

	if err := onuStateMachine.Event("dhcp_discovery_sent"); err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Error while transitioning ONU State %v", err)
	}
	return nil
}

func HandleNextPacket(onuId uint32, ponPortId uint32, serialNumber string, portNo uint32, onuHwAddress net.HardwareAddr, cTag int, onuStateMachine *fsm.FSM, pkt gopacket.Packet, stream openolt.Openolt_EnableIndicationServer) error {

	dhcpLayer, err := getDhcpLayer(pkt)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Can't get DHCP Layer from Packet: %v", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}
	dhcpMessageType, err := getDhcpMessageType(dhcpLayer)
	if err != nil {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Can't get DHCP Message Type from DHCP Layer: %v", err)
		if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}

	if dhcpLayer.Operation == layers.DHCPOpReply {
		if dhcpMessageType == layers.DHCPMsgTypeOffer {
			if err := sendDHCPRequest(ponPortId, onuId, serialNumber, portNo, onuHwAddress, cTag, stream); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":  onuId,
					"IntfId": ponPortId,
					"OnuSn":  serialNumber,
				}).Errorf("Can't send DHCP Request: %s", err)
				if err := updateDhcpFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
					return err
				}
				return err
			}
			if err := onuStateMachine.Event("dhcp_request_sent"); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":  onuId,
					"IntfId": ponPortId,
					"OnuSn":  serialNumber,
				}).Errorf("Error while transitioning ONU State %v", err)
			}

		} else if dhcpMessageType == layers.DHCPMsgTypeAck {
			// NOTE once the ack is received we don't need to do anything but change the state
			if err := onuStateMachine.Event("dhcp_ack_received"); err != nil {
				dhcpLogger.WithFields(log.Fields{
					"OnuId":  onuId,
					"IntfId": ponPortId,
					"OnuSn":  serialNumber,
				}).Errorf("Error while transitioning ONU State %v", err)
			}
			dhcpLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Infof("DHCP State machine completed")
		}
		// NOTE do we need to care about DHCPMsgTypeRelease??
	} else {
		dhcpLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Warnf("Unsupported DHCP Operation: %s", dhcpLayer.Operation.String())
	}
	return nil
}
