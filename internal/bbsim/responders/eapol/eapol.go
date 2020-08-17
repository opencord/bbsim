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
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	log "github.com/sirupsen/logrus"
)

var eapolLogger = log.WithFields(log.Fields{
	"module": "EAPOL",
})

var eapolVersion uint8 = 1
var GetGemPortId = omci.GetGemPortId

func sendEapolPktIn(msg bbsim.ByteMsg, portNo uint32, gemid uint32, stream bbsim.Stream) {
	// FIXME unify sendDHCPPktIn and sendEapolPktIn methods

	log.WithFields(log.Fields{
		"OnuId":   msg.OnuId,
		"IntfId":  msg.IntfId,
		"GemPort": gemid,
		"Type":    "EAPOL",
	}).Trace("sending-pkt")

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

func getMD5Data(eap *layers.EAP) []byte {
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

func createEAPChallengeRequest(eapId uint8, payload []byte) *layers.EAP {
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

func createEAPIdentityRequest(eapId uint8) *layers.EAP {
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

func createEAPSuccess(eapId uint8) *layers.EAP {
	eap := layers.EAP{
		Code:     layers.EAPCodeSuccess,
		Id:       eapId,
		Length:   9,
		Type:     layers.EAPTypeNone,
		TypeData: []byte{0x75, 0x73, 0x65, 0x72}}
	return &eap
}

func createEAPOLPkt(eap *layers.EAP, onuId uint32, intfId uint32) []byte {
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x2e, 0x60, byte(0), byte(intfId), byte(onuId), byte(0)},
		DstMAC:       net.HardwareAddr{0x2e, 0x60, byte(0), byte(intfId), byte(onuId), byte(0)},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	_ = gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: eapolVersion, Type: 0, Length: eap.Length},
		eap,
	)

	bytes := buffer.Bytes()
	return bytes
}

func extractEAP(pkt gopacket.Packet) (*layers.EAP, error) {
	layerEAP := pkt.Layer(layers.LayerTypeEAP)
	eap, _ := layerEAP.(*layers.EAP)
	if eap == nil {
		return nil, errors.New("Cannot extract EAP")
	}
	return eap, nil
}

func extractEAPOL(pkt gopacket.Packet) (*layers.EAPOL, error) {
	layerEAPOL := pkt.Layer(layers.LayerTypeEAPOL)
	eap, _ := layerEAPOL.(*layers.EAPOL)
	if eap == nil {
		return nil, errors.New("Cannot extract EAPOL")
	}
	return eap, nil
}

func sendEapolPktOut(client openolt.OpenoltClient, intfId uint32, onuId uint32, pkt []byte) error {
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

func SendEapStart(onuId uint32, ponPortId uint32, serialNumber string, portNo uint32, macAddress net.HardwareAddr, gemPort uint32, stateMachine *fsm.FSM, stream bbsim.Stream) error {

	// TODO use createEAPOLPkt
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       macAddress,
		DstMAC:       net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	_ = gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: eapolVersion, Type: 1, Length: 0},
	)

	msg := buffer.Bytes()
	// TODO end createEAPOLPkt

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType:  "pon",
			IntfId:    ponPortId,
			GemportId: gemPort,
			Pkt:       msg,
			PortNo:    portNo,
		},
	}

	err := stream.Send(&openolt.Indication{Data: data})
	if err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Can't send EapStart Message: %s", err)

		if err := updateAuthFailed(onuId, ponPortId, serialNumber, stateMachine); err != nil {
			return err
		}
		return err
	}

	eapolLogger.WithFields(log.Fields{
		"OnuId":  onuId,
		"IntfId": ponPortId,
		"OnuSn":  serialNumber,
		"PortNo": portNo,
	}).Debugf("Sent EapStart packet")

	if err := stateMachine.Event("eap_start_sent"); err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Errorf("Error while transitioning ONU State %v", err)
		return err
	}
	return nil
}

