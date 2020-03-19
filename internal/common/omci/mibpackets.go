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
	"github.com/cboling/omci"
	me "github.com/cboling/omci/generated"
	"github.com/google/gopacket"
	omcisim "github.com/opencord/omci-sim"
	log "github.com/sirupsen/logrus"
)

var omciLogger = log.WithFields(log.Fields{
	"module": "OMCI",
})

const galEthernetEID = uint16(1)
const maxGemPayloadSize = uint16(48)
const gemEID = uint16(1)

type txFrameCreator func() ([]byte, error)
type rxFrameParser func(gopacket.Packet) error

type ServiceStep struct {
	MakeTxFrame txFrameCreator
	RxHandler   rxFrameParser
}

func serialize(msgType omci.MessageType, request gopacket.SerializableLayer, tid uint16) ([]byte, error) {
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

func hexEncode(omciPkt []byte) ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(omciPkt)))
	hex.Encode(dst, omciPkt)
	return dst, nil
}

func DecodeOmci(payload []byte) (omci.MessageType, gopacket.Packet) {
	// Perform base OMCI decode (common to all requests)
	packet := gopacket.NewPacket(payload, omci.LayerTypeOMCI, gopacket.NoCopy)

	if omciLayer := packet.Layer(omci.LayerTypeOMCI); omciLayer != nil {

		omciObj, omciOk := omciLayer.(*omci.OMCI)
		if !omciOk {
			panic("Not Expected") // TODO: Do something better or delete...
		}
		if byte(omciObj.MessageType) & ^me.AK == 0 {
			// Not a response, silently discard
			return 0, nil
		}
		return omciObj.MessageType, packet
	}

	// FIXME
	// if we can't properly decode the packet, try using shad helper method
	// most likely this won't be necessary once we move omci-sim to use cboling/omci
	// to generate packets
	_, _, msgType, _, _, _, err := omcisim.ParsePkt(payload)
	if err != nil {
		return 0, nil
	}
	if msgType == omcisim.MibReset {
		return omci.MibResetResponseType, nil
	}
	if msgType == omcisim.MibUpload {
		return omci.MibUploadResponseType, nil
	}
	if msgType == omcisim.MibUploadNext {
		return omci.MibUploadNextResponseType, nil
	}
	if msgType == omcisim.Create {
		return omci.CreateResponseType, nil
	}
	if msgType == omcisim.Set {
		return omci.SetResponseType, nil
	}

	omciLogger.Warnf("omci-sim returns msgType: %d", msgType)

	return 0, nil
}

func CreateMibResetRequest(tid uint16) ([]byte, error) {

	request := &omci.MibResetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassId,
		},
	}
	pkt, err := serialize(omci.MibResetRequestType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibResetRequest")
		return nil, err
	}
	return hexEncode(pkt)
}

func CreateMibUploadRequest(tid uint16) ([]byte, error) {
	request := &omci.MibUploadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassId,
			// Default Instance ID is 0
		},
	}
	pkt, err := serialize(omci.MibUploadRequestType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibUploadRequest")
		return nil, err
	}
	return hexEncode(pkt)
}

func CreateMibUploadNextRequest(tid uint16, seqNumber uint16) ([]byte, error) {

	request := &omci.MibUploadNextRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassId,
			// Default Instance ID is 0
		},
		CommandSequenceNumber: seqNumber,
	}
	pkt, err := serialize(omci.MibUploadNextRequestType, request, tid)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibUploadNextRequest")
		return nil, err
	}
	return hexEncode(pkt)
}

// Return true if msg is an Omci Test Request
func IsTestRequest(payload []byte) (bool, error) {
	_, _, msgType, _, _, _, err := omcisim.ParsePkt(payload)
	if err != nil {
		return false, err
	}

	return ((msgType & 0x1F) == 18), nil
}

