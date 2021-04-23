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
	"encoding/hex"
	"fmt"
	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/alarmsim"
	"sync"

	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	me "github.com/opencord/omci-lib-go/generated"
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/jpillora/backoff"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/omci-lib-go"
	"github.com/opencord/voltha-protos/v4/go/openolt"
	"github.com/opencord/voltha-protos/v4/go/tech_profile"
	log "github.com/sirupsen/logrus"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

const (
	// ONU transitions
	OnuTxInitialize            = "initialize"
	OnuTxDiscover              = "discover"
	OnuTxEnable                = "enable"
	OnuTxDisable               = "disable"
	OnuTxPonDisable            = "pon_disable"
	OnuTxStartImageDownload    = "start_image_download"
	OnuTxProgressImageDownload = "progress_image_download"
	OnuTxCompleteImageDownload = "complete_image_download"
	OnuTxFailImageDownload     = "fail_image_download"
	OnuTxActivateImage         = "activate_image"
	OnuTxCommitImage           = "commit_image"

	// ONU States
	OnuStateCreated                 = "created"
	OnuStateInitialized             = "initialized"
	OnuStateDiscovered              = "discovered"
	OnuStateEnabled                 = "enabled"
	OnuStateDisabled                = "disabled"
	OnuStatePonDisabled             = "pon_disabled"
	OnuStateImageDownloadStarted    = "image_download_started"
	OnuStateImageDownloadInProgress = "image_download_in_progress"
	OnuStateImageDownloadComplete   = "image_download_completed"
	OnuStateImageDownloadError      = "image_download_error"
	OnuStateImageActivated          = "software_image_activated"
	OnuStateImageCommitted          = "software_image_committed"

	// BBR ONU States and Transitions
	BbrOnuTxSendEapolFlow    = "send_eapol_flow"
	BbrOnuStateEapolFlowSent = "eapol_flow_sent"
	BbrOnuTxSendDhcpFlow     = "send_dhcp_flow"
	BbrOnuStateDhcpFlowSent  = "dhcp_flow_sent"
)

type FlowKey struct {
	ID        uint64
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
	PortNo  uint32
	Flows   []FlowKey
	FlowIds []uint64 // keep track of the flows we currently have in the ONU

	OperState    *fsm.FSM
	SerialNumber *openolt.SerialNumber

	Channel chan bbsim.Message // this Channel is to track state changes OMCI messages, EAPOL and DHCP packets

	// OMCI params
	MibDataSync                   uint8
	ImageSoftwareExpectedSections int
	ImageSoftwareReceivedSections int
	ActiveImageEntityId           uint16
	CommittedImageEntityId        uint16

	// OMCI params (Used in BBR)
	tid       uint16
	hpTid     uint16
	seqNumber uint16

	DoneChannel       chan bool // this channel is used to signal once the onu is complete (when the struct is used by BBR)
	TrafficSchedulers *tech_profile.TrafficSchedulers
	onuAlarmsInfoLock sync.RWMutex
	onuAlarmsInfo     map[omcilib.OnuAlarmInfoMapKey]omcilib.OnuAlarmInfo
}

func (o *Onu) Sn() string {
	return common.OnuSnToString(o.SerialNumber)
}

