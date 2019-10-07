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
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

func CreateONU(olt OltDevice, pon PonPort, id uint32, sTag int, cTag int) Onu {

	o := Onu{
		ID:        id,
		PonPortID: pon.ID,
		PonPort:   pon,
		STag:      sTag,
		CTag:      cTag,
		HwAddress: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(pon.ID), byte(id)},
		// NOTE can we combine everything in a single Channel?
		Channel:       make(chan Message, 2048),
		eapolPktOutCh: make(chan *bbsim.ByteMsg, 1024),
		dhcpPktOutCh:  make(chan *bbsim.ByteMsg, 1024),
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
			// DEVICE Lifecycle
			{Name: "discover", Src: []string{"created"}, Dst: "discovered"},
			{Name: "enable", Src: []string{"discovered", "disabled"}, Dst: "enabled"},
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
			{Name: "add_gem_port", Src: []string{"enabled", "eapol_flow_received"}, Dst: "gem_port_added"},
			// NOTE should disabled state be diffente for oper_disabled (emulating an error) and admin_disabled (received a disabled call via VOLTHA)?
			{Name: "disable", Src: []string{"eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed"}, Dst: "disabled"},
			// EAPOL
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
			{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
			{Name: "eap_response_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_response_identity_sent"},
			{Name: "eap_response_challenge_sent", Src: []string{"eap_response_identity_sent"}, Dst: "eap_response_challenge_sent"},
			{Name: "eap_response_success_received", Src: []string{"eap_response_challenge_sent"}, Dst: "eap_response_success_received"},
			{Name: "auth_failed", Src: []string{"auth_started", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent"}, Dst: "auth_failed"},
			// DHCP
			{Name: "start_dhcp", Src: []string{"eap_response_success_received"}, Dst: "dhcp_started"},
			{Name: "dhcp_discovery_sent", Src: []string{"dhcp_started"}, Dst: "dhcp_discovery_sent"},
			{Name: "dhcp_request_sent", Src: []string{"dhcp_discovery_sent"}, Dst: "dhcp_request_sent"},
			{Name: "dhcp_ack_received", Src: []string{"dhcp_request_sent"}, Dst: "dhcp_ack_received"},
			{Name: "dhcp_failed", Src: []string{"dhcp_started", "dhcp_discovery_sent", "dhcp_request_sent"}, Dst: "dhcp_failed"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				o.logStateChange(e.Src, e.Dst)
			},
			"enter_enabled": func(event *fsm.Event) {
				msg := Message{
					Type: OnuIndication,
					Data: OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: UP,
					},
				}
				o.Channel <- msg
			},
			"enter_disabled": func(event *fsm.Event) {
				msg := Message{
					Type: OnuIndication,
					Data: OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: DOWN,
					},
				}
				o.Channel <- msg
			},
			"enter_auth_started": func(e *fsm.Event) {
				o.logStateChange(e.Src, e.Dst)
				msg := Message{
					Type: StartEAPOL,
					Data: PacketMessage{
						PonPortID: o.PonPortID,
						OnuID:     o.ID,
					},
				}
				o.Channel <- msg
			},
			"enter_auth_failed": func(e *fsm.Event) {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("ONU failed to authenticate!")
			},
			"enter_dhcp_started": func(e *fsm.Event) {
				msg := Message{
					Type: StartDHCP,
					Data: PacketMessage{
						PonPortID: o.PonPortID,
						OnuID:     o.ID,
					},
				}
				o.Channel <- msg
			},
			"enter_dhcp_failed": func(e *fsm.Event) {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("ONU failed to DHCP!")
			},
		},
	)
	return o
}

func (o Onu) logStateChange(src string, dst string) {
	onuLogger.WithFields(log.Fields{
		"OnuId":  o.ID,
		"IntfId": o.PonPortID,
		"OnuSn":  o.Sn(),
	}).Debugf("Changing ONU InternalState from %s to %s", src, dst)
}