func BuildTestResult(payload []byte) ([]byte, error) {
	transactionId, deviceId, _, class, instance, _, err := omcisim.ParsePkt(payload)

	if err != nil {
		return []byte{}, err
	}

	resp := make([]byte, 48)
	resp[0] = byte(transactionId >> 8)
	resp[1] = byte(transactionId & 0xFF)
	resp[2] = 27 // Upper nibble 0x0 is fixed (0000), Lower nibbles defines msg type (TestResult=27)
	resp[3] = deviceId
	resp[4] = byte(class >> 8)
	resp[5] = byte(class & 0xFF)
	resp[6] = byte(instance >> 8)
	resp[7] = byte(instance & 0xFF)
	// Each of these is a 1-byte code
	// follow by a 2-byte (high, low) value
	resp[8] = 1 // power feed voltage
	resp[9] = 0
	resp[10] = 123 // 123 mV, 20 mv res --> 6mv
	resp[11] = 3   // received optical power
	resp[12] = 1
	resp[13] = 200 // 456 decibel-microwatts, 0.002 dB res --> 0.912 db-mw
	resp[14] = 5   // mean optical launch power
	resp[15] = 3
	resp[16] = 21 // 789 uA, 0.002 dB res --> 1.578 db-mw
	resp[17] = 9  // laser bias current
	resp[18] = 3
	resp[19] = 244 // 1012 uA, 2uA res --> 505 ua
	resp[20] = 12  // temperature
	resp[21] = 38
	resp[22] = 148 // 9876 deg C, 1/256 resolution --> 38.57 Deg C

	return resp, nil
}

// TODO understand and refactor

func CreateGalEnetRequest(tid uint16) ([]byte, error) {
	params := me.ParamData{
		EntityID:   galEthernetEID,
		Attributes: me.AttributeValueMap{"MaximumGemPayloadSize": maxGemPayloadSize},
	}
	meDef, _ := me.NewGalEthernetProfile(params)
	pkt, err := omci.GenFrame(meDef, omci.CreateRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate GalEnetRequest")
	}
	return hexEncode(pkt)
}

func CreateEnableUniRequest(tid uint16, uniId uint16, enabled bool, isPtp bool) ([]byte, error) {

	var _enabled uint8
	if enabled {
		_enabled = uint8(1)
	} else {
		_enabled = uint8(0)
	}

	data := me.ParamData{
		EntityID: uniId,
		Attributes: me.AttributeValueMap{
			"AdministrativeState": _enabled,
		},
	}
	var medef *me.ManagedEntity
	var omciErr me.OmciErrors

	if isPtp {
		medef, omciErr = me.NewPhysicalPathTerminationPointEthernetUni(data)
	} else {
		medef, omciErr = me.NewVirtualEthernetInterfacePoint(data)
	}
	if omciErr != nil {
		return nil, omciErr.GetError()
	}
	pkt, err := omci.GenFrame(medef, omci.SetRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate EnableUniRequest")
	}
	return hexEncode(pkt)
}

func CreateGemPortRequest(tid uint16) ([]byte, error) {
	params := me.ParamData{
		EntityID: gemEID,
		Attributes: me.AttributeValueMap{
			"PortId":                              1,
			"TContPointer":                        1,
			"Direction":                           0,
			"TrafficManagementPointerForUpstream": 0,
			"TrafficDescriptorProfilePointerForUpstream": 0,
			"UniCounter":                                   0,
			"PriorityQueuePointerForDownStream":            0,
			"EncryptionState":                              0,
			"TrafficDescriptorProfilePointerForDownstream": 0,
			"EncryptionKeyRing":                            0,
		},
	}
	meDef, _ := me.NewGemPortNetworkCtp(params)
	pkt, err := omci.GenFrame(meDef, omci.CreateRequestType, omci.TransactionID(tid))
	if err != nil {
		omciLogger.WithField("err", err).Fatalf("Can't generate GemPortRequest")
	}
	return hexEncode(pkt)
}

// END TODO