func CreateONU(olt *OltDevice, pon *PonPort, id uint32, delay time.Duration, isMock bool) *Onu {

	o := Onu{
		ID:                            id,
		PonPortID:                     pon.ID,
		PonPort:                       pon,
		PortNo:                        0,
		tid:                           0x1,
		hpTid:                         0x8000,
		seqNumber:                     0,
		DoneChannel:                   make(chan bool, 1),
		DiscoveryRetryDelay:           60 * time.Second, // this is used to send OnuDiscoveryIndications until an activate call is received
		Flows:                         []FlowKey{},
		DiscoveryDelay:                delay,
		MibDataSync:                   0,
		ImageSoftwareExpectedSections: 0, // populated during OMCI StartSoftwareDownloadRequest
		ImageSoftwareReceivedSections: 0,
		ActiveImageEntityId:           0, // when we start the SoftwareImage with ID 0 is active and committed
		CommittedImageEntityId:        0,
	}
	o.SerialNumber = NewSN(olt.ID, pon.ID, id)
	// NOTE this state machine is used to track the operational
	// state as requested by VOLTHA
	o.OperState = getOperStateFSM(func(e *fsm.Event) {
		onuLogger.WithFields(log.Fields{
			"ID": o.ID,
		}).Debugf("Changing ONU OperState from %s to %s", e.Src, e.Dst)
	})
	o.onuAlarmsInfo = make(map[omcilib.OnuAlarmInfoMapKey]omcilib.OnuAlarmInfo)
	// NOTE this state machine is used to activate the OMCI, EAPOL and DHCP clients
	o.InternalState = fsm.NewFSM(
		OnuStateCreated,
		fsm.Events{
			// DEVICE Lifecycle
			{Name: OnuTxInitialize, Src: []string{OnuStateCreated, OnuStateDisabled, OnuStatePonDisabled}, Dst: OnuStateInitialized},
			{Name: OnuTxDiscover, Src: []string{OnuStateInitialized}, Dst: OnuStateDiscovered},
			{Name: OnuTxEnable, Src: []string{OnuStateDiscovered, OnuStatePonDisabled}, Dst: OnuStateEnabled},
			// NOTE should disabled state be different for oper_disabled (emulating an error) and admin_disabled (received a disabled call via VOLTHA)?
			{Name: OnuTxDisable, Src: []string{OnuStateEnabled, OnuStatePonDisabled, OnuStateImageActivated, OnuStateImageDownloadError, OnuStateImageCommitted}, Dst: OnuStateDisabled},
			// ONU state when PON port is disabled but ONU is power ON(more states should be added in src?)
			{Name: OnuTxPonDisable, Src: []string{OnuStateEnabled, OnuStateImageActivated, OnuStateImageDownloadError, OnuStateImageCommitted}, Dst: OnuStatePonDisabled},
			// Software Image Download related states
			{Name: OnuTxStartImageDownload, Src: []string{OnuStateEnabled, OnuStateImageDownloadComplete, OnuStateImageDownloadError}, Dst: OnuStateImageDownloadStarted},
			{Name: OnuTxProgressImageDownload, Src: []string{OnuStateImageDownloadStarted}, Dst: OnuStateImageDownloadInProgress},
			{Name: OnuTxCompleteImageDownload, Src: []string{OnuStateImageDownloadInProgress}, Dst: OnuStateImageDownloadComplete},
			{Name: OnuTxFailImageDownload, Src: []string{OnuStateImageDownloadInProgress}, Dst: OnuStateImageDownloadError},
			{Name: OnuTxActivateImage, Src: []string{OnuStateImageDownloadComplete}, Dst: OnuStateImageActivated},
			{Name: OnuTxCommitImage, Src: []string{OnuStateEnabled}, Dst: OnuStateImageCommitted}, // the image is committed after a ONU reboot
			// BBR States
			// TODO add start OMCI state
			{Name: BbrOnuTxSendEapolFlow, Src: []string{OnuStateInitialized}, Dst: BbrOnuStateEapolFlowSent},
			{Name: BbrOnuTxSendDhcpFlow, Src: []string{BbrOnuStateEapolFlowSent}, Dst: BbrOnuStateDhcpFlowSent},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				o.logStateChange(e.Src, e.Dst)
			},
			fmt.Sprintf("enter_%s", OnuStateInitialized): func(e *fsm.Event) {
				// create new channel for ProcessOnuMessages Go routine
				o.Channel = make(chan bbsim.Message, 2048)

				if err := o.OperState.Event(OnuTxEnable); err != nil {
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
			fmt.Sprintf("enter_%s", OnuStateDiscovered): func(e *fsm.Event) {
				msg := bbsim.Message{
					Type: bbsim.OnuDiscIndication,
					Data: bbsim.OnuDiscIndicationMessage{
						OperState: bbsim.UP,
					},
				}
				o.Channel <- msg
			},
			fmt.Sprintf("enter_%s", OnuStateEnabled): func(event *fsm.Event) {

				if used, sn := o.PonPort.isOnuIdAllocated(o.ID); used {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"SerialNumber": o.Sn(),
					}).Errorf("onu-id-duplicated-with-%s", common.OnuSnToString(sn))
					return
				} else {
					o.PonPort.storeOnuId(o.ID, o.SerialNumber)
				}

				msg := bbsim.Message{
					Type: bbsim.OnuIndication,
					Data: bbsim.OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: bbsim.UP,
					},
				}
				o.Channel <- msg

				// Once the ONU is enabled start listening for packets
				for _, s := range o.Services {
					s.Initialize(o.PonPort.Olt.OpenoltStream)
				}
			},
			fmt.Sprintf("enter_%s", OnuStateDisabled): func(event *fsm.Event) {

				o.cleanupOnuState()

				// set the OperState to disabled
				if err := o.OperState.Event("disable"); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
					}).Errorf("Cannot change ONU OperState to down: %s", err.Error())
				}
				// send the OnuIndication DOWN event
				msg := bbsim.Message{
					Type: bbsim.OnuIndication,
					Data: bbsim.OnuIndicationMessage{
						OnuSN:     o.SerialNumber,
						PonPortID: o.PonPortID,
						OperState: bbsim.DOWN,
					},
				}
				o.Channel <- msg

				// verify all the flows removes are handled and
				// terminate the ONU's ProcessOnuMessages Go routine
				if len(o.FlowIds) == 0 {
					close(o.Channel)
				}

				for _, s := range o.Services {
					s.Disable()
				}

			},
			fmt.Sprintf("enter_%s", OnuStatePonDisabled): func(event *fsm.Event) {
				o.cleanupOnuState()
			},
			// BBR states
			fmt.Sprintf("enter_%s", BbrOnuStateEapolFlowSent): func(e *fsm.Event) {
				msg := bbsim.Message{
					Type: bbsim.SendEapolFlow,
				}
				o.Channel <- msg
			},
			fmt.Sprintf("enter_%s", BbrOnuStateDhcpFlowSent): func(e *fsm.Event) {
				msg := bbsim.Message{
					Type: bbsim.SendDhcpFlow,
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

// cleanupOnuState this method is to clean the local state when the ONU is disabled
func (o *Onu) cleanupOnuState() {
	// clean the ONU state
	o.PortNo = 0
	o.Flows = []FlowKey{}
	o.PonPort.removeOnuId(o.ID)
	o.PonPort.removeAllocId(o.SerialNumber)
	o.PonPort.removeGemPortBySn(o.SerialNumber)

	o.onuAlarmsInfoLock.Lock()
	o.onuAlarmsInfo = make(map[omcilib.OnuAlarmInfoMapKey]omcilib.OnuAlarmInfo) //Basically reset everything on onu disable
	o.onuAlarmsInfoLock.Unlock()
}

// ProcessOnuMessages starts indication channel for each ONU
func (o *Onu) ProcessOnuMessages(ctx context.Context, stream openolt.Openolt_EnableIndicationServer, client openolt.OpenoltClient) {
	onuLogger.WithFields(log.Fields{
		"onuID":   o.ID,
		"onuSN":   o.Sn(),
		"ponPort": o.PonPortID,
		"stream":  stream,
	}).Debug("Starting ONU Indication Channel")

loop:
	for {
		select {
		case <-ctx.Done():
			onuLogger.WithFields(log.Fields{
				"onuID": o.ID,
				"onuSN": o.Sn(),
			}).Debug("ONU message handling canceled via context")
			break loop
		case <-stream.Context().Done():
			onuLogger.WithFields(log.Fields{
				"onuID": o.ID,
				"onuSN": o.Sn(),
			}).Debug("ONU message handling canceled via stream context")
			break loop
		case message, ok := <-o.Channel:
			if !ok || ctx.Err() != nil {
				onuLogger.WithFields(log.Fields{
					"onuID": o.ID,
					"onuSN": o.Sn(),
				}).Debug("ONU message handling canceled via channel close")
				break loop
			}
			onuLogger.WithFields(log.Fields{
				"onuID":       o.ID,
				"onuSN":       o.Sn(),
				"messageType": message.Type,
			}).Tracef("Received message on ONU Channel")

			switch message.Type {
			case bbsim.OnuDiscIndication:
				msg, _ := message.Data.(bbsim.OnuDiscIndicationMessage)
				// NOTE we need to slow down and send ONU Discovery Indication in batches to better emulate a real scenario
				time.Sleep(o.DiscoveryDelay)
				o.sendOnuDiscIndication(msg, stream)
			case bbsim.OnuIndication:
				msg, _ := message.Data.(bbsim.OnuIndicationMessage)
				o.sendOnuIndication(msg, stream)
			case bbsim.OMCI:
				// these are OMCI messages received by the ONU
				msg, _ := message.Data.(bbsim.OmciMessage)
				o.handleOmciRequest(msg, stream)
			case bbsim.UniStatusAlarm:
				msg, _ := message.Data.(bbsim.UniStatusAlarmMessage)
				onuAlarmMapKey := omcilib.OnuAlarmInfoMapKey{
					MeInstance: msg.EntityID,
					MeClassID:  me.PhysicalPathTerminationPointEthernetUniClassID,
				}
				seqNo := o.IncrementAlarmSequenceNumber(onuAlarmMapKey)
				o.onuAlarmsInfoLock.Lock()
				var alarmInfo = o.onuAlarmsInfo[onuAlarmMapKey]
				pkt, alarmBitMap := omcilib.CreateUniStatusAlarm(msg.RaiseOMCIAlarm, msg.EntityID, seqNo)
				if pkt != nil { //pkt will be nil if we are unable to create the alarm
					if err := o.sendOmciIndication(pkt, 0, stream); err != nil {
						onuLogger.WithFields(log.Fields{
							"IntfId":       o.PonPortID,
							"SerialNumber": o.Sn(),
							"omciPacket":   pkt,
							"adminState":   msg.AdminState,
							"entityID":     msg.EntityID,
						}).Errorf("failed-to-send-UNI-Link-Alarm: %v", err)
						alarmInfo.SequenceNo--
					}
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"SerialNumber": o.Sn(),
						"omciPacket":   pkt,
						"adminState":   msg.AdminState,
						"entityID":     msg.EntityID,
					}).Trace("UNI-Link-alarm-sent")
					if alarmBitMap == [28]byte{0} {
						delete(o.onuAlarmsInfo, onuAlarmMapKey)
					} else {
						alarmInfo.AlarmBitMap = alarmBitMap
						o.onuAlarmsInfo[onuAlarmMapKey] = alarmInfo
					}
				}
				o.onuAlarmsInfoLock.Unlock()
			case bbsim.FlowAdd:
				msg, _ := message.Data.(bbsim.OnuFlowUpdateMessage)
				o.handleFlowAdd(msg)
			case bbsim.FlowRemoved:
				msg, _ := message.Data.(bbsim.OnuFlowUpdateMessage)
				o.handleFlowRemove(msg)
			case bbsim.OnuPacketOut:

				msg, _ := message.Data.(bbsim.OnuPacketMessage)

				onuLogger.WithFields(log.Fields{
					"IntfId":  msg.IntfId,
					"OnuId":   msg.OnuId,
					"pktType": msg.Type,
				}).Trace("Received OnuPacketOut Message")

				if msg.Type == packetHandlers.EAPOL || msg.Type == packetHandlers.DHCP {

					service, err := o.findServiceByMacAddress(msg.MacAddress)
					if err != nil {
						onuLogger.WithFields(log.Fields{
							"IntfId":     msg.IntfId,
							"OnuId":      msg.OnuId,
							"pktType":    msg.Type,
							"MacAddress": msg.MacAddress,
							"Pkt":        hex.EncodeToString(msg.Packet.Data()),
							"OnuSn":      o.Sn(),
						}).Error("Cannot find Service associated with packet")
						return
					}
					service.PacketCh <- msg
				} else if msg.Type == packetHandlers.IGMP {
					// if it's an IGMP packet we assume we have a single IGMP service
					for _, s := range o.Services {
						service := s.(*Service)

						if service.NeedsIgmp {
							service.PacketCh <- msg
						}
					}
				}

			case bbsim.OnuPacketIn:
				// NOTE we only receive BBR packets here.
				// Eapol.HandleNextPacket can handle both BBSim and BBr cases so the call is the same
				// in the DHCP case VOLTHA only act as a proxy, the behaviour is completely different thus we have a dhcp.HandleNextBbrPacket
				msg, _ := message.Data.(bbsim.OnuPacketMessage)

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
				// BBR specific messages
			case bbsim.OmciIndication:
				// these are OMCI messages received by BBR (VOLTHA emulator)
				msg, _ := message.Data.(bbsim.OmciIndicationMessage)
				o.handleOmciResponse(msg, client)
			case bbsim.SendEapolFlow:
				o.sendEapolFlow(client)
			case bbsim.SendDhcpFlow:
				o.sendDhcpFlow(client)
			default:
				onuLogger.Warnf("Received unknown message data %v for type %v in OLT Channel", message.Data, message.Type)
			}
		}
	}
	onuLogger.WithFields(log.Fields{
		"onuID":  o.ID,
		"onuSN":  o.Sn(),
		"stream": stream,
	}).Debug("Stopped handling ONU Indication Channel")
}