func (o Onu) processOnuMessages(stream openolt.Openolt_EnableIndicationServer) {
	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.Sn(),
	}).Debug("Started ONU Indication Channel")

	for message := range o.Channel {
		onuLogger.WithFields(log.Fields{
			"onuID":       o.ID,
			"onuSN":       o.Sn(),
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
			log.Infof("Receive StartEAPOL message on ONU Channel")
			eapol.SendEapStart(o.ID, o.PonPortID, o.Sn(), o.InternalState, stream)
		case StartDHCP:
			log.Infof("Receive StartDHCP message on ONU Channel")
			dhcp.SendDHCPDiscovery(o.PonPortID, o.ID, o.Sn(), o.InternalState, o.HwAddress, o.CTag, stream)
		case OnuPacketOut:
			msg, _ := message.Data.(OnuPacketOutMessage)
			pkt := msg.Packet
			etherType := pkt.Layer(layers.LayerTypeEthernet).(*layers.Ethernet).EthernetType

			if etherType == layers.EthernetTypeEAPOL {
				eapol.HandleNextPacket(msg.OnuId, msg.IntfId, o.Sn(), o.InternalState, msg.Packet, stream)
			} else if packetHandlers.IsDhcpPacket(pkt) {
				// NOTE here we receive packets going from the DHCP Server to the ONU
				// for now we expect them to be double-tagged, but ideally the should be single tagged
				dhcp.HandleNextPacket(o.ID, o.PonPortID, o.Sn(), o.HwAddress, o.CTag, o.InternalState, msg.Packet, stream)
			}
		case DyingGaspIndication:
			msg, _ := message.Data.(DyingGaspIndicationMessage)
			o.sendDyingGaspInd(msg, stream)
		default:
			onuLogger.Warnf("Received unknown message data %v for type %v in OLT Channel", message.Data, message.Type)
		}
	}
}

func (o Onu) processOmciMessages(stream openolt.Openolt_EnableIndicationServer) {
	ch := omci.GetChannel()

	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.Sn(),
	}).Debug("Started OMCI Indication Channel")

	for message := range ch {
		switch message.Type {
		case omci.GemPortAdded:
			log.WithFields(log.Fields{
				"OnuId":  message.Data.OnuId,
				"IntfId": message.Data.IntfId,
			}).Infof("GemPort Added")

			// NOTE if we receive the GemPort but we don't have EAPOL flows
			// go an intermediate state, otherwise start auth
			if o.InternalState.Is("enabled") {
				if err := o.InternalState.Event("add_gem_port"); err != nil {
					log.Errorf("Can't go to gem_port_added: %v", err)
				}
			} else if o.InternalState.Is("eapol_flow_received") {
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

// NOTE handle_/process methods can change the ONU internal state as they are receiving messages
// send method should not change the ONU state

func (o Onu) sendDyingGaspInd(msg DyingGaspIndicationMessage, stream openolt.Openolt_EnableIndicationServer) error {
	alarmData := &openolt.AlarmIndication_DyingGaspInd{
		DyingGaspInd: &openolt.DyingGaspIndication{
			IntfId: msg.PonPortID,
			OnuId:  msg.OnuID,
			Status: "on",
		},
	}
	data := &openolt.Indication_AlarmInd{AlarmInd: &openolt.AlarmIndication{Data: alarmData}}

	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		onuLogger.Errorf("Failed to send DyingGaspInd : %v", err)
		return err
	}
	onuLogger.WithFields(log.Fields{
		"IntfId": msg.PonPortID,
		"OnuSn":  o.Sn(),
		"OnuId":  msg.OnuID,
	}).Info("sendDyingGaspInd")
	return nil
}

func (o Onu) sendOnuDiscIndication(msg OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId:       msg.Onu.PonPortID,
		SerialNumber: msg.Onu.SerialNumber,
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Errorf("Failed to send Indication_OnuDiscInd: %v", err)
	}

	if err := o.InternalState.Event("discover"); err != nil {
		oltLogger.WithFields(log.Fields{
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
			"OnuId":  o.ID,
		}).Infof("Failed to transition ONU to discovered state: %s", err.Error())
	}

	onuLogger.WithFields(log.Fields{
		"IntfId": msg.Onu.PonPortID,
		"OnuSn":  msg.Onu.Sn(),
		"OnuId":  o.ID,
	}).Debug("Sent Indication_OnuDiscInd")
}

func (o Onu) sendOnuIndication(msg OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	// NOTE voltha returns an ID, but if we use that ID then it complains:
	// expected_onu_id: 1, received_onu_id: 1024, event: ONU-id-mismatch, can happen if both voltha and the olt rebooted
	// so we're using the internal ID that is 1
	// o.ID = msg.OnuID

	indData := &openolt.Indication_OnuInd{OnuInd: &openolt.OnuIndication{
		IntfId:       o.PonPortID,
		OnuId:        o.ID,
		OperState:    msg.OperState.String(),
		AdminState:   o.OperState.Current(),
		SerialNumber: o.SerialNumber,
	}}
	if err := stream.Send(&openolt.Indication{Data: indData}); err != nil {
		// TODO do we need to transition to a broken state?
		log.Errorf("Failed to send Indication_OnuInd: %v", err)
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":     o.PonPortID,
		"OnuId":      o.ID,
		"OperState":  msg.OperState.String(),
		"AdminState": msg.OperState.String(),
		"OnuSn":      o.Sn(),
	}).Debug("Sent Indication_OnuInd")

}

