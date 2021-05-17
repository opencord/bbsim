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

package omci

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	log "github.com/sirupsen/logrus"
	"strconv"
)

var omciLogger = log.WithFields(log.Fields{
	"module": "OMCI",
})

// we have a fixed number of 8 T-CONTS
var reportedTcontsMeId = []uint16{
	32769,
	32770,
	32771,
	32772,
	32773,
	32774,
	32775,
	32776,
}

// NOTE this is basically the same as https://github.com/opencord/voltha-openonu-adapter-go/blob/master/internal/pkg/onuadaptercore/omci_cc.go#L545-L564
// we should probably move it in "omci-lib-go"
func Serialize(msgType omci.MessageType, request gopacket.SerializableLayer, tid uint16) ([]byte, error) {
	omciLayer := &omci.OMCI{
		TransactionID: tid,
		MessageType:   msgType,
	}
	var options gopacket.SerializeOptions
	options.FixLengths = true

	buffer := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buffer, options, omciLayer, request)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func CreateMibResetRequest(tid uint16) ([]byte, error) {

	request := &omci.MibResetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
	}
	pkt, err := Serialize(omci.MibResetRequestType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize MibResetRequest")
		return nil, err
	}
	return HexEncode(pkt)
}

func CreateMibResetResponse(tid uint16) ([]byte, error) {

	request := &omci.MibResetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		Result: me.Success,
	}
	pkt, err := Serialize(omci.MibResetResponseType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot Serialize MibResetResponse")
		return nil, err
	}
	return pkt, nil
}

func CreateMibUploadRequest(tid uint16) ([]byte, error) {
	request := &omci.MibUploadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
			// Default Instance ID is 0
		},
	}
	pkt, err := Serialize(omci.MibUploadRequestType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize MibUploadRequest")
		return nil, err
	}
	return HexEncode(pkt)
}

func CreateMibUploadResponse(tid uint16, numberOfCommands uint16) ([]byte, error) {

	// TODO read this from ONU.MibDB
	// NumberOfCommands := uint16(291) //NOTE should this be configurable? (not until we have moved all the messages away from omci-sim)

	request := &omci.MibUploadResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		NumberOfCommands: numberOfCommands,
	}
	pkt, err := Serialize(omci.MibUploadResponseType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot Serialize MibUploadResponse")
		return nil, err
	}
	return pkt, nil
}

func CreateMibUploadNextRequest(tid uint16, seqNumber uint16) ([]byte, error) {

	request := &omci.MibUploadNextRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
			// Default Instance ID is 0
		},
		CommandSequenceNumber: seqNumber,
	}
	pkt, err := Serialize(omci.MibUploadNextRequestType, request, tid)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize MibUploadNextRequest")
		return nil, err
	}
	return HexEncode(pkt)
}

func ParseMibUploadNextRequest(omciPkt gopacket.Packet) (*omci.MibUploadNextRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeMibUploadNextRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeMibUploadNextRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.MibUploadNextRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for MibUploadNextRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateMibUploadNextResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI, mds uint8, mibDb *MibDb) ([]byte, error) {

	msgObj, err := ParseMibUploadNextRequest(omciPkt)
	if err != nil {
		err := "omci Msg layer could not be assigned for LayerTypeGetRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}

	omciLogger.WithFields(log.Fields{
		"EntityClass":           msgObj.EntityClass,
		"EntityInstance":        msgObj.EntityInstance,
		"CommandSequenceNumber": msgObj.CommandSequenceNumber,
	}).Trace("received-omci-mibUploadNext-request")

	currentEntry := mibDb.items[int(msgObj.CommandSequenceNumber)]
	reportedMe, meErr := me.LoadManagedEntityDefinition(currentEntry.classId, me.ParamData{
		EntityID:   currentEntry.entityId.ToUint16(),
		Attributes: currentEntry.params,
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("Error while generating %s: %v", currentEntry.classId.String(), meErr.Error())
	}

	if reportedMe.GetEntityID() == uint16(me.OnuDataClassID) {
		// if this is ONU-Data we need to replace the MDS
		if err := reportedMe.SetAttribute("MibDataSync", mds); err.GetError() != nil {
			omciLogger.Errorf("Error while setting mds in %s: %v", currentEntry.classId.String(), meErr.Error())
		}
	}

	response := &omci.MibUploadNextResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		ReportedME: *reportedMe,
	}

	omciLogger.WithFields(log.Fields{
		"reportedMe": reportedMe,
	}).Trace("created-omci-mibUploadNext-response")

	pkt, err := Serialize(omci.MibUploadNextResponseType, response, omciMsg.TransactionID)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot Serialize MibUploadNextRequest")
		return nil, err
	}

	return pkt, nil
}

