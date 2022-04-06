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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"

	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"

	pb "github.com/opencord/bbsim/api/bbsim"
	"github.com/opencord/bbsim/internal/bbsim/alarmsim"

	"net"
	"strconv"
	"time"

	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	me "github.com/opencord/omci-lib-go/v2/generated"

	"github.com/boguslaw-wojcik/crc32a"
	"github.com/google/gopacket/layers"
	"github.com/jpillora/backoff"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/omci-lib-go/v2"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/opencord/voltha-protos/v5/go/tech_profile"
	log "github.com/sirupsen/logrus"
)

var onuLogger = log.WithFields(log.Fields{
	"module": "ONU",
})

const (
	maxOmciMsgCounter = 10
)

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
	ID uint64
}

type Onu struct {
	ID                  uint32
	PonPortID           uint32
	PonPort             *PonPort
	InternalState       *fsm.FSM
	DiscoveryRetryDelay time.Duration // this is the time between subsequent Discovery Indication
	DiscoveryDelay      time.Duration // this is the time to send the first Discovery Indication

	Backoff *backoff.Backoff
	// ONU State
	UniPorts  []UniPortIf
	PotsPorts []PotsPortIf
	Flows     []FlowKey
	FlowIds   []uint64 // keep track of the flows we currently have in the ONU

	OperState    *fsm.FSM
	SerialNumber *openolt.SerialNumber

	AdminLockState uint8 // 0 is enabled, 1 is disabled.

	Channel chan bbsim.Message // this Channel is to track state changes OMCI messages, EAPOL and DHCP packets

	// OMCI params
	MibDataSync                   uint8
	ImageSoftwareExpectedSections int
	ImageSoftwareReceivedSections int
	ActiveImageEntityId           uint16
	CommittedImageEntityId        uint16
	StandbyImageVersion           string
	ActiveImageVersion            string
	InDownloadImageVersion        string
	CommittedImageVersion         string
	OmciResponseRate              uint8
	OmciMsgCounter                uint8
	ImageSectionData              []byte

	// OMCI params (Used in BBR)
	tid       uint16
	hpTid     uint16
	seqNumber uint16
	MibDb     *omcilib.MibDb

	DoneChannel       chan bool // this channel is used to signal once the onu is complete (when the struct is used by BBR)
	TrafficSchedulers *tech_profile.TrafficSchedulers
	onuAlarmsInfoLock sync.RWMutex
	onuAlarmsInfo     map[omcilib.OnuAlarmInfoMapKey]omcilib.OnuAlarmInfo
}

func (o *Onu) Sn() string {
	return common.OnuSnToString(o.SerialNumber)
}

