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
	"fmt"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"net"

	"time"

	"github.com/cboling/omci"
	"github.com/google/gopacket/layers"
	"github.com/jpillora/backoff"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	omcisim "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"github.com/opencord/voltha-protos/v2/go/tech_profile"
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
	PonPort             *PonPort
	InternalState       *fsm.FSM
	DiscoveryRetryDelay time.Duration // this is the time between subsequent Discovery Indication
	DiscoveryDelay      time.Duration // this is the time to send the first Discovery Indication

	Services []ServiceIf

	Backoff *backoff.Backoff
	// ONU State
	// PortNo comes with flows and it's used when sending packetIndications,
	// There is one PortNo per UNI Port, for now we're only storing the first one
	// FIXME add support for multiple UNIs (each UNI has a different PortNo)
	PortNo       uint32
	GemPortAdded bool
	Flows        []FlowKey
	FlowIds      []uint32 // keep track of the flows we currently have in the ONU

	OperState    *fsm.FSM
	SerialNumber *openolt.SerialNumber

	Channel chan Message // this Channel is to track state changes OMCI messages, EAPOL and DHCP packets

	// OMCI params
	tid       uint16
	hpTid     uint16
	seqNumber uint16

	DoneChannel       chan bool // this channel is used to signal once the onu is complete (when the struct is used by BBR)
	TrafficSchedulers *tech_profile.TrafficSchedulers
}

func (o *Onu) Sn() string {
	return common.OnuSnToString(o.SerialNumber)
}

func CreateONU(olt *OltDevice, pon *PonPort, id uint32, delay time.Duration, isMock bool) *Onu {
	b := &backoff.Backoff{
		//These are the defaults
		Min:    5 * time.Second,
		Max:    35 * time.Second,
		Factor: 1.5,
		Jitter: false,
	}

	o := Onu{
		ID:                  id,
		PonPortID:           pon.ID,
		PonPort:             pon,
		PortNo:              0,
		tid:                 0x1,
		hpTid:               0x8000,
		seqNumber:           0,
		DoneChannel:         make(chan bool, 1),
		GemPortAdded:        false,
		DiscoveryRetryDelay: 60 * time.Second, // this is used to send OnuDiscoveryIndications until an activate call is received
		Flows:               []FlowKey{},
		DiscoveryDelay:      delay,
		Backoff:             b,
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
			{Name: "initialize", Src: []string{"created", "disabled", "pon_disabled"}, Dst: "initialized"},
			{Name: "discover", Src: []string{"initialized"}, Dst: "discovered"},
			{Name: "enable", Src: []string{"discovered", "pon_disabled"}, Dst: "enabled"},
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
			{Name: "add_gem_port", Src: []string{"enabled", "eapol_flow_received"}, Dst: "gem_port_added"},
			// NOTE should disabled state be different for oper_disabled (emulating an error) and admin_disabled (received a disabled call via VOLTHA)?
			{Name: "disable", Src: []string{"enabled", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed", "pon_disabled"}, Dst: "disabled"},
			// ONU state when PON port is disabled but ONU is power ON(more states should be added in src?)
			{Name: "pon_disabled", Src: []string{"enabled", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed"}, Dst: "pon_disabled"},
			// BBR States
			// TODO add start OMCI state
			{Name: "send_eapol_flow", Src: []string{"initialized"}, Dst: "eapol_flow_sent"},
			{Name: "send_dhcp_flow", Src: []string{"eapol_flow_sent"}, Dst: "dhcp_flow_sent"},
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
					go o.ProcessOnuMessages(olt.enableContext, olt.OpenoltStream, nil)
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

				// clean the ONU state
				o.GemPortAdded = false
				o.PortNo = 0
				o.Flows = []FlowKey{}

				// set the OpenState to disabled
				if err := o.OperState.Event("disable"); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
					}).Errorf("Cannot change ONU OperState to down: %s", err.Error())
				}

				// send the OnuIndication DOWN event
				msg := Message{
					Type: OnuIndication,
					Data: OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: DOWN,
					},
				}
				o.Channel <- msg

				// verify all the flows removes are handled and
				// terminate the ONU's ProcessOnuMessages Go routine
				if len(o.FlowIds) == 0 {
					close(o.Channel)
				}
			},
			// BBR states
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
			case FlowAdd:
				msg, _ := message.Data.(OnuFlowUpdateMessage)
				o.handleFlowAdd(msg)
			case FlowRemoved:
				msg, _ := message.Data.(OnuFlowUpdateMessage)
				o.handleFlowRemove(msg)
			case OnuPacketOut:

				msg, _ := message.Data.(OnuPacketMessage)

				log.WithFields(log.Fields{
					"IntfId":  msg.IntfId,
					"OnuId":   msg.OnuId,
					"pktType": msg.Type,
				}).Trace("Received OnuPacketOut Message")

				service, err := o.findServiceByMacAddress(msg.MacAddress)
				if err != nil {
					onuLogger.WithFields(log.Fields{
						"IntfId":     msg.IntfId,
						"OnuId":      msg.OnuId,
						"pktType":    msg.Type,
						"MacAddress": msg.MacAddress,
						"OnuSn":      o.Sn(),
					}).Error("Cannot find Service associated with packet")
					return
				}

				service.PacketCh <- msg

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
					eapol.HandleNextPacket(msg.OnuId, msg.IntfId, msg.GemPortId, o.Sn(), o.PortNo, o.InternalState, msg.Packet, stream, client)
				} else if msg.Type == packetHandlers.DHCP {
					_ = dhcp.HandleNextBbrPacket(o.ID, o.PonPortID, o.Sn(), o.DoneChannel, msg.Packet, client)
				}
			case OmciIndication:
				msg, _ := message.Data.(OmciIndicationMessage)
				o.handleOmci(msg, client)
			case SendEapolFlow:
				o.sendEapolFlow(client)
			case SendDhcpFlow:
				o.sendDhcpFlow(client)
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
		}).Debug("UNI Link Alarm")
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
		}).Debug("UNI Link alarm sent")
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

	for _, s := range o.Services {
		go s.HandlePackets(stream)
	}

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
		_ = o.sendTestResult(msg, stream)
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

