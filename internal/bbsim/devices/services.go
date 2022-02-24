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
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	"github.com/opencord/bbsim/internal/bbsim/responders/igmp"
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/bbsim/internal/common"
	log "github.com/sirupsen/logrus"
)

var serviceLogger = log.WithFields(log.Fields{
	"module": "SERVICE",
})

// time to wait before fail EAPOL/DHCP
// (it's a variable and not a constant so it can be overridden in the tests)
var eapolWaitTime = 60 * time.Second
var dhcpWaitTime = 60 * time.Second

const (
	ServiceStateCreated     = "created"
	ServiceStateInitialized = "initialized"
	ServiceStateDisabled    = "disabled"

	ServiceTxInitialize = "initialize"
	ServiceTxDisable    = "disable"
)

type ServiceIf interface {
	HandlePackets()                  // start listening on the PacketCh
	HandleAuth()                     // Sends the EapoStart packet
	HandleDhcp(pbit uint8, cTag int) // Sends the DHCPDiscover packet

	Initialize(stream bbsimTypes.Stream)
	UpdateStream(stream bbsimTypes.Stream)
	Disable()
}

type Service struct {
	Id                  uint32
	Name                string
	HwAddress           net.HardwareAddr
	UniPort             *UniPort
	CTag                int
	STag                int
	NeedsEapol          bool
	NeedsDhcp           bool
	NeedsIgmp           bool
	NeedsPPPoE          bool
	TechnologyProfileID int
	UniTagMatch         int
	ConfigureMacAddress bool
	EnableMacLearning   bool
	UsPonCTagPriority   uint8
	UsPonSTagPriority   uint8
	DsPonCTagPriority   uint8
	DsPonSTagPriority   uint8

	// state
	GemPort       uint32
	InternalState *fsm.FSM
	EapolState    *fsm.FSM
	DHCPState     *fsm.FSM
	IGMPState     *fsm.FSM
	Channel       chan bbsimTypes.Message          // drive Service lifecycle
	PacketCh      chan bbsimTypes.OnuPacketMessage // handle packets
	Stream        bbsimTypes.Stream                // the gRPC stream to communicate with the adapter, created in the initialize transition
}

