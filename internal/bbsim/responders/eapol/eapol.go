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

package eapol

import (
	"context"
	"crypto/md5"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
)

var eapolLogger = log.WithFields(log.Fields{
	"module": "EAPOL",
})

var eapolVersion uint8 = 1
var GetGemPortId = omci.GetGemPortId

// FIXME once we handle multiple UNIs. start EAPOL on all of them
func sendEapolPktIn(msg bbsim.ByteMsg, portNo uint32, stream openolt.Openolt_EnableIndicationServer) {
	// FIXME unify sendDHCPPktIn and sendEapolPktIn methods
	gemid, err := omci.GetGemPortId(msg.IntfId, msg.OnuId)
	if err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  msg.OnuId,
			"IntfId": msg.IntfId,
		}).Errorf("Can't retrieve GemPortId: %s", err)
		return
	}
	data := &openolt.Indication_PktInd{PktInd: &openolt.PacketIndication{
		IntfType:  "pon",
		IntfId:    msg.IntfId,
		GemportId: uint32(gemid),
		Pkt:       msg.Bytes,
		PortNo:    portNo,
	}}

	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		eapolLogger.Errorf("Fail to send EAPOL PktInd indication. %v", err)
		return
	}
}

func SendEapolPktOut(client openolt.OpenoltClient, intfId uint32, onuId uint32, pkt []byte) error {
	onuPacket := openolt.OnuPacket{
		IntfId:    intfId,
		OnuId:     onuId,
		PortNo:    onuId,
		GemportId: 1,
		Pkt:       pkt,
	}

	if _, err := client.OnuPacketOut(context.Background(), &onuPacket); err != nil {
		return err
	}
	return nil
}

func GetMD5Data(eap *layers.EAP) []byte {
	i := byte(eap.Id)
	C := []byte(eap.BaseLayer.Contents)[6:]
	P := []byte{i, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64} // "password"
	data := md5.Sum(append(P, C...))
	ret := make([]byte, 16)
	for j := 0; j < 16; j++ {
		ret[j] = data[j]
	}
	return ret
}

func CreateEAPChallengeRequest(eapId uint8, payload []byte) *layers.EAP {
	eap := layers.EAP{
		Code:     layers.EAPCodeRequest,
		Id:       eapId,
		Length:   22,
		Type:     layers.EAPTypeOTP,
		TypeData: payload,
	}
	return &eap
}

func createEAPChallengeResponse(eapId uint8, payload []byte) *layers.EAP {
	eap := layers.EAP{
		Code:     layers.EAPCodeResponse,
		Id:       eapId,
		Length:   22,
		Type:     layers.EAPTypeOTP,
		TypeData: payload,
	}
	return &eap
}

func CreateEAPIdentityRequest(eapId uint8) *layers.EAP {
	eap := layers.EAP{Code: layers.EAPCodeRequest,
		Id:       eapId,
		Length:   9,
		Type:     layers.EAPTypeIdentity,
		TypeData: []byte{0x75, 0x73, 0x65, 0x72}}
	return &eap
}

func createEAPIdentityResponse(eapId uint8) *layers.EAP {
	eap := layers.EAP{Code: layers.EAPCodeResponse,
		Id:       eapId,
		Length:   9,
		Type:     layers.EAPTypeIdentity,
		TypeData: []byte{0x75, 0x73, 0x65, 0x72}}
	return &eap
}

func CreateEAPSuccess(eapId uint8) *layers.EAP {
	eap := layers.EAP{
		Code:     layers.EAPCodeSuccess,
		Id:       eapId,
		Length:   9,
		Type:     layers.EAPTypeNone,
		TypeData: []byte{0x75, 0x73, 0x65, 0x72}}
	return &eap
}

func CreateEAPOLPkt(eap *layers.EAP, onuId uint32, intfId uint32) []byte {
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(intfId), byte(onuId)},
		DstMAC:       net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: eapolVersion, Type: 0, Length: eap.Length},
		eap,
	)

	bytes := buffer.Bytes()
	return bytes
}

func ExtractEAP(pkt gopacket.Packet) (*layers.EAP, error) {
	layerEAP := pkt.Layer(layers.LayerTypeEAP)
	eap, _ := layerEAP.(*layers.EAP)
	if eap == nil {
		return nil, errors.New("Cannot extract EAP")
	}
	return eap, nil
}