func NewSN(oltid int, intfid uint32, onuid uint32) *openolt.SerialNumber {
	sn := new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}
	return sn
}

func (o *Onu) sendOnuDiscIndication(msg bbsim.OnuDiscIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	discoverData := &openolt.Indication_OnuDiscInd{OnuDiscInd: &openolt.OnuDiscIndication{
		IntfId:       o.PonPortID,
		SerialNumber: o.SerialNumber,
	}}

	if err := stream.Send(&openolt.Indication{Data: discoverData}); err != nil {
		log.Errorf("Failed to send Indication_OnuDiscInd: %v", err)
		return
	}

	onuLogger.WithFields(log.Fields{
		"IntfId": o.PonPortID,
		"OnuSn":  o.Sn(),
		"OnuId":  o.ID,
	}).Debug("Sent Indication_OnuDiscInd")
	publishEvent("ONU-discovery-indication-sent", int32(o.PonPortID), int32(o.ID), o.Sn())

	// after DiscoveryRetryDelay check if the state is the same and in case send a new OnuDiscIndication
	go func(delay time.Duration) {
		time.Sleep(delay)
		if o.InternalState.Current() == OnuStateDiscovered {
			o.sendOnuDiscIndication(msg, stream)
		}
	}(o.DiscoveryRetryDelay)
}

func (o *Onu) sendOnuIndication(msg bbsim.OnuIndicationMessage, stream openolt.Openolt_EnableIndicationServer) {
	// NOTE the ONU ID is set by VOLTHA in the ActivateOnu call (via openolt.proto)
	// and stored in the Onu struct via onu.SetID

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
		return
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":      o.PonPortID,
		"OnuId":       o.ID,
		"VolthaOnuId": msg.OnuID,
		"OperState":   msg.OperState.String(),
		"AdminState":  msg.OperState.String(),
		"OnuSn":       o.Sn(),
	}).Debug("Sent Indication_OnuInd")

}