func NewService(id uint32, name string, hwAddress net.HardwareAddr, uni *UniPort, cTag int, sTag int,
	needsEapol bool, needsDchp bool, needsIgmp bool, needsPPPoE bool, tpID int, uniTagMatch int, configMacAddress bool,
	enableMacLearning bool, usPonCTagPriority uint8, usPonSTagPriority uint8, dsPonCTagPriority uint8,
	dsPonSTagPriority uint8) (*Service, error) {

	service := Service{
		Id:                  id,
		Name:                name,
		HwAddress:           hwAddress,
		UniPort:             uni,
		CTag:                cTag,
		STag:                sTag,
		NeedsEapol:          needsEapol,
		NeedsDhcp:           needsDchp,
		NeedsIgmp:           needsIgmp,
		NeedsPPPoE:          needsPPPoE,
		TechnologyProfileID: tpID,
		UniTagMatch:         uniTagMatch,
		ConfigureMacAddress: configMacAddress,
		EnableMacLearning:   enableMacLearning,
		UsPonCTagPriority:   usPonCTagPriority,
		UsPonSTagPriority:   usPonSTagPriority,
		DsPonCTagPriority:   dsPonCTagPriority,
		DsPonSTagPriority:   dsPonSTagPriority,
	}

	service.InternalState = fsm.NewFSM(
		ServiceStateCreated,
		fsm.Events{
			{Name: ServiceTxInitialize, Src: []string{ServiceStateCreated, ServiceStateDisabled}, Dst: ServiceStateInitialized},
			{Name: ServiceTxDisable, Src: []string{ServiceStateInitialized}, Dst: ServiceStateDisabled},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("InternalState", e.Src, e.Dst)
			},
			fmt.Sprintf("enter_%s", ServiceStateInitialized): func(e *fsm.Event) {

				stream, ok := e.Args[0].(bbsimTypes.Stream)
				if !ok {
					serviceLogger.Fatal("initialize invoke with wrong arguments")
				}

				service.UpdateStream(stream)

				service.PacketCh = make(chan bbsimTypes.OnuPacketMessage)
				service.Channel = make(chan bbsimTypes.Message)

				go service.HandlePackets()
				go service.HandleChannel()
			},
			fmt.Sprintf("enter_%s", ServiceStateDisabled): func(e *fsm.Event) {
				// reset the state machines
				service.EapolState.SetState(eapol.StateCreated)
				service.DHCPState.SetState("created")

				// stop listening for packets
				close(service.PacketCh)
				close(service.Channel)

				service.PacketCh = nil
				service.Channel = nil
			},
		},
	)

	service.EapolState = fsm.NewFSM(
		eapol.StateCreated,
		fsm.Events{
			{Name: eapol.EventStartAuth, Src: []string{eapol.StateCreated, eapol.StateResponseSuccessReceived, eapol.StateAuthFailed}, Dst: eapol.StateAuthStarted},
			{Name: eapol.EventStartSent, Src: []string{eapol.StateAuthStarted}, Dst: eapol.StateStartSent},
			{Name: eapol.EventResponseIdentitySent, Src: []string{eapol.StateStartSent}, Dst: eapol.StateResponseIdentitySent},
			{Name: eapol.EventResponseChallengeSent, Src: []string{eapol.StateResponseIdentitySent}, Dst: eapol.StateResponseChallengeSent},
			{Name: eapol.EventResponseSuccessReceived, Src: []string{eapol.StateResponseChallengeSent}, Dst: eapol.StateResponseSuccessReceived},
			{Name: eapol.EventAuthFailed, Src: []string{eapol.StateAuthStarted, eapol.StateStartSent, eapol.StateResponseIdentitySent, eapol.StateResponseChallengeSent}, Dst: eapol.StateAuthFailed},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("EapolState", e.Src, e.Dst)
			},
			fmt.Sprintf("before_%s", eapol.EventStartAuth): func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.StartEAPOL,
				}
				service.Channel <- msg
			},
			fmt.Sprintf("enter_%s", eapol.StateAuthStarted): func(e *fsm.Event) {
				go func() {

					for {
						select {
						case <-service.UniPort.Onu.PonPort.Olt.enableContext.Done():
							serviceLogger.WithFields(log.Fields{
								"context": service.UniPort.Onu.PonPort.Olt.enableContext,
							}).Debug("EAPOL cancelled, OLT is disabled")
							return
						case <-time.After(eapolWaitTime):
							if service.EapolState.Current() != eapol.StateResponseSuccessReceived {
								serviceLogger.WithFields(log.Fields{
									"OnuId":      service.UniPort.Onu.ID,
									"IntfId":     service.UniPort.Onu.PonPortID,
									"OnuSn":      service.UniPort.Onu.Sn(),
									"Name":       service.Name,
									"PortNo":     service.UniPort.PortNo,
									"UniId":      service.UniPort.ID,
									"EapolState": service.EapolState.Current(),
								}).Warn("EAPOL failed, resetting EAPOL State")

								_ = service.EapolState.Event(eapol.EventAuthFailed)
								if common.Config.BBSim.AuthRetry {
									_ = service.EapolState.Event(eapol.EventStartAuth)
								}

								return
							}
						}

					}
				}()
			},
		},
	)

	service.DHCPState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"created", "dhcp_ack_received", "dhcp_failed"}, Dst: "dhcp_started"},
			{Name: "dhcp_discovery_sent", Src: []string{"dhcp_started"}, Dst: "dhcp_discovery_sent"},
			{Name: "dhcp_request_sent", Src: []string{"dhcp_discovery_sent"}, Dst: "dhcp_request_sent"},
			{Name: "dhcp_ack_received", Src: []string{"dhcp_request_sent"}, Dst: "dhcp_ack_received"},
			{Name: "dhcp_failed", Src: []string{"dhcp_started", "dhcp_discovery_sent", "dhcp_request_sent"}, Dst: "dhcp_failed"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("DHCPState", e.Src, e.Dst)
			},
			"before_start_dhcp": func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.StartDHCP,
				}
				service.Channel <- msg
			},
			"enter_dhcp_started": func(e *fsm.Event) {
				go func() {

					for {
						select {
						case <-service.UniPort.Onu.PonPort.Olt.enableContext.Done():
							serviceLogger.WithFields(log.Fields{
								"context": service.UniPort.Onu.PonPort.Olt.enableContext,
							}).Debug("DHCP cancelled, OLT is disabled")
							return
						case <-time.After(dhcpWaitTime):
							if service.DHCPState.Current() != "dhcp_ack_received" {
								serviceLogger.WithFields(log.Fields{
									"OnuId":     service.UniPort.Onu.ID,
									"IntfId":    service.UniPort.Onu.PonPortID,
									"OnuSn":     service.UniPort.Onu.Sn(),
									"Name":      service.Name,
									"PortNo":    service.UniPort.PortNo,
									"UniId":     service.UniPort.ID,
									"DHCPState": service.DHCPState.Current(),
								}).Warn("DHCP failed, resetting DHCP State")

								_ = service.DHCPState.Event("dhcp_failed")
								if common.Config.BBSim.DhcpRetry {
									_ = service.DHCPState.Event("start_dhcp")
								}

								return
							}
						}

					}
				}()
			},
		},
	)

	service.IGMPState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "igmp_join_start", Src: []string{"created", "igmp_left", "igmp_join_error", "igmp_join_started"}, Dst: "igmp_join_started"},
			{Name: "igmp_join_startv3", Src: []string{"igmp_left", "igmp_join_error", "igmp_join_started"}, Dst: "igmp_join_started"},
			{Name: "igmp_join_error", Src: []string{"igmp_join_started"}, Dst: "igmp_join_error"},
			{Name: "igmp_leave", Src: []string{"igmp_join_started"}, Dst: "igmp_left"},
		},
		fsm.Callbacks{
			"igmp_join_start": func(e *fsm.Event) {
				igmpInfo, _ := e.Args[0].(bbsimTypes.IgmpMessage)
				msg := bbsimTypes.Message{
					Type: bbsimTypes.IGMPMembershipReportV2,
					Data: bbsimTypes.IgmpMessage{
						GroupAddress: igmpInfo.GroupAddress,
					},
				}
				service.Channel <- msg
			},
			"igmp_leave": func(e *fsm.Event) {
				igmpInfo, _ := e.Args[0].(bbsimTypes.IgmpMessage)
				msg := bbsimTypes.Message{
					Type: bbsimTypes.IGMPLeaveGroup,
					Data: bbsimTypes.IgmpMessage{
						GroupAddress: igmpInfo.GroupAddress,
					},
				}
				service.Channel <- msg
			},
			"igmp_join_startv3": func(e *fsm.Event) {
				igmpInfo, _ := e.Args[0].(bbsimTypes.IgmpMessage)
				msg := bbsimTypes.Message{
					Type: bbsimTypes.IGMPMembershipReportV3,
					Data: bbsimTypes.IgmpMessage{
						GroupAddress: igmpInfo.GroupAddress,
					},
				}
				service.Channel <- msg
			},
		},
	)

	return &service, nil
}

