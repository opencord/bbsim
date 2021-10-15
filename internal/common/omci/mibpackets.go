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
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	log "github.com/sirupsen/logrus"
)

var omciLogger = log.WithFields(log.Fields{
	"module": "OMCI",
})

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

	if msgObj.CommandSequenceNumber > mibDb.NumberOfCommands {
		omciLogger.WithFields(log.Fields{
			"CommandSequenceNumber": msgObj.CommandSequenceNumber,
			"MibDbNumberOfCommands": mibDb.NumberOfCommands,
		}).Error("mibdb-does-not-contain-item")
		return nil, fmt.Errorf("mibdb-does-not-contain-item")
	}
	currentEntry := mibDb.items[int(msgObj.CommandSequenceNumber)]
	reportedMe, meErr := me.LoadManagedEntityDefinition(currentEntry.classId, me.ParamData{
		EntityID:   currentEntry.entityId.ToUint16(),
		Attributes: currentEntry.params,
	})

	if meErr.GetError() != nil {
		omciLogger.Errorf("Error while generating %s: %v", currentEntry.classId.String(), meErr.Error())
	}

	if reportedMe.GetClassID() == me.OnuDataClassID {
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