func (o *Onu) HandleShutdownONU() error {

	dyingGasp := pb.ONUAlarmRequest{
		AlarmType:    "DYING_GASP",
		SerialNumber: o.Sn(),
		Status:       "on",
	}

	if err := alarmsim.SimulateOnuAlarm(&dyingGasp, o.ID, o.PonPortID, o.PonPort.Olt.channel); err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Errorf("Cannot send Dying Gasp: %s", err.Error())
		return err
	}

	losReq := pb.ONUAlarmRequest{
		AlarmType:    "ONU_ALARM_LOS",
		SerialNumber: o.Sn(),
		Status:       "on",
	}

	if err := alarmsim.SimulateOnuAlarm(&losReq, o.ID, o.PonPortID, o.PonPort.Olt.channel); err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Errorf("Cannot send LOS: %s", err.Error())

		return err
	}
	o.SendOMCIAlarmNotificationMsg(true, losReq.AlarmType)
	// TODO if it's the last ONU on the PON, then send a PON LOS

	if err := o.InternalState.Event(OnuTxDisable); err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Errorf("Cannot shutdown ONU: %s", err.Error())
		return err
	}

	return nil
}

func (o *Onu) HandlePowerOnONU() error {
	intitalState := o.InternalState.Current()

	// initialize the ONU
	if intitalState == OnuStateCreated || intitalState == OnuStateDisabled {
		if err := o.InternalState.Event(OnuTxInitialize); err != nil {
			onuLogger.WithFields(log.Fields{
				"OnuId":  o.ID,
				"IntfId": o.PonPortID,
				"OnuSn":  o.Sn(),
			}).Errorf("Cannot poweron ONU: %s", err.Error())
			return err
		}
	}

	// turn off the LOS Alarm
	losReq := pb.ONUAlarmRequest{
		AlarmType:    "ONU_ALARM_LOS",
		SerialNumber: o.Sn(),
		Status:       "off",
	}

	if err := alarmsim.SimulateOnuAlarm(&losReq, o.ID, o.PonPortID, o.PonPort.Olt.channel); err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Errorf("Cannot send LOS: %s", err.Error())
		return err
	}
	o.SendOMCIAlarmNotificationMsg(false, losReq.AlarmType)

	// Send a ONU Discovery indication
	if err := o.InternalState.Event(OnuTxDiscover); err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Errorf("Cannot poweron ONU: %s", err.Error())
		return err
	}

	// move o directly to enable state only when its a powercycle case
	// in case of first time o poweron o will be moved to enable on
	// receiving ActivateOnu request from openolt adapter
	if intitalState == OnuStateDisabled {
		if err := o.InternalState.Event(OnuTxEnable); err != nil {
			onuLogger.WithFields(log.Fields{
				"OnuId":  o.ID,
				"IntfId": o.PonPortID,
				"OnuSn":  o.Sn(),
			}).Errorf("Cannot enable ONU: %s", err.Error())
			return err
		}
	}

	return nil
}

func (o *Onu) SetAlarm(alarmType string, status string) error {
	alarmReq := pb.ONUAlarmRequest{
		AlarmType:    alarmType,
		SerialNumber: o.Sn(),
		Status:       status,
	}

	err := alarmsim.SimulateOnuAlarm(&alarmReq, o.ID, o.PonPortID, o.PonPort.Olt.channel)
	if err != nil {
		return err
	}
	raiseAlarm := false
	if alarmReq.Status == "on" {
		raiseAlarm = true
	}
	o.SendOMCIAlarmNotificationMsg(raiseAlarm, alarmReq.AlarmType)
	return nil
}

func (o *Onu) publishOmciEvent(msg bbsim.OmciMessage) {
	if olt.PublishEvents {
		_, omciMsg, err := omcilib.ParseOpenOltOmciPacket(msg.OmciPkt.Data())
		if err != nil {
			log.Errorf("error in getting msgType %v", err)
			return
		}
		if omciMsg.MessageType == omci.MibUploadRequestType {
			o.seqNumber = 0
			publishEvent("MIB-upload-received", int32(o.PonPortID), int32(o.ID), common.OnuSnToString(o.SerialNumber))
		} else if omciMsg.MessageType == omci.MibUploadNextRequestType {
			o.seqNumber++
			if o.seqNumber > 290 {
				publishEvent("MIB-upload-done", int32(o.PonPortID), int32(o.ID), common.OnuSnToString(o.SerialNumber))
			}
		}
	}
}

// Create a TestResponse packet and send it
func (o *Onu) sendTestResult(msg bbsim.OmciMessage, stream openolt.Openolt_EnableIndicationServer) error {
	resp, err := omcilib.BuildTestResult(msg.OmciPkt.Data())
	if err != nil {
		return err
	}

	var omciInd openolt.OmciIndication
	omciInd.IntfId = o.PonPortID
	omciInd.OnuId = o.ID
	omciInd.Pkt = resp

	omci := &openolt.Indication_OmciInd{OmciInd: &omciInd}
	if err := stream.Send(&openolt.Indication{Data: omci}); err != nil {
		return err
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
		"omciPacket":   omciInd.Pkt,
	}).Tracef("Sent TestResult OMCI message")

	return nil
}