func ExtractEAPOL(pkt gopacket.Packet) (*layers.EAPOL, error) {
	layerEAPOL := pkt.Layer(layers.LayerTypeEAPOL)
	eap, _ := layerEAPOL.(*layers.EAPOL)
	if eap == nil {
		return nil, errors.New("Cannot extract EAPOL")
	}
	return eap, nil
}

func updateAuthFailed(onuId uint32, ponPortId uint32, serialNumber string, onuStateMachine *fsm.FSM) error {
	if err := onuStateMachine.Event("auth_failed"); err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Error while transitioning ONU State %v", err)
		return err
	}
	return nil
}

func SendEapStart(onuId uint32, ponPortId uint32, portNo uint32, serialNumber string, onuStateMachine *fsm.FSM, stream openolt.Openolt_EnableIndicationServer) error {

	// send the packet (hacked together)
	gemId, err := GetGemPortId(ponPortId, onuId)
	if err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Can't retrieve GemPortId: %s", err)

		if err := updateAuthFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}

	// TODO use CreateEAPOLPkt
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(ponPortId), byte(onuId)}, // TODO move the SrcMAC in the ONU Device
		DstMAC:       net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: eapolVersion, Type: 1, Length: 0},
	)

	msg := buffer.Bytes()
	// TODO end CreateEAPOLPkt

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: uint32(gemId),
			Pkt:       msg,
			PortNo:    portNo,
		},
	}

	err = stream.Send(&openolt.Indication{Data: data})
	if err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Can't send EapStart Message: %s", err)

		if err := updateAuthFailed(onuId, ponPortId, serialNumber, onuStateMachine); err != nil {
			return err
		}
		return err
	}

	eapolLogger.WithFields(log.Fields{
		"OnuId":  onuId,
		"IntfId": ponPortId,
		"OnuSn":  serialNumber,
		"PortNo": portNo,
	}).Debugf("EapStart packet indication sent")

	if err := onuStateMachine.Event("eap_start_sent"); err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Error while transitioning ONU State %v", err)
		return err
	}
	return nil
}

func HandleNextPacket(onuId uint32, ponPortId uint32, portNo uint32, serialNumber string, onuStateMachine *fsm.FSM, recvpkt gopacket.Packet, stream openolt.Openolt_EnableIndicationServer) {
	eap, err := ExtractEAP(recvpkt)
	if err != nil {
		eapolLogger.Errorf("%s", err)
	}

	log.WithFields(log.Fields{
		"eap.Code": eap.Code,
		"eap.Type": eap.Type,
	}).Warn("HandleNextPacket")

	if eap.Code == layers.EAPCodeRequest && eap.Type == layers.EAPTypeIdentity {
		reseap := createEAPIdentityResponse(eap.Id)
		pkt := CreateEAPOLPkt(reseap, onuId, ponPortId)

		msg := bbsim.ByteMsg{
			IntfId: ponPortId,
			OnuId:  onuId,
			Bytes:  pkt,
		}

		sendEapolPktIn(msg, portNo, stream)
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Sent EAPIdentityResponse packet")
		if err := onuStateMachine.Event("eap_response_identity_sent"); err != nil {
			eapolLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Errorf("Error while transitioning ONU State %v", err)
		}

	} else if eap.Code == layers.EAPCodeRequest && eap.Type == layers.EAPTypeOTP {
		senddata := GetMD5Data(eap)
		senddata = append([]byte{0x10}, senddata...)
		sendeap := createEAPChallengeResponse(eap.Id, senddata)
		pkt := CreateEAPOLPkt(sendeap, onuId, ponPortId)

		msg := bbsim.ByteMsg{
			IntfId: ponPortId,
			OnuId:  onuId,
			Bytes:  pkt,
		}

		sendEapolPktIn(msg, portNo, stream)
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Sent EAPChallengeResponse packet")
		if err := onuStateMachine.Event("eap_response_challenge_sent"); err != nil {
			eapolLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Errorf("Error while transitioning ONU State %v", err)
		}
	} else if eap.Code == layers.EAPCodeSuccess && eap.Type == layers.EAPTypeNone {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Received EAPSuccess packet")
		if err := onuStateMachine.Event("eap_response_success_received"); err != nil {
			eapolLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Errorf("Error while transitioning ONU State %v", err)
		}
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("EAPOL State machine completed")
		return
	}
}
