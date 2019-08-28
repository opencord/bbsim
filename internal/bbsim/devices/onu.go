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

package devices

import (
	"fmt"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

func CreateONU(olt OltDevice, pon PonPort, id uint32) Onu {
		o := Onu{
			ID: id,
			PonPortID: pon.ID,
			PonPort: pon,
			// NOTE can we combine everything in a single channel?
			channel: make(chan Message),
			eapolPktOutCh: make(chan *bbsim.ByteMsg, 1024),
		}
		o.SerialNumber = o.NewSN(olt.ID, pon.ID, o.ID)

		// NOTE this state machine is used to track the operational
		// state as requested by VOLTHA
		o.OperState = getOperStateFSM(func(e *fsm.Event) {
			onuLogger.WithFields(log.Fields{
				"ID": o.ID,
			}).Debugf("Changing ONU OperState from %s to %s", e.Src, e.Dst)
		})

		// NOTE this state machine is used to activate the OMCI, EAPOL and DHCP clients
		o.InternalState = fsm.NewFSM(
			"created",
			fsm.Events{
				{Name: "discover", Src: []string{"created"}, Dst: "discovered"},
				{Name: "enable", Src: []string{"discovered"}, Dst: "enabled"},
				{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
				{Name: "add_gem_port", Src: []string{"enabled", "eapol_flow_received"}, Dst: "gem_port_added"},
				{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
				{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
				{Name: "eap_resonse_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_resonse_identity_sent"},
				{Name: "eap_resonse_challenge_sent", Src: []string{"eap_resonse_identity_sent"}, Dst: "eap_resonse_challenge_sent"},
				{Name: "eap_resonse_success_received", Src: []string{"eap_resonse_challenge_sent"}, Dst: "eap_resonse_success_received"},
				{Name: "auth_failed", Src: []string{"auth_started", "eap_start_sent", "eap_resonse_identity_sent", "eap_resonse_challenge_sent"}, Dst: "auth_failed"},
			},
			fsm.Callbacks{
				"enter_state": func(e *fsm.Event) {
					o.logStateChange(e.Src, e.Dst)
				},
				"enter_eapol_flow_received": func(e *fsm.Event) {
					o.logStateChange(e.Src, e.Dst)
					if e.Src == "enter_gem_port_added" {
						if err := o.InternalState.Event("start_auth"); err != nil {
							log.Infof("Transitioning to StartAuth")
							onuLogger.WithFields(log.Fields{
								"OnuId":  o.ID,
								"IntfId": o.PonPortID,
								"OnuSn":  o.SerialNumber,
							}).Errorf("Error while transitioning ONU State")
						}
					}
				},
				"enter_gem_port_added": func(e *fsm.Event) {
					o.logStateChange(e.Src, e.Dst)
					if e.Src == "eapol_flow_received" {
						log.Infof("Transitioning to StartAuth")
						if err := o.InternalState.Event("start_auth"); err != nil {
							onuLogger.WithFields(log.Fields{
								"OnuId": o.ID,
								"IntfId": o.PonPortID,
								"OnuSn": o.SerialNumber,
							}).Errorf("Error while transitioning ONU State")
						}
					}
				},
				"enter_auth_started": func(e *fsm.Event) {
					o.logStateChange(e.Src, e.Dst)
					msg := Message{
						Type:      StartEAPOL,
						Data: EapStartMessage{
							PonPortID: o.PonPortID,
							OnuID: o.ID,
						},
					}
					go func(msg Message){
						// you can only send a value on an unbuffered channel without blocking
						o.channel <- msg
					}(msg)

				},
				"enter_eap_resonse_success_received": func(e *fsm.Event) {
					o.logStateChange(e.Src, e.Dst)
					onuLogger.WithFields(log.Fields{
						"OnuId": o.ID,
						"IntfId": o.PonPortID,
						"OnuSn": o.SerialNumber,
					}).Warnf("TODO start DHCP request")
				},
			},
		)
		return o
}

func (o Onu) logStateChange(src string, dst string) {
	onuLogger.WithFields(log.Fields{
		"OnuId": o.ID,
		"IntfId": o.PonPortID,
		"OnuSn": o.SerialNumber,
	}).Debugf("Changing ONU InternalState from %s to %s", src, dst)
}

func (o Onu) processOnuMessages(stream openolt.Openolt_EnableIndicationServer)  {
	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.SerialNumber,
	}).Debug("Started ONU Indication Channel")

	for message := range o.channel {
		onuLogger.WithFields(log.Fields{
			"onuID": o.ID,
			"onuSN": o.SerialNumber,
			"messageType": message.Type,
		}).Tracef("Received message on ONU Channel")

		switch message.Type {
		case OnuDiscIndication:
			msg, _ := message.Data.(OnuDiscIndicationMessage)
			o.sendOnuDiscIndication(msg, stream)
		case OnuIndication:
			msg, _ := message.Data.(OnuIndicationMessage)
			o.sendOnuIndication(msg, stream)
		case OMCI:
			msg, _ := message.Data.(OmciMessage)
			o.handleOmciMessage(msg, stream)
		case FlowUpdate:
			msg, _ := message.Data.(OnuFlowUpdateMessage)
			o.handleFlowUpdate(msg, stream)
		case StartEAPOL:
			log.Infof("Receive StartEAPOL message on ONU channel")
			go func() {
				// TODO kill this thread
				eapol.CreateWPASupplicant(o.ID, o.PonPortID, o.SerialNumber, o.InternalState, stream, o.eapolPktOutCh)
			}()
		default:
			onuLogger.Warnf("Received unknown message data %v for type %v in OLT channel", message.Data, message.Type)
		}
	}
}

func (o Onu) processOmciMessages(stream openolt.Openolt_EnableIndicationServer)  {
	ch := omci.GetChannel()

	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.SerialNumber,
	}).Debug("Started OMCI Indication Channel")

	for message := range ch {
		switch message.Type {
		case omci.GemPortAdded:
			log.WithFields(log.Fields{
				"OnuId": message.Data.OnuId,
				"IntfId": message.Data.IntfId,
			}).Infof("GemPort Added")

			// NOTE if we receive the GemPort but we don't have EAPOL flows
			// go an intermediate state, otherwise start auth
			if o.InternalState.Is("enabled") {
				if err := o.InternalState.Event("add_gem_port"); err != nil {
					log.Errorf("Can't go to gem_port_added: %v", err)
				}
			} else if  o.InternalState.Is("eapol_flow_received"){
				if err := o.InternalState.Event("start_auth"); err != nil {
					log.Errorf("Can't go to auth_started: %v", err)
				}
			}
		}
	}
}

