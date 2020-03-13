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
	"context"
	"errors"
	"fmt"
	"net"

	"time"

	"github.com/cboling/omci"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"github.com/opencord/bbsim/internal/bbsim/responders/igmp"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	omcisim "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	tech_profile "github.com/opencord/voltha-protos/v2/go/tech_profile"
	log "github.com/sirupsen/logrus"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

type FlowKey struct {
	ID        uint32
	Direction string
}

type Onu struct {
	ID                  uint32
	PonPortID           uint32
	PonPort             PonPort
	STag                int
	CTag                int
	Auth                bool // automatically start EAPOL if set to true
	Dhcp                bool // automatically start DHCP if set to true
	HwAddress           net.HardwareAddr
	InternalState       *fsm.FSM
	DiscoveryRetryDelay time.Duration // this is the time between subsequent Discovery Indication
	DiscoveryDelay      time.Duration // this is the time to send the first Discovery Indication

	// ONU State
	// PortNo comes with flows and it's used when sending packetIndications,
	// There is one PortNo per UNI Port, for now we're only storing the first one
	// FIXME add support for multiple UNIs
	PortNo           uint32
	DhcpFlowReceived bool
	Flows            []FlowKey

	OperState    *fsm.FSM
	SerialNumber *openolt.SerialNumber

	Channel chan Message // this Channel is to track state changes OMCI messages, EAPOL and DHCP packets

	// OMCI params
	tid        uint16
	hpTid      uint16
	seqNumber  uint16
	HasGemPort bool

	DoneChannel       chan bool // this channel is used to signal once the onu is complete (when the struct is used by BBR)
	TrafficSchedulers *tech_profile.TrafficSchedulers
}

func (o *Onu) Sn() string {
	return common.OnuSnToString(o.SerialNumber)
}