// deprecated, use MibDB
func GeneratePriorityQueueMe(sequence uint16) (*me.ManagedEntity, me.OmciErrors) {

	iteration := sequence - 26

	// we report 256 messages for priority queues,
	// 128 downstream and 128 upstream
	// we repeat 64 times the quartet:
	// - downstream with mask
	// - downstream with attributes (same MeID as the previous)
	// - upstream with mask
	// - upstream with attributes (same MeID as the previous)

	// very ugly way to define whether the priority queue is upstream or downstream
	// (first 2 blocks are downstream, second two are upstream)
	// for example sequence 26, 27 are downstream
	// sequence 28, 29 are upstream
	isDownstream := ((sequence-26)%4 == 0) || ((sequence-27)%4 == 0)
	// entityIds are:
	// - 1 to 64 for downstream (hex 0001 to 0040)
	// - 32769 to 32832 for upstream (hex 8001 to 8040)
	var entityId uint16
	if isDownstream {
		entityId = (iteration)/4 + 1
	} else {
		entityId = (iteration)/4 + 32769
	}

	var relatedPort uint32
	if isDownstream {
		// downstream the related port is:
		// - Circuit Pack/ UNI port of UNI Port (01 01/04)
		// - priority 0 to 15 -> iteration%16

		// every 16 iteration (of 4 commands) we move to the next UNI MeID
		uniPort := int(iteration/64) + 1
		priority := (iteration / 4) % 16

		// concat the uniPort and priority in an hex string
		// we have a single circuit pack, so we hardcode it
		v := fmt.Sprintf("010%x000%x", uniPort, priority)

		// convert back to int
		k, _ := strconv.ParseInt(v, 16, 64)
		relatedPort = uint32(k)
	} else {
		// upstream the related port is:
		// - Tcont MeID -> reportedTcontsMeId[ceil(iteration/4/8)]
		// - priority 0 to 7 -> iteration%8

		// every 8 iteration (of 4 commands) we move to the next the TCONT MeID
		tcontMe := reportedTcontsMeId[int(iteration/32)]
		priority := (iteration / 4) % 8

		// concat the tcontMe and priority in an hex string
		v := fmt.Sprintf("%x000%x", tcontMe, priority)

		// convert back to int
		k, _ := strconv.ParseInt(v, 16, 64)
		relatedPort = uint32(k)
	}

	return me.NewPriorityQueue(me.ParamData{
		EntityID: entityId,
		Attributes: me.AttributeValueMap{
			"QueueConfigurationOption":                            0,
			"MaximumQueueSize":                                    100,
			"AllocatedQueueSize":                                  100,
			"DiscardBlockCounterResetInterval":                    0,
			"ThresholdValueForDiscardedBlocksDueToBufferOverflow": 0,
			"RelatedPort":                                         relatedPort,
			"TrafficSchedulerPointer":                             264,
			"Weight":                                              1,
			"BackPressureOperation":                               1,
			"BackPressureTime":                                    0,
			"BackPressureOccurQueueThreshold":                     0,
			"BackPressureClearQueueThreshold":                     0,
		},
	})
}