func (o Onu) handleOmciMessage(msg OmciMessage, stream openolt.Openolt_EnableIndicationServer) {

	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket":   msg.omciMsg.Pkt,
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
		"IntfId":       o.PonPortID,
		"SerialNumber": o.SerialNumber,
		"omciPacket":   omciInd.Pkt,
	}).Tracef("Sent OMCI message")
}

func (o Onu) handleFlowUpdate(msg OnuFlowUpdateMessage, stream openolt.Openolt_EnableIndicationServer) {
	onuLogger.WithFields(log.Fields{
		"DstPort":   msg.Flow.Classifier.DstPort,
		"EthType":   fmt.Sprintf("%x", msg.Flow.Classifier.EthType),
		"FlowId":    msg.Flow.FlowId,
		"FlowType":  msg.Flow.FlowType,
		"InnerVlan": msg.Flow.Classifier.IVid,
		"IntfId":    msg.Flow.AccessIntfId,
		"IpProto":   msg.Flow.Classifier.IpProto,
		"OnuId":     msg.Flow.OnuId,
		"OnuSn":     o.Sn(),
		"OuterVlan": msg.Flow.Classifier.OVid,
		"PortNo":    msg.Flow.PortNo,
		"SrcPort":   msg.Flow.Classifier.SrcPort,
		"UniID":     msg.Flow.UniId,
	}).Debug("ONU receives Flow")

	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		// NOTE if we receive the EAPOL flows but we don't have GemPorts
		// go an intermediate state, otherwise start auth
		if o.InternalState.Is("enabled") {
			if err := o.InternalState.Event("receive_eapol_flow"); err != nil {
				log.Warnf("Can't go to eapol_flow_received: %v", err)
			}
		} else if o.InternalState.Is("gem_port_added") {
			if err := o.InternalState.Event("start_auth"); err != nil {
				log.Warnf("Can't go to auth_started: %v", err)
			}
		}
	} else if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeIPv4) &&
		msg.Flow.Classifier.SrcPort == uint32(68) &&
		msg.Flow.Classifier.DstPort == uint32(67) {
		// NOTE we are receiving mulitple DHCP flows but we shouldn't call the transition multiple times
		if err := o.InternalState.Event("start_dhcp"); err != nil {
			log.Warnf("Can't go to dhcp_started: %v", err)
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