func (s *Service) UpdateStream(stream bbsimTypes.Stream) {
	s.Stream = stream
}

// HandleAuth is used to start EAPOL for a particular Service when the corresponding flow is received
func (s *Service) HandleAuth() {

	if !s.NeedsEapol {
		serviceLogger.WithFields(log.Fields{
			"OnuId":      s.UniPort.Onu.ID,
			"IntfId":     s.UniPort.Onu.PonPortID,
			"OnuSn":      s.UniPort.Onu.Sn(),
			"PortNo":     s.UniPort.PortNo,
			"UniId":      s.UniPort.ID,
			"Name":       s.Name,
			"NeedsEapol": s.NeedsEapol,
		}).Debug("Won't start authentication as EAPOL is not required")
		return
	}

	serviceLogger.WithFields(log.Fields{
		"OnuId":      s.UniPort.Onu.ID,
		"IntfId":     s.UniPort.Onu.PonPortID,
		"OnuSn":      s.UniPort.Onu.Sn(),
		"PortNo":     s.UniPort.PortNo,
		"UniId":      s.UniPort.ID,
		"Name":       s.Name,
		"NeedsEapol": s.NeedsEapol,
	}).Debug("Starting EAPOL for the service")
	if err := s.EapolState.Event(eapol.EventStartAuth); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.UniPort.Onu.ID,
			"IntfId": s.UniPort.Onu.PonPortID,
			"OnuSn":  s.UniPort.Onu.Sn(),
			"PortNo": s.UniPort.PortNo,
			"UniId":  s.UniPort.ID,
			"Name":   s.Name,
			"err":    err.Error(),
		}).Error("Can't start auth for this Service")
	}
}