func CreateONU(olt *OltDevice, pon PonPort, id uint32, sTag int, cTag int, auth bool, dhcp bool, delay time.Duration, isMock bool) *Onu {

	o := Onu{
		ID:                  0,
		PonPortID:           pon.ID,
		PonPort:             pon,
		STag:                sTag,
		CTag:                cTag,
		Auth:                auth,
		Dhcp:                dhcp,
		HwAddress:           net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(pon.ID), byte(id)},
		PortNo:              0,
		tid:                 0x1,
		hpTid:               0x8000,
		seqNumber:           0,
		DoneChannel:         make(chan bool, 1),
		DhcpFlowReceived:    false,
		DiscoveryRetryDelay: 60 * time.Second, // this is used to send OnuDiscoveryIndications until an activate call is received
		Flows:               []FlowKey{},
		DiscoveryDelay:      delay,
	}
	o.SerialNumber = o.NewSN(olt.ID, pon.ID, id)

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
			{Name: "initialize", Src: []string{"created", "disabled"}, Dst: "initialized"},
			{Name: "discover", Src: []string{"initialized", "pon_disabled"}, Dst: "discovered"},
			{Name: "enable", Src: []string{"discovered"}, Dst: "enabled"},
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
			{Name: "add_gem_port", Src: []string{"enabled", "eapol_flow_received"}, Dst: "gem_port_added"},
			// NOTE should disabled state be different for oper_disabled (emulating an error) and admin_disabled (received a disabled call via VOLTHA)?
			{Name: "disable", Src: []string{"enabled", "eapol_flow_received", "gem_port_added", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed", "pon_disabled"}, Dst: "disabled"},
			// ONU state when PON port is disabled but ONU is power ON(more states should be added in src?)
			{Name: "pon_disabled", Src: []string{"enabled", "gem_port_added", "eapol_flow_received", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed"}, Dst: "pon_disabled"},
			// EAPOL
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed", "igmp_join_started", "igmp_left", "igmp_join_error"}, Dst: "auth_started"},
			{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
			{Name: "eap_response_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_response_identity_sent"},
			{Name: "eap_response_challenge_sent", Src: []string{"eap_response_identity_sent"}, Dst: "eap_response_challenge_sent"},
			{Name: "eap_response_success_received", Src: []string{"eap_response_challenge_sent"}, Dst: "eap_response_success_received"},
			{Name: "auth_failed", Src: []string{"auth_started", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent"}, Dst: "auth_failed"},
			// DHCP
			{Name: "start_dhcp", Src: []string{"eap_response_success_received", "dhcp_discovery_sent", "dhcp_request_sent", "dhcp_ack_received", "dhcp_failed", "igmp_join_started", "igmp_left", "igmp_join_error"}, Dst: "dhcp_started"},
			{Name: "dhcp_discovery_sent", Src: []string{"dhcp_started"}, Dst: "dhcp_discovery_sent"},
			{Name: "dhcp_request_sent", Src: []string{"dhcp_discovery_sent"}, Dst: "dhcp_request_sent"},
			{Name: "dhcp_ack_received", Src: []string{"dhcp_request_sent"}, Dst: "dhcp_ack_received"},
			{Name: "dhcp_failed", Src: []string{"dhcp_started", "dhcp_discovery_sent", "dhcp_request_sent"}, Dst: "dhcp_failed"},
			// BBR States
			// TODO add start OMCI state
			{Name: "send_eapol_flow", Src: []string{"initialized"}, Dst: "eapol_flow_sent"},
			{Name: "send_dhcp_flow", Src: []string{"eapol_flow_sent"}, Dst: "dhcp_flow_sent"},
			// IGMP
			{Name: "igmp_join_start", Src: []string{"eap_response_success_received", "gem_port_added", "eapol_flow_received", "dhcp_ack_received", "igmp_left", "igmp_join_error"}, Dst: "igmp_join_started"},
			{Name: "igmp_join_startv3", Src: []string{"eap_response_success_received", "gem_port_added", "eapol_flow_received", "dhcp_ack_received", "igmp_left", "igmp_join_error"}, Dst: "igmp_join_started"},
			{Name: "igmp_join_error", Src: []string{"igmp_join_started"}, Dst: "igmp_join_error"},
			{Name: "igmp_leave", Src: []string{"igmp_join_started", "gem_port_added", "eapol_flow_received", "eap_response_success_received", "dhcp_ack_received"}, Dst: "igmp_left"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				o.logStateChange(e.Src, e.Dst)
			},
			"enter_initialized": func(e *fsm.Event) {
				// create new channel for ProcessOnuMessages Go routine
				o.Channel = make(chan Message, 2048)

				if err := o.OperState.Event("enable"); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
					}).Errorf("Cannot change ONU OperState to up: %s", err.Error())
				}

				if !isMock {
					// start ProcessOnuMessages Go routine
					go o.ProcessOnuMessages(olt.enableContext, *olt.OpenoltStream, nil)
				}
			},
			"enter_discovered": func(e *fsm.Event) {
				msg := Message{
					Type: OnuDiscIndication,
					Data: OnuDiscIndicationMessage{
						Onu:       &o,
						OperState: UP,
					},
				}
				o.Channel <- msg
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
				if err := o.OperState.Event("disable"); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
					}).Errorf("Cannot change ONU OperState to down: %s", err.Error())
				}
				msg := Message{
					Type: OnuIndication,
					Data: OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: DOWN,
					},
				}
				o.Channel <- msg
				// terminate the ONU's ProcessOnuMessages Go routine
				close(o.Channel)
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
			"enter_eap_response_success_received": func(e *fsm.Event) {
				publishEvent("ONU-authentication-done", int32(o.PonPortID), int32(o.ID), o.Sn())
			},
			"enter_auth_failed": func(e *fsm.Event) {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("ONU failed to authenticate!")
			},
			"before_start_dhcp": func(e *fsm.Event) {
				if o.DhcpFlowReceived == false {
					e.Cancel(errors.New("cannot-go-to-dhcp-started-as-dhcp-flow-is-missing"))
				}
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
			"enter_dhcp_ack_received": func(e *fsm.Event) {
				publishEvent("ONU-DHCP-ACK-received", int32(o.PonPortID), int32(o.ID), o.Sn())
			},
			"enter_dhcp_failed": func(e *fsm.Event) {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("ONU failed to DHCP!")
			},
			"enter_eapol_flow_sent": func(e *fsm.Event) {
				msg := Message{
					Type: SendEapolFlow,
				}
				o.Channel <- msg
			},
			"enter_dhcp_flow_sent": func(e *fsm.Event) {
				msg := Message{
					Type: SendDhcpFlow,
				}
				o.Channel <- msg
			},
			"igmp_join_start": func(e *fsm.Event) {
				msg := Message{
					Type: IGMPMembershipReportV2,
				}
				o.Channel <- msg
			},
			"igmp_leave": func(e *fsm.Event) {
				msg := Message{
					Type: IGMPLeaveGroup}
				o.Channel <- msg
			},
			"igmp_join_startv3": func(e *fsm.Event) {
				msg := Message{
					Type: IGMPMembershipReportV3,
				}
				o.Channel <- msg
			},
		},
	)

	return &o
}