func (o Onu) NewSN(oltid int, intfid uint32, onuid uint32) *openolt.SerialNumber {

	sn := new(openolt.SerialNumber)

	//sn = new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}

	return sn
}

func (o Onu) sendOnuDiscIndication(msg OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId: msg.Onu.PonPortID,
		SerialNumber: msg.Onu.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Errorf("Failed to send Indication_OnuDiscInd: %v", err)
	}
	o.InternalState.Event("discover")
	onuLogger.WithFields(log.Fields{
		"IntfId": msg.Onu.PonPortID,
		"SerialNumber": msg.Onu.SerialNumber,
		"OnuId": o.ID,
	}).Debug("Sent Indication_OnuDiscInd")
}

func (o Onu) sendOnuIndication(msg OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	// NOTE voltha returns an ID, but if we use that ID then it complains:
	// expected_onu_id: 1, received_onu_id: 1024, event: ONU-id-mismatch, can happen if both voltha and the olt rebooted
	// so we're using the internal ID that is 1
	// o.ID = msg.OnuID
	o.OperState.Event("enable")

	indData := &openolt.Indication_OnuInd{OnuInd: &openolt.OnuIndication{
		IntfId: o.PonPortID,
		OnuId: o.ID,
		OperState: o.OperState.Current(),
		AdminState: o.OperState.Current(),
		SerialNumber: o.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: indData}); err != nil {
		log.Errorf("Failed to send Indication_OnuInd: %v", err)
	}
	o.InternalState.Event("enable")
	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"OnuId": o.ID,
		"OperState": msg.OperState.String(),
		"AdminState": msg.OperState.String(),
		"SerialNumber": o.SerialNumber,
	}).Debug("Sent Indication_OnuInd")
}

func (o Onu) handleOmciMessage(msg OmciMessage, stream openolt.Openolt_EnableIndicationServer) {

	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket": msg.omciMsg.Pkt,
	}).Tracef("Received OMCI message")

	var omciInd openolt.OmciIndication
	respPkt, err := omci.OmciSim(o.PonPortID, o.ID, HexDecode(msg.omciMsg.Pkt))
	if err != nil {
		onuLogger.Errorf("Error handling OMCI message %v", msg)
	}

	omciInd.IntfId = o.PonPortID
	omciInd.OnuId = o.ID
	omciInd.Pkt = respPkt

	omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
	if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
		onuLogger.Errorf("send omci indication failed: %v", err)
	}
	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket": omciInd.Pkt,
	}).Tracef("Sent OMCI message")
}

func (o Onu) handleFlowUpdate(msg OnuFlowUpdateMessage, stream openolt.Openolt_EnableIndicationServer) {
	onuLogger.WithFields(log.Fields{
		"IntfId": msg.Flow.AccessIntfId,
		"OnuId": msg.Flow.OnuId,
		"EthType": fmt.Sprintf("%x", msg.Flow.Classifier.EthType),
		"InnerVlan": msg.Flow.Classifier.IVid,
		"OuterVlan": msg.Flow.Classifier.OVid,
		"FlowType": msg.Flow.FlowType,
		"FlowId": msg.Flow.FlowId,
		"UniID": msg.Flow.UniId,
		"PortNo": msg.Flow.PortNo,
	}).Infof("ONU receives Flow")
	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		// NOTE if we receive the EAPOL flows but we don't have GemPorts
		// go an intermediate state, otherwise start auth
		if o.InternalState.Is("enabled") {
			if err := o.InternalState.Event("receive_eapol_flow"); err != nil {
				log.Errorf("Can't go to eapol_flow_received: %v", err)
			}
		} else if  o.InternalState.Is("gem_port_added"){
			if err := o.InternalState.Event("start_auth"); err != nil {
				log.Errorf("Can't go to auth_started: %v", err)
			}
		}
	}
}

// HexDecode converts the hex encoding to binary
func HexDecode(pkt []byte) []byte {
	p := make([]byte, len(pkt)/2)
	for i, j := 0, 0; i < len(pkt); i, j = i+2, j+1 {
		// Go figure this ;)
		u := (pkt[i] & 15) + (pkt[i]>>6)*9
		l := (pkt[i+1] & 15) + (pkt[i+1]>>6)*9
		p[j] = u<<4 + l
	}
	onuLogger.Tracef("Omci decoded: %x.", p)
	return p
}