// HandleDhcp is used to start DHCP for a particular Service when the corresponding flow is received
func (s *Service) HandleDhcp(pbit uint8, cTag int) {

	if s.CTag != cTag || (s.UsPonCTagPriority != pbit && pbit != 255) {
		serviceLogger.WithFields(log.Fields{
			"OnuId":                     s.UniPort.Onu.ID,
			"IntfId":                    s.UniPort.Onu.PonPortID,
			"OnuSn":                     s.UniPort.Onu.Sn(),
			"PortNo":                    s.UniPort.PortNo,
			"UniId":                     s.UniPort.ID,
			"Name":                      s.Name,
			"Service.CTag":              s.CTag,
			"Service.UsPonCTagPriority": s.UsPonCTagPriority,
			"cTag":                      cTag,
			"pbit":                      pbit,
		}).Debug("DHCP flow is not for this service, ignoring")
		return
	}

	if !s.NeedsDhcp {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
			"Name":      s.Name,
			"NeedsDhcp": s.NeedsDhcp,
		}).Trace("Won't start DHCP as it is not required")
		return
	}

	serviceLogger.WithFields(log.Fields{
		"OnuId":      s.UniPort.Onu.ID,
		"IntfId":     s.UniPort.Onu.PonPortID,
		"OnuSn":      s.UniPort.Onu.Sn(),
		"PortNo":     s.UniPort.PortNo,
		"UniId":      s.UniPort.ID,
		"Name":       s.Name,
		"NeedsEapol": s.NeedsEapol,
	}).Debug("Starting DHCP for the service")
	if err := s.DHCPState.Event("start_dhcp"); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.UniPort.Onu.ID,
			"IntfId": s.UniPort.Onu.PonPortID,
			"OnuSn":  s.UniPort.Onu.Sn(),
			"PortNo": s.UniPort.PortNo,
			"UniId":  s.UniPort.ID,
			"Name":   s.Name,
			"err":    err.Error(),
		}).Error("Can't start DHCP for this Service")
	}
}

func (s *Service) HandlePackets() {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.UniPort.Onu.ID,
		"IntfId":    s.UniPort.Onu.PonPortID,
		"OnuSn":     s.UniPort.Onu.Sn(),
		"GemPortId": s.GemPort,
		"PortNo":    s.UniPort.PortNo,
		"UniId":     s.UniPort.ID,
		"Name":      s.Name,
	}).Debug("Listening on Service Packet Channel")

	defer func() {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"GemPortId": s.GemPort,
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
			"Name":      s.Name,
		}).Debug("Done Listening on Service Packet Channel")
	}()

	for msg := range s.PacketCh {
		serviceLogger.WithFields(log.Fields{
			"OnuId":       s.UniPort.Onu.ID,
			"IntfId":      s.UniPort.Onu.PonPortID,
			"OnuSn":       s.UniPort.Onu.Sn(),
			"GemPortId":   s.GemPort,
			"PortNo":      s.UniPort.PortNo,
			"UniId":       s.UniPort.ID,
			"Name":        s.Name,
			"messageType": msg.Type,
		}).Trace("Received message on Service Packet Channel")

		if msg.Type == packetHandlers.EAPOL {
			eapol.HandleNextPacket(msg.OnuId, msg.IntfId, s.GemPort, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.UniPort.ID, s.Id, s.UniPort.Onu.PonPort.Olt.ID, s.EapolState, msg.Packet, s.Stream, nil)
		} else if msg.Type == packetHandlers.DHCP {
			_ = dhcp.HandleNextPacket(s.UniPort.Onu.ID, s.UniPort.Onu.PonPortID, s.Name, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.CTag, s.GemPort, s.UniPort.ID, s.HwAddress, s.DHCPState, msg.Packet, s.UsPonCTagPriority, s.Stream)
		} else if msg.Type == packetHandlers.IGMP {
			log.Warn(hex.EncodeToString(msg.Packet.Data()))
			_ = igmp.HandleNextPacket(s.UniPort.Onu.PonPortID, s.UniPort.Onu.ID, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.GemPort, s.HwAddress, msg.Packet, s.CTag, s.UsPonCTagPriority, s.Stream)
		}
	}
}