// handleOmciRequest is responsible to parse the OMCI packets received from the openolt adapter
// and generate the appropriate response to it
func (o *Onu) handleOmciRequest(msg bbsim.OmciMessage, stream openolt.Openolt_EnableIndicationServer) {

	onuLogger.WithFields(log.Fields{
		"omciMsgType":  msg.OmciMsg.MessageType,
		"transCorrId":  strconv.FormatInt(int64(msg.OmciMsg.TransactionID), 16),
		"DeviceIdent":  msg.OmciMsg.DeviceIdentifier,
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
	}).Trace("omci-message-decoded")

	var responsePkt []byte
	var errResp error
	switch msg.OmciMsg.MessageType {
	case omci.MibResetRequestType:
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"SerialNumber": o.Sn(),
		}).Debug("received-mib-reset-request")
		if responsePkt, errResp = omcilib.CreateMibResetResponse(msg.OmciMsg.TransactionID); errResp == nil {
			o.MibDataSync = 0

			// if the MIB reset is successful then remove all the stored AllocIds and GemPorts
			o.PonPort.removeAllocId(o.SerialNumber)
			o.PonPort.removeGemPortBySn(o.SerialNumber)
		}
	case omci.MibUploadRequestType:
		responsePkt, _ = omcilib.CreateMibUploadResponse(msg.OmciMsg.TransactionID)
	case omci.MibUploadNextRequestType:
		responsePkt, _ = omcilib.CreateMibUploadNextResponse(msg.OmciPkt, msg.OmciMsg, o.MibDataSync)
	case omci.GetRequestType:
		onuDown := o.OperState.Current() == "down"
		responsePkt, _ = omcilib.CreateGetResponse(msg.OmciPkt, msg.OmciMsg, o.SerialNumber, o.MibDataSync, o.ActiveImageEntityId, o.CommittedImageEntityId, onuDown)
	case omci.SetRequestType:
		success := true
		msgObj, _ := omcilib.ParseSetRequest(msg.OmciPkt)
		switch msgObj.EntityClass {
		case me.PhysicalPathTerminationPointEthernetUniClassID:
			// if we're Setting a PPTP state
			// we need to send the appropriate alarm

			if msgObj.EntityInstance == 257 {
				// for now we're only caring about the first UNI
				// NOTE that the EntityID for the UNI port is for now hardcoded in
				// omci/mibpackets.go where the PhysicalPathTerminationPointEthernetUni
				// are reported during the MIB Upload sequence
				adminState := msgObj.Attributes["AdministrativeState"].(uint8)
				raiseOMCIAlarm := false
				if adminState == 1 {
					raiseOMCIAlarm = true
					// set the OperState to disabled
					if err := o.OperState.Event(OnuTxDisable); err != nil {
						onuLogger.WithFields(log.Fields{
							"OnuId":  o.ID,
							"IntfId": o.PonPortID,
							"OnuSn":  o.Sn(),
						}).Errorf("Cannot change ONU OperState to down: %s", err.Error())
					}
				} else {
					// set the OperState to enabled
					if err := o.OperState.Event(OnuTxEnable); err != nil {
						onuLogger.WithFields(log.Fields{
							"OnuId":  o.ID,
							"IntfId": o.PonPortID,
							"OnuSn":  o.Sn(),
						}).Errorf("Cannot change ONU OperState to up: %s", err.Error())
					}
				}
				msg := bbsim.Message{
					Type: bbsim.UniStatusAlarm,
					Data: bbsim.UniStatusAlarmMessage{
						OnuSN:          o.SerialNumber,
						OnuID:          o.ID,
						AdminState:     adminState,
						EntityID:       msgObj.EntityInstance,
						RaiseOMCIAlarm: raiseOMCIAlarm,
					},
				}
				o.Channel <- msg
			}
		case me.TContClassID:
			allocId := msgObj.Attributes["AllocId"].(uint16)

			// if the AllocId is 255 (0xFF) or 65535 (0xFFFF) it means we are removing it,
			// otherwise we are adding it
			if allocId == 255 || allocId == 65535 {
				onuLogger.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"TContId":      msgObj.EntityInstance,
					"AllocId":      allocId,
					"SerialNumber": o.Sn(),
				}).Trace("freeing-alloc-id-via-omci")
				o.PonPort.removeAllocId(o.SerialNumber)
			} else {
				if used, sn := o.PonPort.isAllocIdAllocated(allocId); used {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"AllocId":      allocId,
						"SerialNumber": o.Sn(),
					}).Errorf("allocid-already-allocated-to-onu-with-sn-%s", common.OnuSnToString(sn))
					success = false
				} else {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"TContId":      msgObj.EntityInstance,
						"AllocId":      allocId,
						"SerialNumber": o.Sn(),
					}).Trace("storing-alloc-id-via-omci")
					o.PonPort.storeAllocId(allocId, o.SerialNumber)
				}
			}

		}

		if success {
			if responsePkt, errResp = omcilib.CreateSetResponse(msg.OmciPkt, msg.OmciMsg, me.Success); errResp == nil {
				o.MibDataSync++
			}
		} else {
			responsePkt, _ = omcilib.CreateSetResponse(msg.OmciPkt, msg.OmciMsg, me.AttributeFailure)
		}
	case omci.CreateRequestType:
		// check for GemPortNetworkCtp and make sure there are no duplicates on the same PON
		var used bool
		var sn *openolt.SerialNumber
		msgObj, err := omcilib.ParseCreateRequest(msg.OmciPkt)
		if err == nil {
			if msgObj.EntityClass == me.GemPortNetworkCtpClassID {
				// GemPort 4069 is reserved for multicast and shared across ONUs
				if msgObj.EntityInstance != 4069 {
					if used, sn = o.PonPort.isGemPortAllocated(msgObj.EntityInstance); used {
						onuLogger.WithFields(log.Fields{
							"IntfId":       o.PonPortID,
							"OnuId":        o.ID,
							"GemPortId":    msgObj.EntityInstance,
							"SerialNumber": o.Sn(),
						}).Errorf("gemport-already-allocated-to-onu-with-sn-%s", common.OnuSnToString(sn))
					} else {
						onuLogger.WithFields(log.Fields{
							"IntfId":       o.PonPortID,
							"OnuId":        o.ID,
							"GemPortId":    msgObj.EntityInstance,
							"SerialNumber": o.Sn(),
						}).Trace("storing-gem-port-id-via-omci")
						o.PonPort.storeGemPort(msgObj.EntityInstance, o.SerialNumber)
					}
				}
			}
		}

		// if the gemPort is valid then increment the MDS and return a successful response
		// otherwise fail the request
		// for now the CreateRequeste for the gemPort is the only one that can fail, if we start supporting multiple
		// validation this check will need to be rewritten
		if !used {
			if responsePkt, errResp = omcilib.CreateCreateResponse(msg.OmciPkt, msg.OmciMsg, me.Success); errResp == nil {
				o.MibDataSync++
			}
		} else {
			responsePkt, _ = omcilib.CreateCreateResponse(msg.OmciPkt, msg.OmciMsg, me.ProcessingError)
		}
	case omci.DeleteRequestType:
		msgObj, err := omcilib.ParseDeleteRequest(msg.OmciPkt)
		if err == nil {
			if msgObj.EntityClass == me.GemPortNetworkCtpClassID {
				onuLogger.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"GemPortId":    msgObj.EntityInstance,
					"SerialNumber": o.Sn(),
				}).Trace("freeing-gem-port-id-via-omci")
				o.PonPort.removeGemPort(msgObj.EntityInstance)
			}
		}

		if responsePkt, errResp = omcilib.CreateDeleteResponse(msg.OmciPkt, msg.OmciMsg); errResp == nil {
			o.MibDataSync++
		}
	case omci.RebootRequestType:

		responsePkt, _ = omcilib.CreateRebootResponse(msg.OmciPkt, msg.OmciMsg)

		// powercycle the ONU
		// we run this in a separate goroutine so that
		// the RebootRequestResponse is sent to VOLTHA
		go func() {
			if err := o.Reboot(10 * time.Second); err != nil {
				log.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"SerialNumber": o.Sn(),
					"err":          err,
				}).Error("cannot-reboot-onu-after-omci-reboot-request")
			}
		}()
	case omci.TestRequestType:

		// Test message is special, it requires sending two packets:
		//     first packet: TestResponse, says whether test was started successully, handled by omci-sim
		//     second packet, TestResult, reports the result of running the self-test
		// TestResult can come some time after a TestResponse
		//     TODO: Implement some delay between the TestResponse and the TestResult
		isTest, err := omcilib.IsTestRequest(msg.OmciPkt.Data())
		if (err == nil) && (isTest) {
			if sendErr := o.sendTestResult(msg, stream); sendErr != nil {
				onuLogger.WithFields(log.Fields{
					"IntfId":       o.PonPortID,
					"OnuId":        o.ID,
					"SerialNumber": o.Sn(),
					"omciPacket":   msg.OmciPkt.Data(),
					"msg":          msg,
					"err":          sendErr,
				}).Error("send-TestResult-indication-failed")
			}
		}
	case omci.SynchronizeTimeRequestType:
		// MDS counter increment is not required for this message type
		responsePkt, _ = omcilib.CreateSyncTimeResponse(msg.OmciPkt, msg.OmciMsg)
	case omci.StartSoftwareDownloadRequestType:

		o.ImageSoftwareReceivedSections = 0

		o.ImageSoftwareExpectedSections = omcilib.ComputeDownloadSectionsCount(msg.OmciPkt)

		if responsePkt, errResp = omcilib.CreateStartSoftwareDownloadResponse(msg.OmciPkt, msg.OmciMsg); errResp == nil {
			o.MibDataSync++
			if err := o.InternalState.Event(OnuTxStartImageDownload); err != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
					"Err":    err.Error(),
				}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageDownloadStarted)
			}
		} else {
			onuLogger.WithFields(log.Fields{
				"OmciMsgType":  msg.OmciMsg.MessageType,
				"TransCorrId":  msg.OmciMsg.TransactionID,
				"Err":          errResp.Error(),
				"IntfId":       o.PonPortID,
				"SerialNumber": o.Sn(),
			}).Error("error-while-processing-start-software-download-request")
		}
	case omci.DownloadSectionRequestType:
		if msgObj, err := omcilib.ParseDownloadSectionRequest(msg.OmciPkt); err == nil {
			onuLogger.WithFields(log.Fields{
				"OmciMsgType":    msg.OmciMsg.MessageType,
				"TransCorrId":    msg.OmciMsg.TransactionID,
				"EntityInstance": msgObj.EntityInstance,
				"SectionNumber":  msgObj.SectionNumber,
				"SectionData":    msgObj.SectionData,
			}).Trace("received-download-section-request")
			o.ImageSoftwareReceivedSections++
			if o.InternalState.Current() != OnuStateImageDownloadInProgress {
				if err := o.InternalState.Event(OnuTxProgressImageDownload); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
						"Err":    err.Error(),
					}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageDownloadInProgress)
				}
			}
		}
	case omci.DownloadSectionRequestWithResponseType:
		// NOTE we only need to respond if an ACK is requested
		responsePkt, errResp = omcilib.CreateDownloadSectionResponse(msg.OmciPkt, msg.OmciMsg)
		if errResp != nil {
			onuLogger.WithFields(log.Fields{
				"OmciMsgType":  msg.OmciMsg.MessageType,
				"TransCorrId":  msg.OmciMsg.TransactionID,
				"Err":          errResp.Error(),
				"IntfId":       o.PonPortID,
				"SerialNumber": o.Sn(),
			}).Error("error-while-processing-create-download-section-response")
			return
		}
		o.ImageSoftwareReceivedSections++

	case omci.EndSoftwareDownloadRequestType:

		// In the startSoftwareDownload we get the image size and the window size.
		// We calculate how many DownloadSection we should receive and validate
		// that we got the correct amount when we receive this message
		success := true
		if o.ImageSoftwareExpectedSections != o.ImageSoftwareReceivedSections {
			onuLogger.WithFields(log.Fields{
				"OnuId":            o.ID,
				"IntfId":           o.PonPortID,
				"OnuSn":            o.Sn(),
				"ExpectedSections": o.ImageSoftwareExpectedSections,
				"ReceivedSections": o.ImageSoftwareReceivedSections,
			}).Errorf("onu-did-not-receive-all-image-sections")
			success = false
		}

		if success {
			if responsePkt, errResp = omcilib.CreateEndSoftwareDownloadResponse(msg.OmciPkt, msg.OmciMsg, me.Success); errResp == nil {
				o.MibDataSync++
				if err := o.InternalState.Event(OnuTxCompleteImageDownload); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
						"Err":    err.Error(),
					}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageDownloadComplete)
				}
			} else {
				onuLogger.WithFields(log.Fields{
					"OmciMsgType":  msg.OmciMsg.MessageType,
					"TransCorrId":  msg.OmciMsg.TransactionID,
					"Err":          errResp.Error(),
					"IntfId":       o.PonPortID,
					"SerialNumber": o.Sn(),
				}).Error("error-while-processing-end-software-download-request")
			}
		} else {
			if responsePkt, errResp = omcilib.CreateEndSoftwareDownloadResponse(msg.OmciPkt, msg.OmciMsg, me.ProcessingError); errResp == nil {
				if err := o.InternalState.Event(OnuTxFailImageDownload); err != nil {
					onuLogger.WithFields(log.Fields{
						"OnuId":  o.ID,
						"IntfId": o.PonPortID,
						"OnuSn":  o.Sn(),
						"Err":    err.Error(),
					}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageDownloadError)
				}
			}
		}

	case omci.ActivateSoftwareRequestType:
		if responsePkt, errResp = omcilib.CreateActivateSoftwareResponse(msg.OmciPkt, msg.OmciMsg); errResp == nil {
			o.MibDataSync++
			if err := o.InternalState.Event(OnuTxActivateImage); err != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
					"Err":    err.Error(),
				}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageActivated)
			}
			if msgObj, err := omcilib.ParseActivateSoftwareRequest(msg.OmciPkt); err == nil {
				o.ActiveImageEntityId = msgObj.EntityInstance
			} else {
				onuLogger.Errorf("something-went-wrong-while-activating: %s", err)
			}
			onuLogger.WithFields(log.Fields{
				"OnuId":                  o.ID,
				"IntfId":                 o.PonPortID,
				"OnuSn":                  o.Sn(),
				"ActiveImageEntityId":    o.ActiveImageEntityId,
				"CommittedImageEntityId": o.CommittedImageEntityId,
			}).Info("onu-software-image-activated")

			// powercycle the ONU
			// we run this in a separate goroutine so that
			// the ActivateSoftwareResponse is sent to VOLTHA
			// NOTE do we need to wait before rebooting?
			go func() {
				if err := o.Reboot(10 * time.Second); err != nil {
					log.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"SerialNumber": o.Sn(),
						"err":          err,
					}).Error("cannot-reboot-onu-after-omci-activate-software-request")
				}
			}()
		}
	case omci.CommitSoftwareRequestType:
		if responsePkt, errResp = omcilib.CreateCommitSoftwareResponse(msg.OmciPkt, msg.OmciMsg); errResp == nil {
			o.MibDataSync++
			if msgObj, err := omcilib.ParseCommitSoftwareRequest(msg.OmciPkt); err == nil {
				// TODO validate that the image to commit is:
				// - active
				// - not already committed
				o.CommittedImageEntityId = msgObj.EntityInstance
			} else {
				onuLogger.Errorf("something-went-wrong-while-committing: %s", err)
			}
			if err := o.InternalState.Event(OnuTxCommitImage); err != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
					"Err":    err.Error(),
				}).Errorf("cannot-change-onu-internal-state-to-%s", OnuStateImageCommitted)
			}
			onuLogger.WithFields(log.Fields{
				"OnuId":                  o.ID,
				"IntfId":                 o.PonPortID,
				"OnuSn":                  o.Sn(),
				"ActiveImageEntityId":    o.ActiveImageEntityId,
				"CommittedImageEntityId": o.CommittedImageEntityId,
			}).Info("onu-software-image-committed")
		}
	case omci.GetAllAlarmsRequestType:
		// Reset the alarm sequence number on receiving get all alarms request.
		o.onuAlarmsInfoLock.Lock()
		for key, alarmInfo := range o.onuAlarmsInfo {
			// reset the alarm sequence no
			alarmInfo.SequenceNo = 0
			o.onuAlarmsInfo[key] = alarmInfo
		}
		o.onuAlarmsInfoLock.Unlock()
		responsePkt, _ = omcilib.CreateGetAllAlarmsResponse(msg.OmciMsg.TransactionID, o.onuAlarmsInfo)
	case omci.GetAllAlarmsNextRequestType:
		if responsePkt, errResp = omcilib.CreateGetAllAlarmsNextResponse(msg.OmciPkt, msg.OmciMsg, o.onuAlarmsInfo); errResp != nil {
			responsePkt = nil //Do not send any response for error case
		}
	default:
		onuLogger.WithFields(log.Fields{
			"omciBytes":    hex.EncodeToString(msg.OmciPkt.Data()),
			"omciPkt":      msg.OmciPkt,
			"omciMsgType":  msg.OmciMsg.MessageType,
			"transCorrId":  msg.OmciMsg.TransactionID,
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
		}).Warnf("OMCI-message-not-supported")
	}

	if responsePkt != nil {
		if err := o.sendOmciIndication(responsePkt, msg.OmciMsg.TransactionID, stream); err != nil {
			onuLogger.WithFields(log.Fields{
				"IntfId":          o.PonPortID,
				"SerialNumber":    o.Sn(),
				"omciPacket":      responsePkt,
				"msg.OmciMsgType": msg.OmciMsg.MessageType,
				"transCorrId":     msg.OmciMsg.TransactionID,
			}).Errorf("failed-to-send-omci-message: %v", err)
		}
	}

	o.publishOmciEvent(msg)
}