func (o *Onu) logStateChange(src string, dst string) {
	onuLogger.WithFields(log.Fields{
		"OnuId":  o.ID,
		"IntfId": o.PonPortID,
		"OnuSn":  o.Sn(),
	}).Debugf("Changing ONU InternalState from %s to %s", src, dst)
}

// ProcessOnuMessages starts indication channel for each ONU
func (o *Onu) ProcessOnuMessages(ctx context.Context, stream openolt.Openolt_EnableIndicationServer, client openolt.OpenoltClient) {
	onuLogger.WithFields(log.Fields{
		"onuID":   o.ID,
		"onuSN":   o.Sn(),
		"ponPort": o.PonPortID,
	}).Debug("Starting ONU Indication Channel")

loop:
	for {
		select {
		case <-ctx.Done():
			onuLogger.WithFields(log.Fields{
				"onuID": o.ID,
				"onuSN": o.Sn(),
			}).Tracef("ONU message handling canceled via context")
			break loop
		case message, ok := <-o.Channel:
			if !ok || ctx.Err() != nil {
				onuLogger.WithFields(log.Fields{
					"onuID": o.ID,
					"onuSN": o.Sn(),
				}).Tracef("ONU message handling canceled via channel close")
				break loop
			}
			onuLogger.WithFields(log.Fields{
				"onuID":       o.ID,
				"onuSN":       o.Sn(),
				"messageType": message.Type,
			}).Tracef("Received message on ONU Channel")

			switch message.Type {
			case OnuDiscIndication:
				msg, _ := message.Data.(OnuDiscIndicationMessage)
				// NOTE we need to slow down and send ONU Discovery Indication in batches to better emulate a real scenario
				time.Sleep(o.DiscoveryDelay)
				o.sendOnuDiscIndication(msg, stream)
			case OnuIndication:
				msg, _ := message.Data.(OnuIndicationMessage)
				o.sendOnuIndication(msg, stream)
			case OMCI:
				msg, _ := message.Data.(OmciMessage)
				o.handleOmciMessage(msg, stream)
			case FlowUpdate:
				msg, _ := message.Data.(OnuFlowUpdateMessage)
				o.handleFlowUpdate(msg)
			case StartEAPOL:
				log.Infof("Receive StartEAPOL message on ONU Channel")
				eapol.SendEapStart(o.ID, o.PonPortID, o.Sn(), o.PortNo, o.HwAddress, o.InternalState, stream)
			case StartDHCP:
				log.Infof("Receive StartDHCP message on ONU Channel")
				// FIXME use id, ponId as SendEapStart
				dhcp.SendDHCPDiscovery(o.PonPortID, o.ID, o.Sn(), o.PortNo, o.InternalState, o.HwAddress, o.CTag, stream)
			case OnuPacketOut:

				msg, _ := message.Data.(OnuPacketMessage)

				log.WithFields(log.Fields{
					"IntfId":  msg.IntfId,
					"OnuId":   msg.OnuId,
					"pktType": msg.Type,
				}).Trace("Received OnuPacketOut Message")

				if msg.Type == packetHandlers.EAPOL {
					eapol.HandleNextPacket(msg.OnuId, msg.IntfId, o.Sn(), o.PortNo, o.InternalState, msg.Packet, stream, client)
				} else if msg.Type == packetHandlers.DHCP {
					// NOTE here we receive packets going from the DHCP Server to the ONU
					// for now we expect them to be double-tagged, but ideally the should be single tagged
					dhcp.HandleNextPacket(o.ID, o.PonPortID, o.Sn(), o.PortNo, o.HwAddress, o.CTag, o.InternalState, msg.Packet, stream)
				}
			case OnuPacketIn:
				// NOTE we only receive BBR packets here.
				// Eapol.HandleNextPacket can handle both BBSim and BBr cases so the call is the same
				// in the DHCP case VOLTHA only act as a proxy, the behaviour is completely different thus we have a dhcp.HandleNextBbrPacket
				msg, _ := message.Data.(OnuPacketMessage)

				log.WithFields(log.Fields{
					"IntfId":  msg.IntfId,
					"OnuId":   msg.OnuId,
					"pktType": msg.Type,
				}).Trace("Received OnuPacketIn Message")

				if msg.Type == packetHandlers.EAPOL {
					eapol.HandleNextPacket(msg.OnuId, msg.IntfId, o.Sn(), o.PortNo, o.InternalState, msg.Packet, stream, client)
				} else if msg.Type == packetHandlers.DHCP {
					dhcp.HandleNextBbrPacket(o.ID, o.PonPortID, o.Sn(), o.STag, o.HwAddress, o.DoneChannel, msg.Packet, client)
				}
			case OmciIndication:
				msg, _ := message.Data.(OmciIndicationMessage)
				o.handleOmci(msg, client)
			case SendEapolFlow:
				o.sendEapolFlow(client)
			case SendDhcpFlow:
				o.sendDhcpFlow(client)
			case IGMPMembershipReportV2:
				log.Infof("Recieved IGMPMembershipReportV2 message on ONU channel")
				igmp.SendIGMPMembershipReportV2(o.PonPortID, o.ID, o.Sn(), o.PortNo, o.HwAddress, stream)
			case IGMPLeaveGroup:
				log.Infof("Recieved IGMPLeaveGroupV2 message on ONU channel")
				igmp.SendIGMPLeaveGroupV2(o.PonPortID, o.ID, o.Sn(), o.PortNo, o.HwAddress, stream)
			case IGMPMembershipReportV3:
				log.Infof("Recieved IGMPMembershipReportV3 message on ONU channel")
				igmp.SendIGMPMembershipReportV3(o.PonPortID, o.ID, o.Sn(), o.PortNo, o.HwAddress, stream)
			default:
				onuLogger.Warnf("Received unknown message data %v for type %v in OLT Channel", message.Data, message.Type)
			}
		}
	}
	onuLogger.WithFields(log.Fields{
		"onuID": o.ID,
		"onuSN": o.Sn(),
	}).Debug("Stopped handling ONU Indication Channel")
}