func (s *Service) HandleChannel() {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.UniPort.Onu.ID,
		"IntfId":    s.UniPort.Onu.PonPortID,
		"OnuSn":     s.UniPort.Onu.Sn(),
		"GemPortId": s.GemPort,
		"PortNo":    s.UniPort.PortNo,
		"UniId":     s.UniPort.ID,
		"Name":      s.Name,
	}).Debug("Listening on Service Channel")

	defer func() {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"GemPortId": s.GemPort,
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
			"Name":      s.Name,
		}).Debug("Done Listening on Service Channel")
	}()
	for msg := range s.Channel {
		switch msg.Type {
		case bbsimTypes.StartEAPOL:
			if err := s.handleEapolStart(s.Stream); err != nil {
				serviceLogger.WithFields(log.Fields{
					"OnuId":     s.UniPort.Onu.ID,
					"IntfId":    s.UniPort.Onu.PonPortID,
					"OnuSn":     s.UniPort.Onu.Sn(),
					"GemPortId": s.GemPort,
					"PortNo":    s.UniPort.PortNo,
					"UniId":     s.UniPort.ID,
					"Name":      s.Name,
					"err":       err,
				}).Error("Error while sending EapolStart packet")
				_ = s.EapolState.Event(eapol.EventAuthFailed)
			}
		case bbsimTypes.StartDHCP:
			if err := s.handleDHCPStart(s.Stream); err != nil {
				serviceLogger.WithFields(log.Fields{
					"OnuId":     s.UniPort.Onu.ID,
					"IntfId":    s.UniPort.Onu.PonPortID,
					"OnuSn":     s.UniPort.Onu.Sn(),
					"GemPortId": s.GemPort,
					"PortNo":    s.UniPort.PortNo,
					"UniId":     s.UniPort.ID,
					"Name":      s.Name,
					"err":       err,
				}).Error("Error while sending DHCPDiscovery packet")
				_ = s.DHCPState.Event("dhcp_failed")

			}
		case bbsimTypes.IGMPMembershipReportV2:
			igmpInfo, _ := msg.Data.(bbsimTypes.IgmpMessage)
			serviceLogger.WithFields(log.Fields{
				"OnuId":     s.UniPort.Onu.ID,
				"IntfId":    s.UniPort.Onu.PonPortID,
				"OnuSn":     s.UniPort.Onu.Sn(),
				"GemPortId": s.GemPort,
				"PortNo":    s.UniPort.PortNo,
				"UniId":     s.UniPort.ID,
				"Name":      s.Name,
			}).Debug("Received IGMPMembershipReportV2 message on ONU channel")
			_ = igmp.SendIGMPMembershipReportV2(s.UniPort.Onu.PonPortID, s.UniPort.Onu.ID, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.GemPort, s.HwAddress, s.CTag, s.UsPonCTagPriority, s.Stream, igmpInfo.GroupAddress)
		case bbsimTypes.IGMPLeaveGroup:
			igmpInfo, _ := msg.Data.(bbsimTypes.IgmpMessage)
			serviceLogger.WithFields(log.Fields{
				"OnuId":     s.UniPort.Onu.ID,
				"IntfId":    s.UniPort.Onu.PonPortID,
				"OnuSn":     s.UniPort.Onu.Sn(),
				"GemPortId": s.GemPort,
				"PortNo":    s.UniPort.PortNo,
				"UniId":     s.UniPort.ID,
				"Name":      s.Name,
			}).Debug("Received IGMPLeaveGroupV2 message on ONU channel")
			_ = igmp.SendIGMPLeaveGroupV2(s.UniPort.Onu.PonPortID, s.UniPort.Onu.ID, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.GemPort, s.HwAddress, s.CTag, s.UsPonCTagPriority, s.Stream, igmpInfo.GroupAddress)
		case bbsimTypes.IGMPMembershipReportV3:
			igmpInfo, _ := msg.Data.(bbsimTypes.IgmpMessage)
			serviceLogger.WithFields(log.Fields{
				"OnuId":     s.UniPort.Onu.ID,
				"IntfId":    s.UniPort.Onu.PonPortID,
				"OnuSn":     s.UniPort.Onu.Sn(),
				"GemPortId": s.GemPort,
				"PortNo":    s.UniPort.PortNo,
				"UniId":     s.UniPort.ID,
				"Name":      s.Name,
			}).Debug("Received IGMPMembershipReportV3 message on ONU channel")
			_ = igmp.SendIGMPMembershipReportV3(s.UniPort.Onu.PonPortID, s.UniPort.Onu.ID, s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.GemPort, s.HwAddress, s.CTag, s.UsPonCTagPriority, s.Stream, igmpInfo.GroupAddress)

		}
	}
}

