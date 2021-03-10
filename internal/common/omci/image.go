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
	"encoding/hex"
	"errors"
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ParseStartSoftwareDownloadRequest(omciPkt gopacket.Packet) (*omci.StartSoftwareDownloadRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeStartSoftwareDownloadRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeStartSoftwareDownloadRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.StartSoftwareDownloadRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeStartSoftwareDownloadRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func ParseDownloadSectionRequest(omciPkt gopacket.Packet) (*omci.DownloadSectionRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeDownloadSectionRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeDownloadSectionRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.DownloadSectionRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeDownloadSectionRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func ParseEndSoftwareDownloadRequest(omciPkt gopacket.Packet) (*omci.EndSoftwareDownloadRequest, error) {
	msgLayer := omciPkt.Layer(omci.LayerTypeEndSoftwareDownloadRequest)
	if msgLayer == nil {
		err := "omci Msg layer could not be detected for LayerTypeEndSoftwareDownloadRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	msgObj, msgOk := msgLayer.(*omci.EndSoftwareDownloadRequest)
	if !msgOk {
		err := "omci Msg layer could not be assigned for LayerTypeEndSoftwareDownloadRequest"
		omciLogger.Error(err)
		return nil, errors.New(err)
	}
	return msgObj, nil
}

func CreateStartSoftwareDownloadResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {
	responeCode := me.Success
	msgObj, err := ParseStartSoftwareDownloadRequest(omciPkt)
	if err != nil {
		responeCode = me.ProcessingError
	}

	omciLogger.WithFields(log.Fields{
		"OmciMsgType":          omciMsg.MessageType,
		"TransCorrId":          omciMsg.TransactionID,
		"EntityInstance":       msgObj.EntityInstance,
		"WindowSize":           msgObj.WindowSize,
		"ImageSize":            msgObj.ImageSize,
		"NumberOfCircuitPacks": msgObj.NumberOfCircuitPacks,
		"CircuitPacks":         msgObj.CircuitPacks,
	}).Info("received-start-software-download-request")

	response := &omci.StartSoftwareDownloadResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		WindowSize:        msgObj.WindowSize,
		Result:            responeCode,
		NumberOfInstances: 0,                        // NOTE this can't be bigger than 0 this we can populate downloadResults
		MeResults:         []omci.DownloadResults{}, // FIXME downloadResults is not exported
	}

	pkt, err := Serialize(omci.StartSoftwareDownloadResponseType, response, omciMsg.TransactionID)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("cannot-Serialize-CreateResponse")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Debug("omci-start-software-download-response")

	return pkt, nil
}

func CreateDownloadSectionResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseDownloadSectionRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"OmciMsgType":    omciMsg.MessageType,
		"TransCorrId":    omciMsg.TransactionID,
		"EntityInstance": msgObj.EntityInstance,
		"SectionNumber":  msgObj.SectionNumber,
		"SectionData":    msgObj.SectionData,
	}).Info("received-download-section-request")

	response := &omci.DownloadSectionResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		Result:        me.Success,
		SectionNumber: msgObj.SectionNumber,
	}

	pkt, err := Serialize(omci.DownloadSectionResponseType, response, omciMsg.TransactionID)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Debug("omci-download-section-with-response-response")

	return pkt, nil
}

func CreateEndSoftwareDownloadResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

	msgObj, err := ParseEndSoftwareDownloadRequest(omciPkt)

	if err != nil {
		return nil, err
	}

	omciLogger.WithFields(log.Fields{
		"OmciMsgType":       omciMsg.MessageType,
		"TransCorrId":       omciMsg.TransactionID,
		"EntityInstance":    msgObj.EntityInstance,
		"Crc32":             msgObj.CRC32,
		"ImageSize":         msgObj.ImageSize,
		"NumberOfInstances": msgObj.NumberOfInstances,
		"ImageInstances":    msgObj.ImageInstances,
	}).Info("received-end-software-download-request")

	response := &omci.EndSoftwareDownloadResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    msgObj.EntityClass,
			EntityInstance: msgObj.EntityInstance,
		},
		Result:            me.Success,
		NumberOfInstances: 0, // NOTE this can't be bigger than 0 this we can populate downloadResults
		//MeResults: []omci.downloadResults{}, // FIXME downloadResults is not exported
	}

	pkt, err := Serialize(omci.EndSoftwareDownloadResponseType, response, omciMsg.TransactionID)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"TxID": strconv.FormatInt(int64(omciMsg.TransactionID), 16),
		"pkt":  hex.EncodeToString(pkt),
	}).Debug("omci-end-software-download-response")

	return pkt, nil
}