func (o *Onu) processOmciMessage(message omcisim.OmciChMessage, stream openolt.Openolt_EnableIndicationServer) {
	switch message.Type {
	case omcisim.UniLinkUp, omcisim.UniLinkDown:
		onuLogger.WithFields(log.Fields{
			"OnuId":  message.Data.OnuId,
			"IntfId": message.Data.IntfId,
			"Type":   message.Type,
		}).Infof("UNI Link Alarm")
		// TODO send to OLT

		omciInd := openolt.OmciIndication{
			IntfId: message.Data.IntfId,
			OnuId:  message.Data.OnuId,
			Pkt:    message.Packet,
		}

		omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
		if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
			onuLogger.WithFields(log.Fields{
				"IntfId":       o.PonPortID,
				"SerialNumber": o.Sn(),
				"Type":         message.Type,
				"omciPacket":   omciInd.Pkt,
			}).Errorf("Failed to send UNI Link Alarm: %v", err)
			return
		}

		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
			"Type":         message.Type,
			"omciPacket":   omciInd.Pkt,
		}).Info("UNI Link alarm sent")

	case omcisim.GemPortAdded:
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
			if o.Auth == true {
				if err := o.InternalState.Event("start_auth"); err != nil {
					log.Warnf("Can't go to auth_started: %v", err)
				}
			} else {
				onuLogger.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"SerialNumber": o.Sn(),
				}).Warn("Not starting authentication as Auth bit is not set in CLI parameters")
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

