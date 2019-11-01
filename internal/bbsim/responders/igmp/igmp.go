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

package igmp

import (
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

var GetGemPortId = omci.GetGemPortId

func SendIGMPLeaveGroupV2(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32, macAddress net.HardwareAddr, stream bbsim.Stream) error {
	log.WithFields(log.Fields{
		"OnuId": onuId,
		"SerialNumber": serialNumber,
		"PortNo": portNo,
    }).Debugf("Entered SendIGMPLeaveGroupV2")
	igmp := createIGMPV2LeaveRequestPacket()
	pkt, err := serializeIgmpPacket(ponPortId, onuId, macAddress, igmp)

	if err != nil {
		log.WithFields(log.Fields{
                        "OnuId":  onuId,
                        "IntfId": ponPortId,
                        "SerialNumber": serialNumber,
                }).Errorf("Seriliazation of igmp packet failed :  %s", err)
                return err
	}
	if pkt != nil {
	}

	gemid, err := GetGemPortId(ponPortId, onuId)
	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
                        "SerialNumber": serialNumber,
		}).Errorf("Can't retrieve GemPortId for IGMP: %s", err)
		return err
	}

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: uint32(gemid),
			Pkt:       pkt,
			PortNo:    portNo,
		},
	}
        //Sending IGMP packets
        if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
                log.Errorf("Fail to send IGMP PktInd indication for ONU: %s, IntfId: %s, SerialNumber: %s,  error: %v", onuId, ponPortId, serialNumber, err)
                return err
        }
        return nil
}

func SendIGMPMembershipReportV2(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32, macAddress net.HardwareAddr, stream bbsim.Stream) error {
	log.WithFields(log.Fields{
		"OnuId": onuId,
		"SerialNumber": serialNumber,
		"PortNo": portNo,
    }).Debugf("Entered SendIGMPMembershipReportV2")
	igmp := createIGMPV2MembershipReportPacket()
	pkt, err := serializeIgmpPacket(ponPortId, onuId, macAddress, igmp)

	if err != nil {
		log.WithFields(log.Fields{
                        "OnuId":  onuId,
                        "IntfId": ponPortId,
                        "SerialNumber": serialNumber,
                }).Errorf("Seriliazation of igmp packet failed :  %s", err)
                return err
	}
	if pkt != nil {
	}

	gemid, err := GetGemPortId(ponPortId, onuId)
	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
                        "SerialNumber": serialNumber,
		}).Errorf("Can't retrieve GemPortId for IGMP: %s", err)
		return err
	}

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: uint32(gemid),
			Pkt:       pkt,
			PortNo:    portNo,
		},
	}
        //Sending IGMP packets
        if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
                log.Errorf("Fail to send IGMP PktInd indication for ONU: %s, IntfId: %s, SerialNumber: %s,  error: %v", onuId, ponPortId, serialNumber, err)
                return err
        }
        return nil
}

//func serializeIgmpPacket(intfId uint32, onuId uint32, srcMac net.HardwareAddr, igmp *layers.IGMP) ([]byte, error) {
func createIGMPV2MembershipReportPacket() IGMP {
	igmpDefault := IGMP{
		Type:            0x16, //IGMPV2 Membership Report
		MaxResponseTime: time.Duration(1),
		Checksum:        0,
		GroupAddress:    net.IPv4(224, 0, 0, 22),
		Version:         2,
	}

	return igmpDefault
}

func createIGMPV2LeaveRequestPacket() IGMP {
	igmpDefault := IGMP{
		Type:            0x17, //IGMPV2 Leave Group
		MaxResponseTime: time.Duration(1),
		Checksum:        0,
		GroupAddress:    net.IPv4(224, 0, 0, 22),
		Version:         2,
	}

	return igmpDefault
}

func serializeIgmpPacket(intfId uint32, onuId uint32, srcMac net.HardwareAddr, igmp IGMP) ([]byte, error) {
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
		Id:       0,
		TTL:      128,
		SrcIP:    []byte{0, 0, 0, 0},
		DstIP:    []byte{224, 0, 0, 22},
		Protocol: layers.IPProtocolIGMP,
		Options:  []layers.IPv4Option{{OptionType: 148, OptionLength: 4, OptionData: make([]byte, 0)}}, //Adding router alert option
	}

	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, igmp); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return bytes, nil
}

//-----------------------------------------***********************---------------------------------
// BaseLayer is a convenience struct which implements the LayerData and
// LayerPayload functions of the Layer interface.
type BaseLayer struct {
	// Contents is the set of bytes that make up this layer.  IE: for an
	// Ethernet packet, this would be the set of bytes making up the
	// Ethernet frame.
	Contents []byte
	// Payload is the set of bytes contained by (but not part of) this
	// Layer.  Again, to take Ethernet as an example, this would be the
	// set of bytes encapsulated by the Ethernet protocol.
	Payload []byte
}

type IGMPType uint8

type IGMP struct {
	BaseLayer
	Type                    IGMPType
	MaxResponseTime         time.Duration
	Checksum                uint16
	GroupAddress            net.IP
	SupressRouterProcessing bool
	RobustnessValue         uint8
	IntervalTime            time.Duration
	SourceAddresses         []net.IP
	NumberOfGroupRecords    uint16
	NumberOfSources         uint16
	//      GroupRecords            []IGMPv3GroupRecord
	Version uint8 // IGMP protocol version
}

// SerializeTo writes the serialized form of this layer into the
// SerializationBuffer, implementing gopacket.SerializableLayer.
// See the docs for gopacket.SerializableLayer for more info.
// SerializeTo writes the serialized form of this layer into the
// SerializationBuffer, implementing gopacket.SerializableLayer.
// See the docs for gopacket.SerializableLayer for more info.
func (igmp IGMP) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	//	func (igmp *IGMP) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	log.Debugf("Serializing IGMP Packet")
	//      plen := len(igmp)
	//TODO - add  length check here
	data, err := b.PrependBytes(8915)
	if err != nil {
		return err
	}

    data[0] = byte(igmp.Type)
	data[1] = byte(igmp.MaxResponseTime)
	data[2] = 0
	data[3] = 0
	copy(data[4:8], igmp.GroupAddress.To4())
	if opts.ComputeChecksums {
        igmp.Checksum = tcpipChecksum(data, 0)
	    binary.BigEndian.PutUint16(data[2:4], igmp.Checksum)
	}
	return nil
}

// Calculate the TCP/IP checksum defined in rfc1071.  The passed-in csum is any
// initial checksum data that's already been computed.
func tcpipChecksum(data []byte, csum uint32) uint16 {
	// to handle odd lengths, we loop to length - 1, incrementing by 2, then
	// handle the last byte specifically by checking against the original
	// length.
	length := len(data) - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum)
}

func (i IGMP) LayerType() gopacket.LayerType { return layers.LayerTypeIGMP }
