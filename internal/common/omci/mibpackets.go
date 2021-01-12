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

// NOTE this is basically the same as https://github.com/opencord/voltha-openonu-adapter-go/blob/master/internal/pkg/onuadaptercore/omci_cc.go#L545-L564
// we should probably move it in "omci-lib-go"
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

func CreateMibResetRequest(tid uint16) ([]byte, error) {

	request := &omci.MibResetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
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

func CreateMibResetResponse(tid uint16) ([]byte, error) {

	// TODO reset MDX
	request := &omci.MibResetResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		Result: me.Success,
	}
	pkt, err := serialize(omci.MibResetResponseType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot serialize MibResetResponse")
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
	pkt, err := serialize(omci.MibUploadRequestType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibUploadRequest")
		return nil, err
	}
	return hexEncode(pkt)
}

func CreateMibUploadResponse(tid uint16) ([]byte, error) {

	numberOfCommands := uint16(291) //NOTE should this be configurable? (not until we have moved all the messages away from omci-sim)

	request := &omci.MibUploadResponse{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
		NumberOfCommands: numberOfCommands,
	}
	pkt, err := serialize(omci.MibUploadResponseType, request, tid)
	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Error("Cannot serialize MibUploadResponse")
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
	pkt, err := serialize(omci.MibUploadNextRequestType, request, tid)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibUploadNextRequest")
		return nil, err
	}
	return hexEncode(pkt)
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

func CreateMibUploadNextResponse(omciPkt gopacket.Packet, omciMsg *omci.OMCI) ([]byte, error) {

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

	// depending on the sequenceNumber we'll report a different
	reportedMe := &me.ManagedEntity{}
	var meErr me.OmciErrors
	//var entityInstance uint16
	switch msgObj.CommandSequenceNumber {
	case 0:
		reportedMe, meErr = me.NewOnuData(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId": me.OnuDataClassID,
			"MibDataSync":     0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewOnuData %v", meErr.Error())
		}

	case 1:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId": me.CircuitPackClassID,
			"Type":            47,
			"NumberOfPorts":   4,
			"SerialNumber":    toOctets("BBSM-Circuit-Pack", 20),
			"Version":         toOctets("v0.0.1", 20),
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 2:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":     me.CircuitPackClassID,
			"VendorId":            "ONF",
			"AdministrativeState": 0,
			"OperationalState":    0,
			"BridgedOrIpInd":      0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 3:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":             me.CircuitPackClassID,
			"EquipmentId":                 toOctets("BBSM-Circuit-Pack", 20),
			"CardConfiguration":           0,
			"TotalTContBufferNumber":      0,
			"TotalPriorityQueueNumber":    8,
			"TotalTrafficSchedulerNumber": 0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 4:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":   me.CircuitPackClassID,
			"PowerShedOverride": uint32(0),
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 5:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId": me.CircuitPackClassID,
			"Type":            238,
			"NumberOfPorts":   1,
			"SerialNumber":    toOctets("BBSM-Circuit-Pack-2", 20),
			"Version":         toOctets("v0.0.1", 20),
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 6:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":     me.CircuitPackClassID,
			"VendorId":            "ONF",
			"AdministrativeState": 0,
			"OperationalState":    0,
			"BridgedOrIpInd":      0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 7:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":             me.CircuitPackClassID,
			"EquipmentId":                 toOctets("BBSM-Circuit-Pack", 20),
			"CardConfiguration":           0,
			"TotalTContBufferNumber":      8,
			"TotalPriorityQueueNumber":    40,
			"TotalTrafficSchedulerNumber": 10,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 8:
		reportedMe, meErr = me.NewCircuitPack(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":   me.CircuitPackClassID,
			"PowerShedOverride": uint32(0),
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewCircuitPack %v", meErr.Error())
		}
	case 9, 10, 11, 12:
		// NOTE we're reporting for different UNIs, the IDs are 257, 258, 259, 260
		meInstance := 248 + msgObj.CommandSequenceNumber
		reportedMe, meErr = me.NewPhysicalPathTerminationPointEthernetUni(me.ParamData{
			EntityID: meInstance,
			Attributes: me.AttributeValueMap{
				"ExpectedType":                  0,
				"SensedType":                    47,
				"AutoDetectionConfiguration":    0,
				"EthernetLoopbackConfiguration": 0,
				"AdministrativeState":           0,
				"OperationalState":              0,
				"ConfigurationInd":              3,
				"MaxFrameSize":                  1518,
				"DteOrDceInd":                   0,
				"PauseTime":                     0,
				"BridgedOrIpInd":                2,
				"Arc":                           0,
				"ArcInterval":                   0,
				"PppoeFilter":                   0,
				"PowerControl":                  0,
			}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewPhysicalPathTerminationPointEthernetUni %v", meErr.Error())
		}
	case 13, 14, 15, 16, 17, 18, 19, 20:
		// TODO report different MeID (see omci-sim pcap filter "frame[22:2] == 01:06")
		reportedMe, meErr = me.NewTCont(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId": me.TContClassID,
			"AllocId":         0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewTCont %v", meErr.Error())
		}
	case 21:
		reportedMe, meErr = me.NewAniG(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":             me.AniGClassID,
			"SrIndication":                1,
			"TotalTcontNumber":            8,
			"GemBlockLength":              30,
			"PiggybackDbaReporting":       0,
			"Deprecated":                  0,
			"SignalFailThreshold":         5,
			"SignalDegradeThreshold":      9,
			"Arc":                         0,
			"ArcInterval":                 0,
			"OpticalSignalLevel":          57428,
			"LowerOpticalThreshold":       255,
			"UpperOpticalThreshold":       255,
			"OnuResponseTime":             0,
			"TransmitOpticalLevel":        3171,
			"LowerTransmitPowerThreshold": 129,
			"UpperTransmitPowerThreshold": 129,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewAniG %v", meErr.Error())
		}
	case 22, 23, 24, 25:
		// TODO report different MeID (see omci-sim pcap filter "frame[22:2] == 01:08")
		reportedMe, meErr = me.NewUniG(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":     me.UniGClassID,
			"Deprecated":          0,
			"AdministrativeState": 0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewUniG %v", meErr.Error())
		}
	case 26, 30, 34, 38, 42, 46, 50, 54,
		58, 62, 66, 70, 74, 78, 82, 86,
		90, 94, 98, 102, 106, 110, 114, 118,
		122, 126, 130, 134, 138, 142, 146, 150,
		154, 158, 162, 166, 170, 174, 178, 182,
		186, 190, 194, 198, 202, 206, 210, 214,
		218, 222, 226, 230, 234, 238, 242, 246,
		250, 254, 258, 262, 266, 270, 274, 278:
	case 27, 31, 35, 39, 43, 47, 51, 55,
		59, 63, 67, 71, 75, 79, 83, 87,
		91, 95, 99, 103, 107, 111, 115, 119,
		123, 127, 131, 135, 139, 143, 147, 151,
		155, 159, 163, 167, 171, 175, 179, 183,
		187, 191, 195, 199, 203, 207, 211, 215,
		219, 223, 227, 231, 235, 239, 243, 247,
		251, 255, 259, 263, 267, 271, 275, 279:
	case 28, 32, 36, 40, 44, 48, 52, 56,
		60, 64, 68, 72, 76, 80, 84, 88,
		92, 96, 100, 104, 108, 112, 116, 120,
		124, 128, 132, 136, 140, 144, 148, 152,
		156, 160, 164, 168, 172, 176, 180, 184,
		188, 192, 196, 200, 204, 208, 212, 216,
		220, 224, 228, 232, 236, 240, 244, 248,
		252, 256, 260, 264, 268, 272, 276, 280:
	case 29, 33, 37, 41, 45, 49, 53, 57,
		61, 65, 69, 73, 77, 81, 85, 89,
		93, 97, 101, 105, 109, 113, 117, 121,
		125, 129, 133, 137, 141, 145, 149, 153,
		157, 161, 165, 169, 173, 177, 181, 185,
		189, 193, 197, 201, 205, 209, 213, 217,
		221, 225, 229, 233, 237, 241, 245, 249,
		253, 257, 261, 265, 269, 273, 277, 281:
		reportedMe, meErr = me.NewPriorityQueue(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":                                     me.PriorityQueueClassID,
			"QueueConfigurationOption":                            0,
			"MaximumQueueSize":                                    100,
			"AllocatedQueueSize":                                  100,
			"DiscardBlockCounterResetInterval":                    0,
			"ThresholdValueForDiscardedBlocksDueToBufferOverflow": 0,
			"RelatedPort":                                         80010000, // does this need to change?
			"TrafficSchedulerPointer":                             8008,     // does this need to change?
			"Weight":                                              1,
			"BackPressureOperation":                               1,
			"BackPressureTime":                                    0,
			"BackPressureOccurQueueThreshold":                     0,
			"BackPressureClearQueueThreshold":                     0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewPriorityQueue %v", meErr.Error())
		}
	case 282, 283, 284, 285, 286, 287, 288, 289:
		reportedMe, meErr = me.NewTrafficScheduler(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":         me.TrafficSchedulerClassID,
			"TContPointer":            8008, // NOTE does this need to change?
			"TrafficSchedulerPointer": 0,
			"Policy":                  02,
			"PriorityWeight":          0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewTrafficScheduler %v", meErr.Error())
		}
	case 290:
		reportedMe, meErr = me.NewOnu2G(me.ParamData{Attributes: me.AttributeValueMap{
			"ManagedEntityId":             me.Onu2GClassID,
			"TotalPriorityQueueNumber":    40,
			"SecurityMode":                1,
			"TotalTrafficSchedulerNumber": 8,
			"TotalGemPortIdNumber":        0,
			"Sysuptime":                   0,
		}})
		if meErr.GetError() != nil {
			omciLogger.Errorf("NewOnu2G %v", meErr.Error())
		}
	default:
		omciLogger.Warn("unsupported-CommandSequenceNumber-in-mib-upload-next", msgObj.CommandSequenceNumber)
		return nil, nil
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

	pkt, err := serialize(omci.MibUploadNextResponseType, response, omciMsg.TransactionID)

	if err != nil {
		omciLogger.WithFields(log.Fields{
			"Err": err,
		}).Fatalf("Cannot serialize MibUploadNextRequest")
		return nil, err
	}

	return pkt, nil
}