func (o *Onu) sendOnuDiscIndication(msg OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId:       msg.Onu.PonPortID,
		SerialNumber: msg.Onu.SerialNumber,
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Errorf("Failed to send Indication_OnuDiscInd: %v", err)
		return
	}

	onuLogger.WithFields(log.Fields{
		"IntfId": msg.Onu.PonPortID,
		"OnuSn":  msg.Onu.Sn(),
		"OnuId":  o.ID,
	}).Debug("Sent Indication_OnuDiscInd")
	publishEvent("ONU-discovery-indication-sent", int32(msg.Onu.PonPortID), int32(o.ID), msg.Onu.Sn())

	// after DiscoveryRetryDelay check if the state is the same and in case send a new OnuDiscIndication
	go func(delay time.Duration) {
		time.Sleep(delay)
		if o.InternalState.Current() == "discovered" {
			o.sendOnuDiscIndication(msg, stream)
		}
	}(o.DiscoveryRetryDelay)
}

func (o *Onu) sendOnuIndication(msg OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
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
		// NOTE do we need to transition to a broken state?
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

func (o *Onu) publishOmciEvent(msg OmciMessage) {
	if olt.PublishEvents {
		_, _, msgType, _, _, _, err := omcisim.ParsePkt(HexDecode(msg.omciMsg.Pkt))
		if err != nil {
			log.Errorf("error in getting msgType %v", err)
			return
		}
		if msgType == omcisim.MibUpload {
			o.seqNumber = 0
			publishEvent("MIB-upload-received", int32(o.PonPortID), int32(o.ID), common.OnuSnToString(o.SerialNumber))
		} else if msgType == omcisim.MibUploadNext {
			o.seqNumber++
			if o.seqNumber > 290 {
				publishEvent("MIB-upload-done", int32(o.PonPortID), int32(o.ID), common.OnuSnToString(o.SerialNumber))
			}
		}
	}
}

// Create a TestResponse packet and send it
func (o *Onu) sendTestResult(msg OmciMessage, stream openolt.Openolt_EnableIndicationServer) error {
	resp, err := omcilib.BuildTestResult(HexDecode(msg.omciMsg.Pkt))
	if err != nil {
		return err
	}

	var omciInd openolt.OmciIndication
	omciInd.IntfId = o.PonPortID
	omciInd.OnuId = o.ID
	omciInd.Pkt = resp

	omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
	if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
			"omciPacket":   omciInd.Pkt,
			"msg":          msg,
		}).Errorf("send TestResult omcisim indication failed: %v", err)
		return err
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
		"omciPacket":   omciInd.Pkt,
	}).Tracef("Sent TestResult OMCI message")

	return nil
}

func (o *Onu) handleOmciMessage(msg OmciMessage, stream openolt.Openolt_EnableIndicationServer) {

	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
		"omciPacket":   msg.omciMsg.Pkt,
	}).Tracef("Received OMCI message")

	o.publishOmciEvent(msg)

	var omciInd openolt.OmciIndication
	respPkt, err := omcisim.OmciSim(o.PonPortID, o.ID, HexDecode(msg.omciMsg.Pkt))
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
			"omciPacket":   omciInd.Pkt,
			"msg":          msg,
		}).Errorf("Error handling OMCI message %v", msg)
		return
	}

	omciInd.IntfId = o.PonPortID
	omciInd.OnuId = o.ID
	omciInd.Pkt = respPkt

	omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
	if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
			"omciPacket":   omciInd.Pkt,
			"msg":          msg,
		}).Errorf("send omcisim indication failed: %v", err)
		return
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
		"omciPacket":   omciInd.Pkt,
	}).Tracef("Sent OMCI message")

	// Test message is special, it requires sending two packets:
	//     first packet: TestResponse, says whether test was started successully, handled by omci-sim
	//     second packet, TestResult, reports the result of running the self-test
	// TestResult can come some time after a TestResponse
	//     TODO: Implement some delay between the TestResponse and the TestResult
	isTest, err := omcilib.IsTestRequest(HexDecode(msg.omciMsg.Pkt))
	if (err == nil) && (isTest) {
		o.sendTestResult(msg, stream)
	}
}

func (o *Onu) storePortNumber(portNo uint32) {
	// NOTE this needed only as long as we don't support multiple UNIs
	// we need to add support for multiple UNIs
	// the action plan is:
	// - refactor the omcisim-sim library to use https://github.com/cboling/omci instead of canned messages
	// - change the library so that it reports a single UNI and remove this workaroung
	// - add support for multiple UNIs in BBSim
	if o.PortNo == 0 || portNo < o.PortNo {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"SerialNumber": o.Sn(),
			"OnuPortNo":    o.PortNo,
			"FlowPortNo":   portNo,
		}).Debug("Storing ONU portNo")
		o.PortNo = portNo
	}
}