func HandleNextPacket(onuId uint32, ponPortId uint32, gemPortId uint32, serialNumber string, portNo uint32, stateMachine *fsm.FSM, pkt gopacket.Packet, stream bbsim.Stream, client openolt.OpenoltClient) {

	eap, eapErr := extractEAP(pkt)

	eapol, eapolErr := extractEAPOL(pkt)

	if eapErr != nil && eapolErr != nil {
		log.Fatalf("Failed to Extract EAP: %v - %v", eapErr, eapolErr)
		return
	}

	fields := log.Fields{}
	if eap != nil {
		fields = log.Fields{
			"Code":   eap.Code,
			"Type":   eap.Type,
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}
	} else if eapol != nil {
		fields = log.Fields{
			"Type":   eapol.Type,
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}
	}

	log.WithFields(fields).Tracef("Handle Next EAPOL Packet")

	if eapol != nil && eapol.Type == layers.EAPOLTypeStart {
		identityRequest := createEAPIdentityRequest(1)
		pkt := createEAPOLPkt(identityRequest, onuId, ponPortId)

		if err := sendEapolPktOut(client, ponPortId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"error":  err,
			}).Errorf("Error while sending EAPIdentityRequest packet")
			return
		}

		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Sent EAPIdentityRequest packet")
		return
	} else if eap.Code == layers.EAPCodeRequest && eap.Type == layers.EAPTypeIdentity {
		reseap := createEAPIdentityResponse(eap.Id)
		pkt := createEAPOLPkt(reseap, onuId, ponPortId)

		msg := bbsim.ByteMsg{
			IntfId: ponPortId,
			OnuId:  onuId,
			Bytes:  pkt,
		}

		sendEapolPktIn(msg, portNo, gemPortId, stream)
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"PortNo": portNo,
		}).Debugf("Sent EAPIdentityResponse packet")
		if err := stateMachine.Event("eap_response_identity_sent"); err != nil {
			eapolLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Errorf("Error while transitioning ONU State %v", err)
		}

	} else if eap.Code == layers.EAPCodeResponse && eap.Type == layers.EAPTypeIdentity {
		senddata := getMD5Data(eap)
		senddata = append([]byte{0x10}, senddata...)
		challengeRequest := createEAPChallengeRequest(eap.Id, senddata)
		pkt := createEAPOLPkt(challengeRequest, onuId, ponPortId)

		if err := sendEapolPktOut(client, ponPortId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"error":  err,
			}).Errorf("Error while sending EAPChallengeRequest packet")
			return
		}
		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Sent EAPChallengeRequest packet")
		return
	} else if eap.Code == layers.EAPCodeRequest && eap.Type == layers.EAPTypeOTP {
		senddata := getMD5Data(eap)
		senddata = append([]byte{0x10}, senddata...)
		sendeap := createEAPChallengeResponse(eap.Id, senddata)
		pkt := createEAPOLPkt(sendeap, onuId, ponPortId)

		msg := bbsim.ByteMsg{
			IntfId: ponPortId,
			OnuId:  onuId,
			Bytes:  pkt,
		}

		sendEapolPktIn(msg, portNo, gemPortId, stream)
		eapolLogger.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
			"PortNo": portNo,
		}).Debugf("Sent EAPChallengeResponse packet")
		if err := stateMachine.Event("eap_response_challenge_sent"); err != nil {
			eapolLogger.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
			}).Errorf("Error while transitioning ONU State %v", err)
		}
	} else if eap.Code == layers.EAPCodeResponse && eap.Type == layers.EAPTypeOTP {
		eapSuccess := createEAPSuccess(eap.Id)
		pkt := createEAPOLPkt(eapSuccess, onuId, ponPortId)

		if err := sendEapolPktOut(client, ponPortId, onuId, pkt); err != nil {
			log.WithFields(log.Fields{
				"OnuId":  onuId,
				"IntfId": ponPortId,
				"OnuSn":  serialNumber,
				"error":  err,
			}).Errorf("Error while sending EAPSuccess packet")
			return
		}

		log.WithFields(log.Fields{
			"OnuId":  onuId,
			"IntfId": ponPortId,
			"OnuSn":  serialNumber,
		}).Infof("Sent EAP Success packet")

		if err := stateMachine.Event("send_dhcp_flow"); err != nil {
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
			"PortNo": portNo,
		}).Debugf("Received EAPSuccess packet")
		if err := stateMachine.Event("eap_response_success_received"); err != nil {
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
