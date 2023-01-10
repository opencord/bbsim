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

package igmp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	log "github.com/sirupsen/logrus"
)

func SendIGMPLeaveGroupV2(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32,
	gemPortId uint32, macAddress net.HardwareAddr, cTag int, pbit uint8, stream bbsim.Stream, groupAddress string) error {
	log.WithFields(log.Fields{
		"OnuId":        onuId,
		"SerialNumber": serialNumber,
		"PortNo":       portNo,
		"GroupAddress": groupAddress,
	}).Debugf("Entered SendIGMPLeaveGroupV2")
	igmp := createIGMPV2LeaveRequestPacket(groupAddress)
	pkt, err := serializeIgmpPacket(ponPortId, onuId, cTag, macAddress, pbit, igmp)

	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"IntfId":       ponPortId,
			"SerialNumber": serialNumber,
			"GroupAddress": groupAddress,
		}).Errorf("Seriliazation of igmp packet failed :  %s", err)
		return err
	}

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: gemPortId,
			Pkt:       pkt,
			PortNo:    portNo,
			OnuId:     onuId,
			UniId:     0, // FIXME: When multi-uni support comes in, this hardcoding has to be removed
		},
	}
	//Sending IGMP packets
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"SerialNumber": serialNumber,
			"PortNo":       portNo,
			"IntfId":       ponPortId,
			"GroupAddress": groupAddress,
			"err":          err,
		}).Error("Fail to send IGMP PktInd indication for ONU")
		return err
	}
	return nil
}

func SendIGMPMembershipReportV2(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32,
	gemPortId uint32, macAddress net.HardwareAddr, cTag int, pbit uint8, stream bbsim.Stream, groupAddress string) error {

	igmp := createIGMPV2MembershipReportPacket(groupAddress)
	pkt, err := serializeIgmpPacket(ponPortId, onuId, cTag, macAddress, pbit, igmp)

	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"IntfId":       ponPortId,
			"SerialNumber": serialNumber,
			"GroupAddress": groupAddress,
		}).Errorf("Seriliazation of igmp packet failed :  %s", err)
		return err
	}

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: gemPortId,
			Pkt:       pkt,
			PortNo:    portNo,
			OnuId:     onuId,
			UniId:     0,
		},
	}
	//Sending IGMP packets
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"SerialNumber": serialNumber,
			"PortNo":       portNo,
			"IntfId":       ponPortId,
			"GroupAddress": groupAddress,
			"err":          err,
		}).Errorf("Fail to send IGMP PktInd indication")
		return err
	}

	log.WithFields(log.Fields{
		"OnuId":        onuId,
		"SerialNumber": serialNumber,
		"PortNo":       portNo,
		"GroupAddress": groupAddress,
	}).Debugf("Sent SendIGMPMembershipReportV2")
	return nil
}

func SendIGMPMembershipReportV3(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32,
	gemPortId uint32, macAddress net.HardwareAddr, cTag int, pbit uint8, stream bbsim.Stream, groupAddress string) error {

	log.WithFields(log.Fields{
		"OnuId":        onuId,
		"SerialNumber": serialNumber,
		"PortNo":       portNo,
		"GroupAddress": groupAddress,
	}).Debugf("Entered SendIGMPMembershipReportV3")
	igmp := createIGMPV3MembershipReportPacket(groupAddress)
	pkt, err := serializeIgmpPacket(ponPortId, onuId, cTag, macAddress, pbit, igmp)

	if err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"IntfId":       ponPortId,
			"SerialNumber": serialNumber,
			"GroupAddress": groupAddress,
		}).Errorf("Seriliazation of igmp packet failed :  %s", err)
		return err
	}

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: gemPortId,
			Pkt:       pkt,
			PortNo:    portNo,
			OnuId:     onuId,
			UniId:     0, // FIXME: When multi-uni support comes in, this hardcoding has to be removed
		},
	}
	//Sending IGMP packets
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"IntfId":       ponPortId,
			"SerialNumber": serialNumber,
			"GroupAddress": groupAddress,
			"err":          err,
		}).Errorf("Fail to send IGMP PktInd indication")
		return err
	}
	return nil
}