func (o *Onu) SetID(id uint32) {
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        id,
		"SerialNumber": o.Sn(),
	}).Debug("Storing OnuId ")
	o.ID = id
}

func (o *Onu) handleFlowUpdate(msg OnuFlowUpdateMessage) {
	onuLogger.WithFields(log.Fields{
		"DstPort":          msg.Flow.Classifier.DstPort,
		"EthType":          fmt.Sprintf("%x", msg.Flow.Classifier.EthType),
		"FlowId":           msg.Flow.FlowId,
		"FlowType":         msg.Flow.FlowType,
		"GemportId":        msg.Flow.GemportId,
		"InnerVlan":        msg.Flow.Classifier.IVid,
		"IntfId":           msg.Flow.AccessIntfId,
		"IpProto":          msg.Flow.Classifier.IpProto,
		"OnuId":            msg.Flow.OnuId,
		"OnuSn":            o.Sn(),
		"OuterVlan":        msg.Flow.Classifier.OVid,
		"PortNo":           msg.Flow.PortNo,
		"SrcPort":          msg.Flow.Classifier.SrcPort,
		"UniID":            msg.Flow.UniId,
		"ClassifierOPbits": msg.Flow.Classifier.OPbits,
	}).Debug("ONU receives Flow")

	if msg.Flow.UniId != 0 {
		// as of now BBSim only support a single UNI, so ignore everything that is not targeted to it
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"SerialNumber": o.Sn(),
		}).Debug("Ignoring flow as it's not for the first UNI")
		return
	}

	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		// NOTE storing the PortNO, it's needed when sending PacketIndications
		o.storePortNumber(uint32(msg.Flow.PortNo))

		// NOTE if we receive the EAPOL flows but we don't have GemPorts
		// go an intermediate state, otherwise start auth
		if o.InternalState.Is("enabled") {
			if err := o.InternalState.Event("receive_eapol_flow"); err != nil {
				log.Warnf("Can't go to eapol_flow_received: %v", err)
			}
		} else if o.InternalState.Is("gem_port_added") {

			if o.Auth == true {
				if err := o.InternalState.Event("start_auth"); err != nil {
					log.Warnf("Can't go to auth_started: %v", err)
				}
			} else {
				onuLogger.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"SerialNumber": o.Sn(),
				}).Warn("Not starting authentication as Auth bit is not set in CLI parameters")
			}

		}
	} else if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeIPv4) &&
		msg.Flow.Classifier.SrcPort == uint32(68) &&
		msg.Flow.Classifier.DstPort == uint32(67) &&
		msg.Flow.Classifier.OPbits == 0 {

		// keep track that we received the DHCP Flows so that we can transition the state to dhcp_started
		o.DhcpFlowReceived = true

		if o.Dhcp == true {
			// NOTE we are receiving multiple DHCP flows but we shouldn't call the transition multiple times
			if err := o.InternalState.Event("start_dhcp"); err != nil {
				log.Errorf("Can't go to dhcp_started: %v", err)
			}
		} else {
			onuLogger.WithFields(log.Fields{
				"IntfId":       o.PonPortID,
				"OnuId":        o.ID,
				"SerialNumber": o.Sn(),
			}).Warn("Not starting DHCP as Dhcp bit is not set in CLI parameters")
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

// BBR methods

func sendOmciMsg(pktBytes []byte, intfId uint32, onuId uint32, sn *openolt.SerialNumber, msgType string, client openolt.OpenoltClient) {
	omciMsg := openolt.OmciMsg{
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

func (onu *Onu) getNextTid(highPriority ...bool) uint16 {
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

// TODO move this method in responders/omcisim
func (o *Onu) StartOmci(client openolt.OpenoltClient) {
	mibReset, _ := omcilib.CreateMibResetRequest(o.getNextTid(false))
	sendOmciMsg(mibReset, o.PonPortID, o.ID, o.SerialNumber, "mibReset", client)
}

func (o *Onu) handleOmci(msg OmciIndicationMessage, client openolt.OpenoltClient) {
	msgType, packet := omcilib.DecodeOmci(msg.OmciInd.Pkt)

	log.WithFields(log.Fields{
		"IntfId":  msg.OmciInd.IntfId,
		"OnuId":   msg.OmciInd.OnuId,
		"OnuSn":   common.OnuSnToString(o.SerialNumber),
		"Pkt":     msg.OmciInd.Pkt,
		"msgType": msgType,
	}).Trace("ONU Receives OMCI Msg")
	switch msgType {
	default:
		log.WithFields(log.Fields{
			"IntfId":  msg.OmciInd.IntfId,
			"OnuId":   msg.OmciInd.OnuId,
			"OnuSn":   common.OnuSnToString(o.SerialNumber),
			"Pkt":     msg.OmciInd.Pkt,
			"msgType": msgType,
		}).Fatalf("unexpected frame: %v", packet)
	case omci.MibResetResponseType:
		mibUpload, _ := omcilib.CreateMibUploadRequest(o.getNextTid(false))
		sendOmciMsg(mibUpload, o.PonPortID, o.ID, o.SerialNumber, "mibUpload", client)
	case omci.MibUploadResponseType:
		mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
		sendOmciMsg(mibUploadNext, o.PonPortID, o.ID, o.SerialNumber, "mibUploadNext", client)
	case omci.MibUploadNextResponseType:
		o.seqNumber++

		if o.seqNumber > 290 {
			// NOTE we are done with the MIB Upload (290 is the number of messages the omci-sim library will respond to)
			galEnet, _ := omcilib.CreateGalEnetRequest(o.getNextTid(false))
			sendOmciMsg(galEnet, o.PonPortID, o.ID, o.SerialNumber, "CreateGalEnetRequest", client)
		} else {
			mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
			sendOmciMsg(mibUploadNext, o.PonPortID, o.ID, o.SerialNumber, "mibUploadNext", client)
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
			sendOmciMsg(gemReq, o.PonPortID, o.ID, o.SerialNumber, "CreateGemPortRequest", client)
			o.HasGemPort = true
		} else {
			if err := o.InternalState.Event("send_eapol_flow"); err != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("Error while transitioning ONU State %v", err)
			}
		}
	}
}

func (o *Onu) sendEapolFlow(client openolt.OpenoltClient) {

	classifierProto := openolt.Classifier{
		EthType: uint32(layers.EthernetTypeEAPOL),
		OVid:    4091,
	}

	actionProto := openolt.Action{}

	downstreamFlow := openolt.Flow{
		AccessIntfId:  int32(o.PonPortID),
		OnuId:         int32(o.ID),
		UniId:         int32(0), // NOTE do not hardcode this, we need to support multiple UNIs
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
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
		}).Fatalf("Failed to EAPOL Flow")
	}
	log.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        o.ID,
		"FlowId":       downstreamFlow.FlowId,
		"PortNo":       downstreamFlow.PortNo,
		"SerialNumber": common.OnuSnToString(o.SerialNumber),
	}).Info("Sent EAPOL Flow")
}