func (o *Onu) handleFlowAdd(msg OnuFlowUpdateMessage) {
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
	}).Debug("ONU receives FlowAdd")

	if msg.Flow.UniId != 0 {
		// as of now BBSim only support a single UNI, so ignore everything that is not targeted to it
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"SerialNumber": o.Sn(),
		}).Debug("Ignoring flow as it's not for the first UNI")
		return
	}

	o.FlowIds = append(o.FlowIds, msg.Flow.FlowId)
	o.addGemPortToService(uint32(msg.Flow.GemportId), msg.Flow.Classifier.EthType, msg.Flow.Classifier.OVid, msg.Flow.Classifier.IVid)

	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		// NOTE storing the PortNO, it's needed when sending PacketIndications
		o.storePortNumber(uint32(msg.Flow.PortNo))

		for _, s := range o.Services {
			s.HandleAuth(o.PonPort.Olt.OpenoltStream)
		}
	} else if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeIPv4) &&
		msg.Flow.Classifier.SrcPort == uint32(68) &&
		msg.Flow.Classifier.DstPort == uint32(67) &&
		(msg.Flow.Classifier.OPbits == 0 || msg.Flow.Classifier.OPbits == 255) {

		for _, s := range o.Services {
			s.HandleDhcp(o.PonPort.Olt.OpenoltStream, int(msg.Flow.Classifier.OVid))
		}
	}
}

func (o *Onu) handleFlowRemove(msg OnuFlowUpdateMessage) {
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        o.ID,
		"SerialNumber": o.Sn(),
		"FlowId":       msg.Flow.FlowId,
		"FlowType":     msg.Flow.FlowType,
	}).Debug("ONU receives FlowRemove")

	for idx, flow := range o.FlowIds {
		// If the gemport is found, delete it from local cache.
		if flow == msg.Flow.FlowId {
			o.FlowIds = append(o.FlowIds[:idx], o.FlowIds[idx+1:]...)
			break
		}
	}

	if len(o.FlowIds) == 0 {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"SerialNumber": o.Sn(),
		}).Info("Resetting GemPort")
		o.GemPortAdded = false

		// check if ONU delete is performed and
		// terminate the ONU's ProcessOnuMessages Go routine
		if o.InternalState.Current() == "disabled" {
			close(o.Channel)
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

		if !o.GemPortAdded {
			// NOTE this sends a CreateRequestType and BBSim replies with a CreateResponseType
			// thus we send this request only once
			gemReq, _ := omcilib.CreateGemPortRequest(o.getNextTid(false))
			sendOmciMsg(gemReq, o.PonPortID, o.ID, o.SerialNumber, "CreateGemPortRequest", client)
			o.GemPortAdded = true
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
		}).Fatalf("Failed to add EAPOL Flow")
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

	// BBR only works with a single service (ATT HSIA)
	hsia := o.Services[0].(*Service)

	classifierProto := openolt.Classifier{
		EthType: uint32(layers.EthernetTypeIPv4),
		SrcPort: uint32(68),
		DstPort: uint32(67),
		OVid:    uint32(hsia.CTag),
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

func (onu *Onu) ReDiscoverOnu() {
	// Wait for few seconds to be sure of the cleanup
	time.Sleep(5 * time.Second)

	onuLogger.WithFields(log.Fields{
		"IntfId": onu.PonPortID,
		"OnuId":  onu.ID,
		"OnuSn":  onu.Sn(),
	}).Debug("Send ONU Re-Discovery")

	// ONU Re-Discovery
	if err := onu.InternalState.Event("initialize"); err != nil {
		log.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
			"OnuId":  onu.ID,
		}).Infof("Failed to transition ONU to initialized state: %s", err.Error())
	}

	if err := onu.InternalState.Event("discover"); err != nil {
		log.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
			"OnuId":  onu.ID,
		}).Infof("Failed to transition ONU to discovered state: %s", err.Error())
	}
}

func (onu *Onu) addGemPortToService(gemport uint32, ethType uint32, oVlan uint32, iVlan uint32) {
	for _, s := range onu.Services {
		if service, ok := s.(*Service); ok {
			// EAPOL is a strange case, as packets are untagged
			// but we assume we will have a single service requiring EAPOL
			if ethType == uint32(layers.EthernetTypeEAPOL) && service.NeedsEapol {
				service.GemPort = gemport
			}

			// For DHCP services we single tag the outgoing packets,
			// thus the flow only contains the CTag and we can use that to match the service
			if ethType == uint32(layers.EthernetTypeIPv4) && service.NeedsDhcp && service.CTag == int(oVlan) {
				service.GemPort = gemport
			}

			// for dataplane services match both C and S tags
			if service.CTag == int(iVlan) && service.STag == int(oVlan) {
				service.GemPort = gemport
			}
		}
	}
}

func (onu *Onu) findServiceByMacAddress(macAddress net.HardwareAddr) (*Service, error) {
	for _, s := range onu.Services {
		service := s.(*Service)
		if service.HwAddress.String() == macAddress.String() {
			return service, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-service-with-mac-address-%s", macAddress.String())
}