// sendOmciIndication takes an OMCI packet and sends it up to VOLTHA
func (o *Onu) sendOmciIndication(responsePkt []byte, txId uint16, stream bbsim.Stream) error {
	indication := &openolt.Indication_OmciInd{
		OmciInd: &openolt.OmciIndication{
			IntfId: o.PonPortID,
			OnuId:  o.ID,
			Pkt:    responsePkt,
		},
	}
	if err := stream.Send(&openolt.Indication{Data: indication}); err != nil {
		return fmt.Errorf("failed-to-send-omci-message: %v", err)
	}
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
		"omciPacket":   indication.OmciInd.Pkt,
		"transCorrId":  txId,
	}).Trace("omci-message-sent")
	return nil
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

func (o *Onu) handleFlowAdd(msg bbsim.OnuFlowUpdateMessage) {
	onuLogger.WithFields(log.Fields{
		"AllocId":           msg.Flow.AllocId,
		"Cookie":            msg.Flow.Cookie,
		"DstPort":           msg.Flow.Classifier.DstPort,
		"FlowId":            msg.Flow.FlowId,
		"FlowType":          msg.Flow.FlowType,
		"GemportId":         msg.Flow.GemportId,
		"InnerVlan":         msg.Flow.Classifier.IVid,
		"IntfId":            msg.Flow.AccessIntfId,
		"IpProto":           msg.Flow.Classifier.IpProto,
		"OnuId":             msg.Flow.OnuId,
		"OnuSn":             o.Sn(),
		"OuterVlan":         msg.Flow.Classifier.OVid,
		"PortNo":            msg.Flow.PortNo,
		"SrcPort":           msg.Flow.Classifier.SrcPort,
		"UniID":             msg.Flow.UniId,
		"ClassifierEthType": fmt.Sprintf("%x", msg.Flow.Classifier.EthType),
		"ClassifierOPbits":  msg.Flow.Classifier.OPbits,
		"ClassifierIVid":    msg.Flow.Classifier.IVid,
		"ClassifierOVid":    msg.Flow.Classifier.OVid,
		"ReplicateFlow":     msg.Flow.ReplicateFlow,
		"PbitToGemport":     msg.Flow.PbitToGemport,
	}).Debug("OLT receives FlowAdd for ONU")

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

	var gemPortId uint32
	if msg.Flow.ReplicateFlow {
		// This means that the OLT should replicate the flow for each PBIT, for BBSim it's enough to use the
		// first available gemport (we only need to send one packet)
		// NOTE different TP may create different mapping between PBits and GemPorts, this may require some changes
		gemPortId = msg.Flow.PbitToGemport[0]
	} else {
		// if replicateFlows is false, then the flow is carrying the correct GemPortId
		gemPortId = uint32(msg.Flow.GemportId)
	}
	o.addGemPortToService(gemPortId, msg.Flow.Classifier.EthType, msg.Flow.Classifier.OVid, msg.Flow.Classifier.IVid)

	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		// NOTE storing the PortNO, it's needed when sending PacketIndications
		o.storePortNumber(uint32(msg.Flow.PortNo))

		for _, s := range o.Services {
			s.HandleAuth()
		}
	} else if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeIPv4) &&
		msg.Flow.Classifier.SrcPort == uint32(68) &&
		msg.Flow.Classifier.DstPort == uint32(67) {

		for _, s := range o.Services {
			s.HandleDhcp(uint8(msg.Flow.Classifier.OPbits), int(msg.Flow.Classifier.OVid))
		}
	}
}