func (o *Onu) sendDhcpFlow(client openolt.OpenoltClient) {
	classifierProto := openolt.Classifier{
		EthType: uint32(layers.EthernetTypeIPv4),
		SrcPort: uint32(68),
		DstPort: uint32(67),
	}

	actionProto := openolt.Action{}

	downstreamFlow := openolt.Flow{
		AccessIntfId:  int32(o.PonPortID),
		OnuId:         int32(o.ID),
		UniId:         int32(0), // FIXME do not hardcode this
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
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
		}).Fatalf("Failed to send DHCP Flow")
	}
	log.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        o.ID,
		"FlowId":       downstreamFlow.FlowId,
		"PortNo":       downstreamFlow.PortNo,
		"SerialNumber": common.OnuSnToString(o.SerialNumber),
	}).Info("Sent DHCP Flow")
}

// DeleteFlow method search and delete flowKey from the onu flows slice
func (onu *Onu) DeleteFlow(key FlowKey) {
	for pos, flowKey := range onu.Flows {
		if flowKey == key {
			// delete the flowKey by shifting all flowKeys by one
			onu.Flows = append(onu.Flows[:pos], onu.Flows[pos+1:]...)
			t := make([]FlowKey, len(onu.Flows))
			copy(t, onu.Flows)
			onu.Flows = t
			break
		}
	}
}