func CreateONU(olt *OltDevice, pon *PonPort, id uint32, delay time.Duration, nextCtag map[string]int, nextStag map[string]int, isMock bool) *Onu {

	o := Onu{
		ID:                            id,
		PonPortID:                     pon.ID,
		PonPort:                       pon,
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
		//TODO this needs reworking, it's always 0 or 1, possibly base all on the version
		ActiveImageEntityId:    0, // when we start the SoftwareImage with ID 0 is active and committed
		CommittedImageEntityId: 0,
		StandbyImageVersion:    "BBSM_IMG_00000",
		ActiveImageVersion:     "BBSM_IMG_00001",
		CommittedImageVersion:  "BBSM_IMG_00001",
		OmciResponseRate:       olt.OmciResponseRate,
		OmciMsgCounter:         0,
	}
	o.SerialNumber = NewSN(olt.ID, pon.ID, id)
	// NOTE this state machine is used to track the operational
	// state as requested by VOLTHA
	o.OperState = getOperStateFSM(func(e *fsm.Event) {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
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
			{Name: OnuTxPonDisable, Src: []string{OnuStateEnabled, OnuStateImageActivated, OnuStateImageDownloadError, OnuStateImageCommitted, OnuStateImageDownloadComplete}, Dst: OnuStatePonDisabled},
			// Software Image Download related states
			{Name: OnuTxStartImageDownload, Src: []string{OnuStateEnabled, OnuStateImageDownloadComplete, OnuStateImageDownloadError, OnuStateImageCommitted}, Dst: OnuStateImageDownloadStarted},
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

				// disable the UNI ports
				for _, uni := range o.UniPorts {
					_ = uni.Disable()
				}

				// disable the POTS UNI ports
				for _, pots := range o.PotsPorts {
					_ = pots.Disable()
				}

				// verify all the flows removes are handled and
				// terminate the ONU's ProcessOnuMessages Go routine
				// NOTE may need to wait for the UNIs to be down too before shutting down the channel
				if len(o.FlowIds) == 0 {
					close(o.Channel)
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
	onuLogger.WithFields(log.Fields{
		"OnuId":   o.ID,
		"IntfId":  o.PonPortID,
		"OnuSn":   o.Sn(),
		"NumUni":  olt.NumUni,
		"NumPots": olt.NumPots,
	}).Debug("creating-uni-ports")

	// create Ethernet UNIs
	for i := 0; i < olt.NumUni; i++ {
		uni, err := NewUniPort(uint32(i), &o, nextCtag, nextStag)
		if err != nil {
			onuLogger.WithFields(log.Fields{
				"OnuId":  o.ID,
				"IntfId": o.PonPortID,
				"OnuSn":  o.Sn(),
				"Err":    err,
			}).Fatal("cannot-create-uni-port")
		}
		o.UniPorts = append(o.UniPorts, uni)
	}
	// create POTS UNIs, with progressive IDs
	for i := olt.NumUni; i < (olt.NumUni + olt.NumPots); i++ {
		pots, err := NewPotsPort(uint32(i), &o)
		if err != nil {
			onuLogger.WithFields(log.Fields{
				"OnuId":  o.ID,
				"IntfId": o.PonPortID,
				"OnuSn":  o.Sn(),
				"Err":    err,
			}).Fatal("cannot-create-pots-port")
		}
		o.PotsPorts = append(o.PotsPorts, pots)
	}

	mibDb, err := omcilib.GenerateMibDatabase(len(o.UniPorts), len(o.PotsPorts), o.PonPort.Technology)
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"OnuId":  o.ID,
			"IntfId": o.PonPortID,
			"OnuSn":  o.Sn(),
		}).Fatal("cannot-generate-mibdb-for-onu")
	}
	o.MibDb = mibDb

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
	o.Flows = []FlowKey{}
	o.PonPort.removeOnuId(o.ID)
	o.PonPort.removeAllocIdsForOnuSn(o.SerialNumber)
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

	defer onuLogger.WithFields(log.Fields{
		"onuID":  o.ID,
		"onuSN":  o.Sn(),
		"stream": stream,
	}).Debug("Stopped handling ONU Indication Channel")

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
				_ = o.handleOmciRequest(msg, stream)
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

				uni, err := o.findUniByPortNo(msg.PortNo)

				if err != nil {
					onuLogger.WithFields(log.Fields{
						"IntfId":     msg.IntfId,
						"OnuId":      msg.OnuId,
						"pktType":    msg.Type,
						"portNo":     msg.PortNo,
						"MacAddress": msg.MacAddress,
						"Pkt":        hex.EncodeToString(msg.Packet.Data()),
						"OnuSn":      o.Sn(),
					}).Error("Cannot find Uni associated with packet")
					return
				}
				uni.PacketCh <- msg
			// BBR specific messages
			case bbsim.OnuPacketIn:
				// NOTE we only receive BBR packets here.
				// Eapol.HandleNextPacket can handle both BBSim and BBr cases so the call is the same
				// in the DHCP case VOLTHA only act as a proxy, the behaviour is completely different thus we have a dhcp.HandleNextBbrPacket
				msg, _ := message.Data.(bbsim.OnuPacketMessage)

				onuLogger.WithFields(log.Fields{
					"IntfId":    msg.IntfId,
					"OnuId":     msg.OnuId,
					"PortNo":    msg.PortNo,
					"GemPortId": msg.GemPortId,
					"pktType":   msg.Type,
				}).Trace("Received OnuPacketIn Message")

				uni, err := o.findUniByPortNo(msg.PortNo)
				if err != nil {
					onuLogger.WithFields(log.Fields{
						"IntfId":    msg.IntfId,
						"OnuId":     msg.OnuId,
						"PortNo":    msg.PortNo,
						"GemPortId": msg.GemPortId,
						"pktType":   msg.Type,
					}).Error(err.Error())
				}

				// BBR has one service and one UNI
				serviceId := uint32(0)
				oltId := 0
				if msg.Type == packetHandlers.EAPOL {
					eapol.HandleNextPacket(msg.OnuId, msg.IntfId, msg.GemPortId, o.Sn(), msg.PortNo, uni.ID, serviceId, oltId, o.InternalState, msg.Packet, stream, client)
				} else if msg.Type == packetHandlers.DHCP {
					_ = dhcp.HandleNextBbrPacket(o.ID, o.PonPortID, o.Sn(), o.DoneChannel, msg.Packet, client)
				}
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

// handleOmciRequest is responsible to parse the OMCI packets received from the openolt adapter
// and generate the appropriate response to it
func (o *Onu) handleOmciRequest(msg bbsim.OmciMessage, stream openolt.Openolt_EnableIndicationServer) error {

	onuLogger.WithFields(log.Fields{
		"omciMsgType":  msg.OmciMsg.MessageType,
		"transCorrId":  strconv.FormatInt(int64(msg.OmciMsg.TransactionID), 16),
		"DeviceIdent":  msg.OmciMsg.DeviceIdentifier,
		"IntfId":       o.PonPortID,
		"SerialNumber": o.Sn(),
	}).Trace("omci-message-decoded")

	if o.OmciMsgCounter < maxOmciMsgCounter {
		o.OmciMsgCounter++
	} else {
		o.OmciMsgCounter = 1
	}
	if o.OmciMsgCounter > o.OmciResponseRate {
		onuLogger.WithFields(log.Fields{
			"OmciMsgCounter":   o.OmciMsgCounter,
			"OmciResponseRate": o.OmciResponseRate,
			"omciMsgType":      msg.OmciMsg.MessageType,
			"txId":             msg.OmciMsg.TransactionID,
		}).Debug("skipping-omci-msg-response")
		return fmt.Errorf("skipping-omci-msg-response-because-of-response-rate-%d", o.OmciResponseRate)
	}
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
			o.PonPort.removeAllocIdsForOnuSn(o.SerialNumber)
			o.PonPort.removeGemPortBySn(o.SerialNumber)
		}
	case omci.MibUploadRequestType:
		responsePkt, _ = omcilib.CreateMibUploadResponse(msg.OmciMsg.TransactionID, o.MibDb.NumberOfCommands)
	case omci.MibUploadNextRequestType:
		responsePkt, _ = omcilib.CreateMibUploadNextResponse(msg.OmciPkt, msg.OmciMsg, o.MibDataSync, o.MibDb)
	case omci.GetRequestType:
		onuDown := o.AdminLockState == 1
		responsePkt, _ = omcilib.CreateGetResponse(msg.OmciPkt, msg.OmciMsg, o.SerialNumber, o.MibDataSync, o.ActiveImageEntityId,
			o.CommittedImageEntityId, o.StandbyImageVersion, o.ActiveImageVersion, o.CommittedImageVersion, onuDown)

	case omci.SetRequestType:
		success := true
		msgObj, _ := omcilib.ParseSetRequest(msg.OmciPkt)
		switch msgObj.EntityClass {
		case me.PhysicalPathTerminationPointEthernetUniClassID:
			// if we're Setting a PPTP state
			// we need to send the appropriate alarm (handled in the UNI struct)
			uni, err := o.FindUniByEntityId(msgObj.EntityInstance)
			if err != nil {
				onuLogger.Error(err)
				success = false
			} else {
				// 1 locks the UNI, 0 unlocks it
				adminState := msgObj.Attributes[me.PhysicalPathTerminationPointEthernetUni_AdministrativeState].(uint8)
				var err error
				if adminState == 1 {
					err = uni.Disable()
				} else {
					err = uni.Enable()
				}
				if err != nil {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"UniMeId":      uni.MeId,
						"UniId":        uni.ID,
						"SerialNumber": o.Sn(),
						"Err":          err.Error(),
					}).Warn("cannot-change-uni-status")
				}
			}
		case me.PhysicalPathTerminationPointPotsUniClassID:
			// if we're Setting a PPTP state
			// we need to send the appropriate alarm (handled in the POTS struct)
			pots, err := o.FindPotsByEntityId(msgObj.EntityInstance)
			if err != nil {
				onuLogger.Error(err)
				success = false
			} else {
				// 1 locks the UNI, 0 unlocks it
				adminState := msgObj.Attributes[me.PhysicalPathTerminationPointPotsUni_AdministrativeState].(uint8)
				var err error
				if adminState == 1 {
					err = pots.Disable()
				} else {
					err = pots.Enable()
				}
				if err != nil {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"PotsMeId":     pots.MeId,
						"PotsId":       pots.ID,
						"SerialNumber": o.Sn(),
						"Err":          err.Error(),
					}).Warn("cannot-change-pots-status")
				}
			}
		case me.OnuGClassID:
			o.AdminLockState = msgObj.Attributes[me.OnuG_AdministrativeState].(uint8)
			onuLogger.WithFields(log.Fields{
				"IntfId":         o.PonPortID,
				"OnuId":          o.ID,
				"SerialNumber":   o.Sn(),
				"AdminLockState": o.AdminLockState,
			}).Debug("set-onu-admin-lock-state")
		case me.TContClassID:
			allocId := msgObj.Attributes[me.TCont_AllocId].(uint16)
			entityID := msgObj.Attributes["ManagedEntityId"].(uint16)

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
				o.PonPort.removeAllocId(o.PonPortID, o.ID, entityID)
			} else {
				if used, allocObj := o.PonPort.isAllocIdAllocated(o.PonPortID, o.ID, entityID); used {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"AllocId":      allocId,
						"SerialNumber": o.Sn(),
					}).Errorf("allocid-already-allocated-to-onu-with-sn-%s", common.OnuSnToString(allocObj.OnuSn))
					success = false
				} else {
					onuLogger.WithFields(log.Fields{
						"IntfId":       o.PonPortID,
						"OnuId":        o.ID,
						"TContId":      msgObj.EntityInstance,
						"AllocId":      allocId,
						"SerialNumber": o.Sn(),
					}).Trace("storing-alloc-id-via-omci")
					o.PonPort.storeAllocId(o.PonPortID, o.ID, entityID, allocId, o.SerialNumber)
				}
			}
		case me.EthernetFrameExtendedPmClassID,
			me.EthernetFrameExtendedPm64BitClassID:
			onuLogger.WithFields(log.Fields{
				"me-instance": msgObj.EntityInstance,
			}).Debug("set-request-received")
			// No need to reset counters as onu adapter will simply send the set control block request to actually reset
			// the counters, and respond with 0's without sending the get request to device.
			// Also, if we even reset the counters here in cache, then on get we need to restore the counters back which
			// would be of no use as ultimately the counters need to be restored.
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
		var classID me.ClassID
		var omciResult me.Results
		var instID uint16
		responsePkt, errResp, classID, instID, omciResult = omcilib.CreateTestResponse(msg.OmciPkt, msg.OmciMsg)
		// Send TestResult only in case the TestResponse omci result code is me.Success
		if responsePkt != nil && errResp == nil && omciResult == me.Success {
			if testResultPkt, err := omcilib.CreateTestResult(classID, instID, msg.OmciMsg.TransactionID); err == nil {
				// send test results asynchronously
				go func() {
					// Send test results after a second to emulate async behavior
					time.Sleep(1 * time.Second)
					if testResultPkt != nil {
						if err := o.sendOmciIndication(testResultPkt, msg.OmciMsg.TransactionID, stream); err != nil {
							onuLogger.WithFields(log.Fields{
								"IntfId":          o.PonPortID,
								"SerialNumber":    o.Sn(),
								"omciPacket":      testResultPkt,
								"msg.OmciMsgType": msg.OmciMsg.MessageType,
								"transCorrId":     msg.OmciMsg.TransactionID,
							}).Errorf("failed-to-send-omci-message: %v", err)
						}
					}
				}()
			}
		}
	case omci.SynchronizeTimeRequestType:
		// MDS counter increment is not required for this message type
		responsePkt, _ = omcilib.CreateSyncTimeResponse(msg.OmciPkt, msg.OmciMsg)
	case omci.StartSoftwareDownloadRequestType:

		o.ImageSoftwareReceivedSections = 0
		o.ImageSectionData = []byte{}
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
			//Extracting the first 14 bytes to use as a version for this image.
			if o.ImageSoftwareReceivedSections == 0 {
				o.InDownloadImageVersion = string(msgObj.SectionData[0:14])
			}
			o.ImageSectionData = append(o.ImageSectionData, msgObj.SectionData...)
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
		if msgObj, err := omcilib.ParseDownloadSectionRequest(msg.OmciPkt); err == nil {
			onuLogger.WithFields(log.Fields{
				"OmciMsgType":    msg.OmciMsg.MessageType,
				"TransCorrId":    msg.OmciMsg.TransactionID,
				"EntityInstance": msgObj.EntityInstance,
				"SectionNumber":  msgObj.SectionNumber,
				"SectionData":    msgObj.SectionData,
			}).Trace("received-download-section-request-with-response-type")
			o.ImageSectionData = append(o.ImageSectionData, msgObj.SectionData...)
			responsePkt, errResp = omcilib.CreateDownloadSectionResponse(msg.OmciPkt, msg.OmciMsg)

			if errResp != nil {
				onuLogger.WithFields(log.Fields{
					"OmciMsgType":  msg.OmciMsg.MessageType,
					"TransCorrId":  msg.OmciMsg.TransactionID,
					"Err":          errResp.Error(),
					"IntfId":       o.PonPortID,
					"SerialNumber": o.Sn(),
				}).Error("error-while-processing-create-download-section-response")
				return fmt.Errorf("error-while-processing-create-download-section-response: %s", errResp.Error())
			}
			o.ImageSoftwareReceivedSections++
		}
	case omci.EndSoftwareDownloadRequestType:
		success := o.handleEndSoftwareDownloadRequest(msg)

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
				}).Error("error-while-responding-to-end-software-download-request")
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
				previousActiveImage := o.ActiveImageVersion
				o.ActiveImageVersion = o.StandbyImageVersion
				o.StandbyImageVersion = previousActiveImage
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
				o.ActiveImageEntityId = msgObj.EntityInstance
				o.CommittedImageEntityId = msgObj.EntityInstance
				//committed becomes standby
				o.StandbyImageVersion = o.CommittedImageVersion
				o.CommittedImageVersion = o.ActiveImageVersion
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
	return nil
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
		"omciPacket":   hex.EncodeToString(indication.OmciInd.Pkt),
		"transCorrId":  txId,
	}).Trace("omci-message-sent")
	return nil
}