func (o *Onu) handleFlowRemove(msg bbsim.OnuFlowUpdateMessage) {
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

		// check if ONU delete is performed and
		// terminate the ONU's ProcessOnuMessages Go routine
		if o.InternalState.Current() == OnuStateDisabled {
			close(o.Channel)
		}
	}
}

func (o *Onu) Reboot(timeout time.Duration) error {
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        o.ID,
		"SerialNumber": o.Sn(),
	}).Debug("shutting-down-onu")
	if err := o.HandleShutdownONU(); err != nil {
		return err
	}
	time.Sleep(timeout)
	onuLogger.WithFields(log.Fields{
		"IntfId":       o.PonPortID,
		"OnuId":        o.ID,
		"SerialNumber": o.Sn(),
	}).Debug("power-on-onu")
	if err := o.HandlePowerOnONU(); err != nil {
		return err
	}
	return nil
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

// handleOmciResponse is used in BBR to generate the OMCI packets the openolt-adapter would send to the device
func (o *Onu) handleOmciResponse(msg bbsim.OmciIndicationMessage, client openolt.OpenoltClient) {

	// we need to encode the packet in HEX
	pkt := make([]byte, len(msg.OmciInd.Pkt)*2)
	hex.Encode(pkt, msg.OmciInd.Pkt)
	packet, omciMsg, err := omcilib.ParseOpenOltOmciPacket(pkt)
	if err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
			"omciPacket":   msg.OmciInd.Pkt,
		}).Error("BBR Cannot parse OMCI packet")
	}

	log.WithFields(log.Fields{
		"IntfId":  msg.OmciInd.IntfId,
		"OnuId":   msg.OmciInd.OnuId,
		"OnuSn":   o.Sn(),
		"Pkt":     msg.OmciInd.Pkt,
		"msgType": omciMsg.MessageType,
	}).Trace("ONU Receives OMCI Msg")
	switch omciMsg.MessageType {
	default:
		log.WithFields(log.Fields{
			"IntfId":  msg.OmciInd.IntfId,
			"OnuId":   msg.OmciInd.OnuId,
			"OnuSn":   o.Sn(),
			"Pkt":     msg.OmciInd.Pkt,
			"msgType": omciMsg.MessageType,
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
			// start sending the flows, we don't care about the OMCI setup in BBR, just that a lot of messages can go through
			if err := o.InternalState.Event(BbrOnuTxSendEapolFlow); err != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Errorf("Error while transitioning ONU State %v", err)
			}
		} else {
			mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
			sendOmciMsg(mibUploadNext, o.PonPortID, o.ID, o.SerialNumber, "mibUploadNext", client)
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
		FlowId:        uint64(o.ID),
		FlowType:      "downstream",
		NetworkIntfId: int32(0),
		Classifier:    &classifierProto,
		Action:        &actionProto,
		Priority:      int32(100),
		Cookie:        uint64(o.ID),
		PortNo:        o.ID, // NOTE we are using this to map an incoming packetIndication to an ONU
		// AllocId and GemPorts need to be unique per PON
		// for now use the ONU-ID, will need to change once we support multiple UNIs
		AllocId:   int32(o.ID),
		GemportId: int32(o.ID),
	}

	if _, err := client.FlowAdd(context.Background(), &downstreamFlow); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
			"Err":          err,
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
		FlowId:        uint64(o.ID),
		FlowType:      "downstream",
		NetworkIntfId: int32(0),
		Classifier:    &classifierProto,
		Action:        &actionProto,
		Priority:      int32(100),
		Cookie:        uint64(o.ID),
		PortNo:        o.ID, // NOTE we are using this to map an incoming packetIndication to an ONU
		// AllocId and GemPorts need to be unique per PON
		// for now use the ONU-ID, will need to change once we support multiple UNIs
		AllocId:   int32(o.ID),
		GemportId: int32(o.ID),
	}

	if _, err := client.FlowAdd(context.Background(), &downstreamFlow); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"FlowId":       downstreamFlow.FlowId,
			"PortNo":       downstreamFlow.PortNo,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
			"Err":          err,
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
	if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
		log.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
			"OnuId":  onu.ID,
		}).Infof("Failed to transition ONU to %s state: %s", OnuStateInitialized, err.Error())
	}

	if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
		log.WithFields(log.Fields{
			"IntfId": onu.PonPortID,
			"OnuSn":  onu.Sn(),
			"OnuId":  onu.ID,
		}).Infof("Failed to transition ONU to %s state: %s", OnuStateDiscovered, err.Error())
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