func HandleNextPacket(ponPortId uint32, onuId uint32, serialNumber string, portNo uint32,
	gemPortId uint32, macAddress net.HardwareAddr, pkt gopacket.Packet, cTag int, pbit uint8, stream bbsim.Stream) error {

	igmpLayer := pkt.Layer(layers.LayerTypeIGMP)
	if igmpLayer == nil {
		log.WithFields(log.Fields{
			"OnuId":        onuId,
			"SerialNumber": serialNumber,
			"PortNo":       portNo,
			"Pkt":          hex.EncodeToString(pkt.Data()),
		}).Error("This is not an IGMP packet")
		return errors.New("packet-is-not-igmp")
	}

	log.WithFields(log.Fields{
		"Pkt": pkt.Data(),
	}).Trace("IGMP packet")

	igmp := igmpLayer.(*layers.IGMPv1or2)

	if igmp.Type == layers.IGMPMembershipQuery {
		_ = SendIGMPMembershipReportV2(ponPortId, onuId, serialNumber, portNo, gemPortId, macAddress,
			cTag, pbit, stream, igmp.GroupAddress.String())
	}

	return nil
}

func createIGMPV3MembershipReportPacket(groupAddress string) *IGMP {

	groupRecord1 := IGMPv3GroupRecord{
		Type:             IGMPv3GroupRecordType(IGMPIsIn),
		AuxDataLen:       0, // this should always be 0 as per IGMPv3 spec.
		NumberOfSources:  3,
		MulticastAddress: net.IPv4(224, 0, 0, 22),
		SourceAddresses:  []net.IP{net.IPv4(15, 14, 20, 24), net.IPv4(15, 14, 20, 26), net.IPv4(15, 14, 20, 25)},
		AuxData:          0, // NOT USED
	}

	groupRecord2 := IGMPv3GroupRecord{
		Type:             IGMPv3GroupRecordType(IGMPIsIn),
		AuxDataLen:       0, // this should always be 0 as per IGMPv3 spec.
		NumberOfSources:  2,
		MulticastAddress: net.IPv4(224, 0, 0, 25),
		SourceAddresses:  []net.IP{net.IPv4(15, 14, 20, 30), net.IPv4(15, 14, 20, 31)},
		AuxData:          0, // NOT USED
	}

	igmpDefault := &IGMP{
		Type:                    layers.IGMPMembershipReportV3, //IGMPV3 Membership Report
		MaxResponseTime:         time.Duration(1),
		Checksum:                0,
		GroupAddress:            net.ParseIP(groupAddress),
		SupressRouterProcessing: false,
		RobustnessValue:         0,
		IntervalTime:            time.Duration(1),
		SourceAddresses:         []net.IP{net.IPv4(224, 0, 0, 24)},
		NumberOfGroupRecords:    2,
		NumberOfSources:         1,
		GroupRecords:            []IGMPv3GroupRecord{groupRecord1, groupRecord2},
		Version:                 3,
	}

	return igmpDefault
}

func createIGMPV2MembershipReportPacket(groupAddress string) *IGMP {
	return &IGMP{
		Type:         layers.IGMPMembershipReportV2, //IGMPV2 Membership Report
		Checksum:     0,
		GroupAddress: net.ParseIP(groupAddress),
		Version:      2,
	}
}

func createIGMPV2LeaveRequestPacket(groupAddress string) *IGMP {
	return &IGMP{
		Type:            layers.IGMPLeaveGroup, //IGMPV2 Leave Group
		MaxResponseTime: time.Duration(1),
		Checksum:        0,
		GroupAddress:    net.ParseIP(groupAddress),
		Version:         2,
	}
}

func serializeIgmpPacket(intfId uint32, onuId uint32, cTag int, srcMac net.HardwareAddr, pbit uint8, igmp *IGMP) ([]byte, error) {
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
		DstIP:    igmp.GroupAddress,
		Protocol: layers.IPProtocolIGMP,
		Options:  []layers.IPv4Option{{OptionType: 148, OptionLength: 4, OptionData: make([]byte, 0)}}, //Adding router alert option
	}

	if err := gopacket.SerializeLayers(buffer, options, ethernetLayer, ipLayer, igmp); err != nil {
		return nil, err
	}

	untaggedPkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	taggedPkt, err := packetHandlers.PushSingleTag(cTag, untaggedPkt, pbit)

	if err != nil {
		log.Error("TagPacket")
		return nil, err
	}

	return taggedPkt.Data(), nil
}

//-----------------------------------------***********************---------------------------------

