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

package bbr

import (
	"context"
	"github.com/cboling/omci"
	"github.com/google/gopacket/layers"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	pb "github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
)

type ONUMock struct {
	OmciChannel  chan *pb.OmciIndication
	SerialNumber *pb.SerialNumber
	IntfId       uint32
	ID           uint32
	HasGemPort   bool
	STag         int
	CTag         int

	// OMCI params
	tid       uint16
	hpTid     uint16
	seqNumber uint16

	// send a message to this channel when the ONU has completed its run (DHCP Ack)
	Done chan bool
}

func GetSN(oltid int, intfid uint32, onuid uint32) *pb.SerialNumber {

	sn := new(pb.SerialNumber)

	//sn = new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}

	return sn
}

func sendOmciMsg(pktBytes []byte, intfId uint32, onuId uint32, sn *pb.SerialNumber, msgType string, client pb.OpenoltClient) {
	omciMsg := pb.OmciMsg{
		IntfId: intfId,
		OnuId:  onuId,
		Pkt:    pktBytes,
	}

	if _, err := client.OmciMsgOut(context.Background(), &omciMsg); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       intfId,
			"OnuId":        onuId,
			"SerialNumber": common.OnuSnToString(sn),
			"Pkt":          omciMsg.Pkt,
		}).Fatalf("Failed to send MIB Reset")
	}
	log.WithFields(log.Fields{
		"IntfId":       intfId,
		"OnuId":        onuId,
		"SerialNumber": common.OnuSnToString(sn),
		"Pkt":          omciMsg.Pkt,
	}).Tracef("Sent OMCI message %s", msgType)
}

func (onu *ONUMock) getNextTid(highPriority ...bool) uint16 {
	var next uint16
	if len(highPriority) > 0 && highPriority[0] {
		next = onu.hpTid
		onu.hpTid += 1
		if onu.hpTid < 0x8000 {
			onu.hpTid = 0x8000
		}
	} else {
		next = onu.tid
		onu.tid += 1
		if onu.tid >= 0x8000 {
			onu.tid = 1
		}
	}
	return next
}

func (o *ONUMock) startOmci(client pb.OpenoltClient) {

	mibReset, _ := omcilib.CreateMibResetRequest(o.getNextTid(false))
	sendOmciMsg(mibReset, o.IntfId, o.ID, o.SerialNumber, "mibReset", client)

	defer func() {
		log.WithField("onuSn", common.OnuSnToString(o.SerialNumber)).Trace("OMCI channel done done")
	}()

	go func() {
		// TODO close the channel once done
		for msg := range o.OmciChannel {
			msgType, packet := omcilib.DecodeOmci(msg.Pkt)

			log.WithFields(log.Fields{
				"IntfId":  msg.IntfId,
				"OnuId":   msg.OnuId,
				"OnuSn":   common.OnuSnToString(o.SerialNumber),
				"Pkt":     msg.Pkt,
				"msgType": msgType,
			}).Trace("ONU Receveives OMCI Msg")
			switch msgType {
			default:
				log.Fatalf("unexpected frame: %v", packet)
			case omci.MibResetResponseType:
				mibUpload, _ := omcilib.CreateMibUploadRequest(o.getNextTid(false))
				sendOmciMsg(mibUpload, o.IntfId, o.ID, o.SerialNumber, "mibUpload", client)
			case omci.MibUploadResponseType:
				mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
				sendOmciMsg(mibUploadNext, o.IntfId, o.ID, o.SerialNumber, "mibUploadNext", client)
			case omci.MibUploadNextResponseType:
				o.seqNumber++

				if o.seqNumber > 290 {
					// NOTE we are done with the MIB Upload (290 is the number of messages the omci-sim library will respond to)
					galEnet, _ := omcilib.CreateGalEnetRequest(o.getNextTid(false))
					sendOmciMsg(galEnet, o.IntfId, o.ID, o.SerialNumber, "CreateGalEnetRequest", client)
				} else {
					mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
					sendOmciMsg(mibUploadNext, o.IntfId, o.ID, o.SerialNumber, "mibUploadNext", client)
				}
			case omci.CreateResponseType:
				// NOTE Creating a GemPort,
				// BBsim actually doesn't care about the values, so we can do we want with the parameters
				// In the same way we can create a GemPort even without setting up UNIs/TConts/...
				// but we need the GemPort to trigger the state change

				if !o.HasGemPort {
					// NOTE this sends a CreateRequestType and BBSim replies with a CreateResponseType
					// thus we send this request only once
					gemReq, _ := omcilib.CreateGemPortRequest(o.getNextTid(false))
					sendOmciMsg(gemReq, o.IntfId, o.ID, o.SerialNumber, "CreateGemPortRequest", client)
					o.HasGemPort = true
				} else {
					o.sendEapolFlow(client)
				}

			}

		}
	}()
}

func (o *ONUMock) sendEapolFlow(client pb.OpenoltClient) {

	classifierProto := pb.Classifier{
		EthType: uint32(layers.EthernetTypeEAPOL),
		OVid:    4091,
	}

	actionProto := pb.Action{}

	downstreamFlow := pb.Flow{
		AccessIntfId:  int32(o.IntfId),
		OnuId:         int32(o.ID),
		UniId:         int32(0x101), // FIXME do not hardcode this (do we need to create a UNI port via OMCI, probably not)
		FlowId:        uint32(o.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		GemportId:     int32(1), // FIXME use the same value as CreateGemPortRequest PortID, do not hardcode
		Classifier:    &classifierProto,
		Action:        &actionProto,
		Priority:      int32(100),
		Cookie:        uint64(o.ID),
		PortNo:        uint32(o.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	if _, err := client.FlowAdd(context.Background(), &downstreamFlow); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.IntfId,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
		}).Fatalf("Failed to EAPOL Flow")
	}
	log.WithFields(log.Fields{
		"IntfId":       o.IntfId,
		"OnuId":        o.ID,
		"FlowId":       downstreamFlow.FlowId,
		"PortNo":       downstreamFlow.PortNo,
		"SerialNumber": common.OnuSnToString(o.SerialNumber),
	}).Info("Sent EAPOL Flow")
}

func (o *ONUMock) sendDhcpFlow(client pb.OpenoltClient) {
	classifierProto := pb.Classifier{
		EthType: uint32(layers.EthernetTypeIPv4),
		SrcPort: uint32(68),
		DstPort: uint32(67),
	}

	actionProto := pb.Action{}

	downstreamFlow := pb.Flow{
		AccessIntfId:  int32(o.IntfId),
		OnuId:         int32(o.ID),
		UniId:         int32(0x101), // FIXME do not hardcode this (do we need to create a UNI port via OMCI, probably not)
		FlowId:        uint32(o.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		GemportId:     int32(1), // FIXME use the same value as CreateGemPortRequest PortID, do not hardcode
		Classifier:    &classifierProto,
		Action:        &actionProto,
		Priority:      int32(100),
		Cookie:        uint64(o.ID),
		PortNo:        uint32(o.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	if _, err := client.FlowAdd(context.Background(), &downstreamFlow); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.IntfId,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
		}).Fatalf("Failed to send DHCP Flow")
	}
	log.WithFields(log.Fields{
		"IntfId":       o.IntfId,
		"OnuId":        o.ID,
		"FlowId":       downstreamFlow.FlowId,
		"PortNo":       downstreamFlow.PortNo,
		"SerialNumber": common.OnuSnToString(o.SerialNumber),
	}).Info("Sent DHCP Flow")
}