func (o *Onu) SendOMCIAlarmNotificationMsg(raiseOMCIAlarm bool, alarmType string) {
	switch alarmType {
	case "ONU_ALARM_LOS":
		msg := bbsim.Message{
			Type: bbsim.UniStatusAlarm,
			Data: bbsim.UniStatusAlarmMessage{
				OnuSN:          o.SerialNumber,
				OnuID:          o.ID,
				EntityID:       257,
				RaiseOMCIAlarm: raiseOMCIAlarm,
			},
		}
		o.Channel <- msg
	}

}

func (o *Onu) IncrementAlarmSequenceNumber(key omcilib.OnuAlarmInfoMapKey) uint8 {
	o.onuAlarmsInfoLock.Lock()
	defer o.onuAlarmsInfoLock.Unlock()
	if alarmInfo, ok := o.onuAlarmsInfo[key]; ok {
		if alarmInfo.SequenceNo == 255 {
			alarmInfo.SequenceNo = 1
		} else {
			alarmInfo.SequenceNo++
		}
		o.onuAlarmsInfo[key] = alarmInfo
		return alarmInfo.SequenceNo
	} else {
		// This is the first time alarm notification message is being sent
		o.onuAlarmsInfo[key] = omcilib.OnuAlarmInfo{
			SequenceNo: 1,
		}
		return 1
	}
}