type IGMP struct {
	layers.BaseLayer
	Type                    layers.IGMPType
	MaxResponseTime         time.Duration
	Checksum                uint16
	GroupAddress            net.IP
	SupressRouterProcessing bool
	RobustnessValue         uint8
	IntervalTime            time.Duration
	SourceAddresses         []net.IP
	NumberOfGroupRecords    uint16
	NumberOfSources         uint16
	GroupRecords            []IGMPv3GroupRecord
	Version                 uint8 // IGMP protocol version
}

// IGMPv3GroupRecord stores individual group records for a V3 Membership Report message.
type IGMPv3GroupRecord struct {
	Type             IGMPv3GroupRecordType
	AuxDataLen       uint8 // this should always be 0 as per IGMPv3 spec.
	NumberOfSources  uint16
	MulticastAddress net.IP
	SourceAddresses  []net.IP
	AuxData          uint32 // NOT USED
}

type IGMPv3GroupRecordType uint8

const (
	IGMPIsIn  IGMPv3GroupRecordType = 0x01 // Type MODE_IS_INCLUDE, source addresses x
	IGMPIsEx  IGMPv3GroupRecordType = 0x02 // Type MODE_IS_EXCLUDE, source addresses x
	IGMPToIn  IGMPv3GroupRecordType = 0x03 // Type CHANGE_TO_INCLUDE_MODE, source addresses x
	IGMPToEx  IGMPv3GroupRecordType = 0x04 // Type CHANGE_TO_EXCLUDE_MODE, source addresses x
	IGMPAllow IGMPv3GroupRecordType = 0x05 // Type ALLOW_NEW_SOURCES, source addresses x
	IGMPBlock IGMPv3GroupRecordType = 0x06 // Type BLOCK_OLD_SOURCES, source addresses x
)

func (i IGMPv3GroupRecordType) String() string {
	switch i {
	case IGMPIsIn:
		return "MODE_IS_INCLUDE"
	case IGMPIsEx:
		return "MODE_IS_EXCLUDE"
	case IGMPToIn:
		return "CHANGE_TO_INCLUDE_MODE"
	case IGMPToEx:
		return "CHANGE_TO_EXCLUDE_MODE"
	case IGMPAllow:
		return "ALLOW_NEW_SOURCES"
	case IGMPBlock:
		return "BLOCK_OLD_SOURCES"
	default:
		return ""
	}
}

// SerializeTo writes the serialized form of this layer into the
// SerializationBuffer, implementing gopacket.SerializableLayer.
// See the docs for gopacket.SerializableLayer for more info.
func (igmp *IGMP) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	data, err := b.PrependBytes(8915)
	if err != nil {
		return err
	}
	if igmp.Version == 2 {
		data[0] = byte(igmp.Type)
		data[1] = byte(igmp.MaxResponseTime)
		data[2] = 0
		data[3] = 0
		copy(data[4:8], igmp.GroupAddress.To4())
		if opts.ComputeChecksums {
			igmp.Checksum = tcpipChecksum(data, 0)
			binary.BigEndian.PutUint16(data[2:4], igmp.Checksum)
		}
	} else if igmp.Version == 3 {

		data[0] = byte(igmp.Type)
		data[1] = 0
		data[2] = 0
		data[3] = 0
		data[4] = 0
		data[5] = 0
		binary.BigEndian.PutUint16(data[6:8], igmp.NumberOfGroupRecords)
		j := 8
		for i := uint16(0); i < igmp.NumberOfGroupRecords; i++ {
			data[j] = byte(igmp.GroupRecords[i].Type)
			data[j+1] = byte(0)
			binary.BigEndian.PutUint16(data[j+2:j+4], igmp.GroupRecords[i].NumberOfSources)
			copy(data[j+4:j+8], igmp.GroupRecords[i].MulticastAddress.To4())
			j = j + 8
			for m := uint16(0); m < igmp.GroupRecords[i].NumberOfSources; m++ {
				copy(data[j:(j+4)], igmp.GroupRecords[i].SourceAddresses[m].To4())
				j = j + 4
			}
		}
		if opts.ComputeChecksums {
			igmp.Checksum = tcpipChecksum(data, 0)
			binary.BigEndian.PutUint16(data[2:4], igmp.Checksum)
		}
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

func (i *IGMP) LayerType() gopacket.LayerType { return layers.LayerTypeIGMP }
func (i *IGMP) LayerContents() []byte         { return i.Contents }
func (i *IGMP) LayerPayload() []byte          { return i.Payload }