func (s *Service) Initialize(stream bbsimTypes.Stream) {
	if err := s.InternalState.Event(ServiceTxInitialize, stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"GemPortId": s.GemPort,
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
			"Name":      s.Name,
			"Err":       err,
		}).Error("Cannot initialize service")
	}
}

func (s *Service) Disable() {
	if err := s.InternalState.Event(ServiceTxDisable); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"GemPortId": s.GemPort,
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
			"Name":      s.Name,
			"Err":       err,
		}).Error("Cannot disable service")
	}
}

func (s *Service) handleEapolStart(stream bbsimTypes.Stream) error {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.UniPort.Onu.ID,
		"IntfId":    s.UniPort.Onu.PonPortID,
		"OnuSn":     s.UniPort.Onu.Sn(),
		"GemPortId": s.GemPort,
		"PortNo":    s.UniPort.PortNo,
		"UniId":     s.UniPort.ID,
		"GemPort":   s.GemPort,
		"Name":      s.Name,
	}).Trace("handleEapolStart")

	if err := eapol.SendEapStart(s.UniPort.Onu.ID, s.UniPort.Onu.PonPortID, s.UniPort.Onu.Sn(), s.UniPort.PortNo,
		s.HwAddress, s.GemPort, s.UniPort.ID, s.EapolState, stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":   s.UniPort.Onu.ID,
			"IntfId":  s.UniPort.Onu.PonPortID,
			"OnuSn":   s.UniPort.Onu.Sn(),
			"GemPort": s.GemPort,
			"Name":    s.Name,
		}).Error("handleEapolStart")
		return err
	}
	return nil
}

func (s *Service) handleDHCPStart(stream bbsimTypes.Stream) error {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.UniPort.Onu.ID,
		"IntfId":    s.UniPort.Onu.PonPortID,
		"OnuSn":     s.UniPort.Onu.Sn(),
		"Name":      s.Name,
		"GemPortId": s.GemPort,
		"PortNo":    s.UniPort.PortNo,
		"UniId":     s.UniPort.ID,
	}).Debugf("HandleDHCPStart")

	if err := dhcp.SendDHCPDiscovery(s.UniPort.Onu.PonPortID, s.UniPort.Onu.ID, s.Name, int(s.CTag), s.GemPort,
		s.UniPort.Onu.Sn(), s.UniPort.PortNo, s.UniPort.ID, s.DHCPState, s.HwAddress, s.UsPonCTagPriority, stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.UniPort.Onu.ID,
			"IntfId":    s.UniPort.Onu.PonPortID,
			"OnuSn":     s.UniPort.Onu.Sn(),
			"Name":      s.Name,
			"GemPortId": s.GemPort,
			"PortNo":    s.UniPort.PortNo,
			"UniId":     s.UniPort.ID,
		}).Error("HandleDHCPStart")
		return err
	}
	return nil
}

func (s *Service) logStateChange(stateMachine string, src string, dst string) {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.UniPort.Onu.ID,
		"IntfId":    s.UniPort.Onu.PonPortID,
		"OnuSn":     s.UniPort.Onu.Sn(),
		"GemPortId": s.GemPort,
		"PortNo":    s.UniPort.PortNo,
		"UniId":     s.UniPort.ID,
		"Name":      s.Name,
	}).Debugf("Changing Service.%s InternalState from %s to %s", stateMachine, src, dst)
}