// FindUniById retrieves a UNI by ID
func (o *Onu) FindUniById(uniID uint32) (*UniPort, error) {
	for _, u := range o.UniPorts {
		uni := u.(*UniPort)
		if uni.ID == uniID {
			return uni, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-uni-with-id-%d-on-onu-%s", uniID, o.Sn())
}

// FindPotsById retrieves a POTS port by ID
func (o *Onu) FindPotsById(uniID uint32) (*PotsPort, error) {
	for _, p := range o.PotsPorts {
		pots := p.(*PotsPort)
		if pots.ID == uniID {
			return pots, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-pots-with-id-%d-on-onu-%s", uniID, o.Sn())
}

// FindUniByEntityId retrieves a uni by MeID (the OMCI entity ID)
func (o *Onu) FindUniByEntityId(meId uint16) (*UniPort, error) {
	entityId := omcilib.EntityID{}.FromUint16(meId)
	for _, u := range o.UniPorts {
		uni := u.(*UniPort)
		if uni.MeId.Equals(entityId) {
			return uni, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-uni-with-meid-%s-on-onu-%s", entityId.ToString(), o.Sn())
}

// FindPotsByEntityId retrieves a POTS uni by MeID (the OMCI entity ID)
func (o *Onu) FindPotsByEntityId(meId uint16) (*PotsPort, error) {
	entityId := omcilib.EntityID{}.FromUint16(meId)
	for _, p := range o.PotsPorts {
		pots := p.(*PotsPort)
		if pots.MeId.Equals(entityId) {
			return pots, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-pots-with-meid-%s-on-onu-%s", entityId.ToString(), o.Sn())
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

	uni, err := o.FindUniById(uint32(msg.Flow.UniId))
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"UniId":        msg.Flow.UniId,
			"PortNo":       msg.Flow.PortNo,
			"SerialNumber": o.Sn(),
			"FlowId":       msg.Flow.FlowId,
			"FlowType":     msg.Flow.FlowType,
		}).Error("cannot-find-uni-port-for-flow")
	}

	uni.addGemPortToService(gemPortId, msg.Flow.Classifier.EthType, msg.Flow.Classifier.OVid, msg.Flow.Classifier.IVid)
	uni.StorePortNo(msg.Flow.PortNo)

	if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeEAPOL) && msg.Flow.Classifier.OVid == 4091 {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"UniId":        msg.Flow.UniId,
			"PortNo":       msg.Flow.PortNo,
			"SerialNumber": o.Sn(),
			"FlowId":       msg.Flow.FlowId,
		}).Debug("EAPOL flow detected")
		uni.HandleAuth()
	} else if msg.Flow.Classifier.EthType == uint32(layers.EthernetTypeIPv4) &&
		msg.Flow.Classifier.SrcPort == uint32(68) &&
		msg.Flow.Classifier.DstPort == uint32(67) {
		onuLogger.WithFields(log.Fields{
			"IntfId":       o.PonPortID,
			"OnuId":        o.ID,
			"UniId":        msg.Flow.UniId,
			"PortNo":       msg.Flow.PortNo,
			"SerialNumber": o.Sn(),
			"FlowId":       msg.Flow.FlowId,
			"FlowType":     msg.Flow.FlowType,
		}).Debug("DHCP flow detected")
		uni.HandleDhcp(uint8(msg.Flow.Classifier.OPbits), int(msg.Flow.Classifier.OVid))
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

// returns true if the request is successful, false otherwise
func (o *Onu) handleEndSoftwareDownloadRequest(msg bbsim.OmciMessage) bool {
	msgObj, err := omcilib.ParseEndSoftwareDownloadRequest(msg.OmciPkt)
	if err != nil {
		onuLogger.WithFields(log.Fields{
			"OmciMsgType":  msg.OmciMsg.MessageType,
			"TransCorrId":  msg.OmciMsg.TransactionID,
			"Err":          err.Error(),
			"IntfId":       o.PonPortID,
			"SerialNumber": o.Sn(),
		}).Error("error-while-processing-end-software-download-request")
		return false
	}

	onuLogger.WithFields(log.Fields{
		"OnuId":  o.ID,
		"IntfId": o.PonPortID,
		"OnuSn":  o.Sn(),
		"msgObj": msgObj,
	}).Trace("EndSoftwareDownloadRequest received message")

	// if the image download is ongoing and we receive a message with
	// ImageSize = 0 and Crc = 4294967295 (0xFFFFFFFF) respond with success
	if o.ImageSoftwareReceivedSections > 0 &&
		msgObj.ImageSize == 0 &&
		msgObj.CRC32 == 4294967295 {
		o.ImageSoftwareReceivedSections = 0
		// NOTE potentially we may want to add a ONU state to reflect
		// the software download abort
		return true
	}

	// In the startSoftwareDownload we get the image size and the window size.
	// We calculate how many DownloadSection we should receive and validate
	// that we got the correct amount when we receive this message
	// If the received sections are different from the expected sections
	// respond with failure
	if o.ImageSoftwareExpectedSections != o.ImageSoftwareReceivedSections {
		onuLogger.WithFields(log.Fields{
			"OnuId":            o.ID,
			"IntfId":           o.PonPortID,
			"OnuSn":            o.Sn(),
			"ExpectedSections": o.ImageSoftwareExpectedSections,
			"ReceivedSections": o.ImageSoftwareReceivedSections,
		}).Errorf("onu-did-not-receive-all-image-sections")
		return false
	}

	// check the received CRC vs the computed CRC
	computedCRC := crc32a.Checksum(o.ImageSectionData[:int(msgObj.ImageSize)])
	//Convert the crc to network byte order
	var byteSlice = make([]byte, 4)
	binary.LittleEndian.PutUint32(byteSlice, computedCRC)
	computedCRC = binary.BigEndian.Uint32(byteSlice)
	if msgObj.CRC32 != computedCRC {
		onuLogger.WithFields(log.Fields{
			"OnuId":         o.ID,
			"IntfId":        o.PonPortID,
			"OnuSn":         o.Sn(),
			"ReceivedCRC":   msgObj.CRC32,
			"CalculatedCRC": computedCRC,
		}).Errorf("onu-image-crc-validation-failed")
		return false
	}

	o.StandbyImageVersion = o.InDownloadImageVersion
	onuLogger.WithFields(log.Fields{
		"OnuId":          o.ID,
		"IntfId":         o.PonPortID,
		"OnuSn":          o.Sn(),
		"StandbyVersion": o.StandbyImageVersion,
	}).Debug("onu-image-version-updated")
	return true
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
// StartOmci is called in BBR to start the OMCI state machine
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
		// once the mibUpload is complete send a SetRequest for the PPTP to enable the UNI
		// NOTE that in BBR we only enable the first UNI
		if o.seqNumber == o.MibDb.NumberOfCommands {
			meId := omcilib.GenerateUniPortEntityId(1)

			meParams := me.ParamData{
				EntityID:   meId.ToUint16(),
				Attributes: me.AttributeValueMap{me.PhysicalPathTerminationPointEthernetUni_AdministrativeState: 0},
			}
			managedEntity, omciError := me.NewPhysicalPathTerminationPointEthernetUni(meParams)
			if omciError.GetError() != nil {
				onuLogger.WithFields(log.Fields{
					"OnuId":  o.ID,
					"IntfId": o.PonPortID,
					"OnuSn":  o.Sn(),
				}).Fatal(omciError.GetError())
			}

			setPPtp, _ := omcilib.CreateSetRequest(managedEntity, 1)
			sendOmciMsg(setPPtp, o.PonPortID, o.ID, o.SerialNumber, "setRquest", client)
		} else {
			mibUploadNext, _ := omcilib.CreateMibUploadNextRequest(o.getNextTid(false), o.seqNumber)
			sendOmciMsg(mibUploadNext, o.PonPortID, o.ID, o.SerialNumber, "mibUploadNext", client)
		}
	case omci.SetResponseType:
		// once we set the PPTP to active we can start sending flows

		if err := o.InternalState.Event(BbrOnuTxSendEapolFlow); err != nil {
			onuLogger.WithFields(log.Fields{
				"OnuId":  o.ID,
				"IntfId": o.PonPortID,
				"OnuSn":  o.Sn(),
			}).Errorf("Error while transitioning ONU State %v", err)
		}
	case omci.AlarmNotificationType:
		log.Info("bbr-received-alarm")
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
		FlowType:      flowTypeDownstream,
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

	// BBR only works with a single UNI and a single service (ATT HSIA)
	hsia := o.UniPorts[0].(*UniPort).Services[0].(*Service)
	classifierProto := openolt.Classifier{
		EthType: uint32(layers.EthernetTypeIPv4),
		SrcPort: uint32(68),
		DstPort: uint32(67),
		OVid:    uint32(hsia.CTag),
		OPbits:  255,
	}

	actionProto := openolt.Action{}

	downstreamFlow := openolt.Flow{
		AccessIntfId:  int32(o.PonPortID),
		OnuId:         int32(o.ID),
		UniId:         int32(0), // BBR only supports a single UNI
		FlowId:        uint64(o.ID),
		FlowType:      flowTypeDownstream,
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

// deprecated, delegate this to the uniPort
func (onu *Onu) findServiceByMacAddress(macAddress net.HardwareAddr) (*Service, error) {
	// FIXME is there a better way to avoid this loop?
	for _, u := range onu.UniPorts {
		uni := u.(*UniPort)
		for _, s := range uni.Services {
			service := s.(*Service)
			if service.HwAddress.String() == macAddress.String() {
				return service, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot-find-service-with-mac-address-%s", macAddress.String())
}

func (onu *Onu) findUniByPortNo(portNo uint32) (*UniPort, error) {
	for _, u := range onu.UniPorts {
		uni := u.(*UniPort)
		if uni.PortNo == portNo {
			return uni, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-uni-with-port-no-%d", portNo)
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

func (o *Onu) InvalidateMibDataSync() {
	rand.Seed(time.Now().UnixNano())
	r := uint8(rand.Intn(10) + 1)

	o.MibDataSync += r

	// Since MibDataSync is a uint8, summing to it will never
	// result in a value higher than 255, but could be 0
	if o.MibDataSync == 0 {
		o.MibDataSync++
	}
}
