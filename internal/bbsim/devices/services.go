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
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/bbsim/internal/bbsim/responders/dhcp"
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
	"net"
)

var serviceLogger = log.WithFields(log.Fields{
	"module": "SERVICE",
})

type ServiceIf interface {
	HandlePackets()      // start listening on the PacketCh
	HandleAuth()         // Sends the EapoStart packet
	HandleDhcp(cTag int) // Sends the DHCPDiscover packet

	Initialize(stream bbsimTypes.Stream)
	Disable()
}

type Service struct {
	Name                string
	HwAddress           net.HardwareAddr
	Onu                 *Onu
	CTag                int
	STag                int
	NeedsEapol          bool
	NeedsDhcp           bool
	NeedsIgmp           bool
	TechnologyProfileID int
	UniTagMatch         int
	ConfigureMacAddress bool
	UsPonCTagPriority   uint8
	UsPonSTagPriority   uint8
	DsPonCTagPriority   uint8
	DsPonSTagPriority   uint8

	// state
	GemPort       uint32
	InternalState *fsm.FSM
	EapolState    *fsm.FSM
	DHCPState     *fsm.FSM
	Channel       chan Message          // drive Service lifecycle
	PacketCh      chan OnuPacketMessage // handle packets
	Stream        bbsimTypes.Stream     // the gRPC stream to communicate with the adapter, created in the initialize transition
}

func NewService(name string, hwAddress net.HardwareAddr, onu *Onu, cTag int, sTag int,
	needsEapol bool, needsDchp bool, needsIgmp bool, tpID int, uniTagMatch int, configMacAddress bool,
	usPonCTagPriority uint8, usPonSTagPriority uint8, dsPonCTagPriority uint8, dsPonSTagPriority uint8) (*Service, error) {

	service := Service{
		Name:                name,
		HwAddress:           hwAddress,
		Onu:                 onu,
		CTag:                cTag,
		STag:                sTag,
		NeedsEapol:          needsEapol,
		NeedsDhcp:           needsDchp,
		NeedsIgmp:           needsIgmp,
		TechnologyProfileID: tpID,
		UniTagMatch:         uniTagMatch,
		ConfigureMacAddress: configMacAddress,
		UsPonCTagPriority:   usPonCTagPriority,
		UsPonSTagPriority:   usPonSTagPriority,
		DsPonCTagPriority:   dsPonCTagPriority,
		DsPonSTagPriority:   dsPonSTagPriority,
	}

	service.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "initialized", Src: []string{"created", "disabled"}, Dst: "initialized"},
			{Name: "disabled", Src: []string{"initialized"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("InternalState", e.Src, e.Dst)
			},
			"enter_initialized": func(e *fsm.Event) {

				stream, ok := e.Args[0].(bbsimTypes.Stream)
				if !ok {
					serviceLogger.Fatal("initialize invoke with wrong arguments")
				}

				service.Stream = stream

				service.PacketCh = make(chan OnuPacketMessage)
				service.Channel = make(chan Message)

				go service.HandlePackets()
				go service.HandleChannel()
			},
			"enter_disabled": func(e *fsm.Event) {
				// reset the state machines
				service.EapolState.SetState("created")
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
		"created",
		fsm.Events{
			{Name: "start_auth", Src: []string{"created", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent", "eap_response_success_received", "auth_failed"}, Dst: "auth_started"},
			{Name: "eap_start_sent", Src: []string{"auth_started"}, Dst: "eap_start_sent"},
			{Name: "eap_response_identity_sent", Src: []string{"eap_start_sent"}, Dst: "eap_response_identity_sent"},
			{Name: "eap_response_challenge_sent", Src: []string{"eap_response_identity_sent"}, Dst: "eap_response_challenge_sent"},
			{Name: "eap_response_success_received", Src: []string{"eap_response_challenge_sent"}, Dst: "eap_response_success_received"},
			{Name: "auth_failed", Src: []string{"auth_started", "eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent"}, Dst: "auth_failed"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("EapolState", e.Src, e.Dst)
			},
			"before_start_auth": func(e *fsm.Event) {
				msg := Message{
					Type: StartEAPOL,
				}
				service.Channel <- msg
			},
		},
	)

	service.DHCPState = fsm.NewFSM(
		"created",
		fsm.Events{
			// TODO only allow transitions to dhcp_start from success or failure, not in-between states
			// TODO forcefully fail DHCP if we don't get an ack in X seconds
			{Name: "start_dhcp", Src: []string{"created", "dhcp_discovery_sent", "dhcp_request_sent", "dhcp_ack_received", "dhcp_failed"}, Dst: "dhcp_started"},
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
				msg := Message{
					Type: StartDHCP,
				}
				service.Channel <- msg
			},
		},
	)

	return &service, nil
}

// HandleAuth is used to start EAPOL for a particular Service when the corresponding flow is received
func (s *Service) HandleAuth() {

	if !s.NeedsEapol {
		serviceLogger.WithFields(log.Fields{
			"OnuId":      s.Onu.ID,
			"IntfId":     s.Onu.PonPortID,
			"OnuSn":      s.Onu.Sn(),
			"Name":       s.Name,
			"NeedsEapol": s.NeedsEapol,
		}).Debug("Won't start authentication as EAPOL is not required")
		return
	}

	if err := s.EapolState.Event("start_auth"); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
			"err":    err.Error(),
		}).Error("Can't start auth for this Service")
	}
}

// HandleDhcp is used to start DHCP for a particular Service when the corresponding flow is received
func (s *Service) HandleDhcp(cTag int) {

	if s.CTag != cTag {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
		}).Trace("DHCP flow is not for this service, ignoring")
		return
	}

	if !s.NeedsDhcp {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.Onu.ID,
			"IntfId":    s.Onu.PonPortID,
			"OnuSn":     s.Onu.Sn(),
			"Name":      s.Name,
			"NeedsDhcp": s.NeedsDhcp,
		}).Trace("Won't start DHCP as it is not required")
		return
	}

	// TODO check if the DHCP flow was received before starting auth

	if err := s.DHCPState.Event("start_dhcp"); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
			"err":    err.Error(),
		}).Error("Can't start DHCP for this Service")
	}
}

func (s *Service) HandlePackets() {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.Onu.ID,
		"IntfId":    s.Onu.PonPortID,
		"OnuSn":     s.Onu.Sn(),
		"GemPortId": s.GemPort,
		"Name":      s.Name,
	}).Debug("Listening on Service Packet Channel")

	defer func() {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.Onu.ID,
			"IntfId":    s.Onu.PonPortID,
			"OnuSn":     s.Onu.Sn(),
			"GemPortId": s.GemPort,
			"Name":      s.Name,
		}).Debug("Done Listening on Service Packet Channel")
	}()

	for msg := range s.PacketCh {
		serviceLogger.WithFields(log.Fields{
			"OnuId":       s.Onu.ID,
			"IntfId":      s.Onu.PonPortID,
			"OnuSn":       s.Onu.Sn(),
			"Name":        s.Name,
			"messageType": msg.Type,
		}).Trace("Received message on Service Packet Channel")

		if msg.Type == packetHandlers.EAPOL {
			eapol.HandleNextPacket(msg.OnuId, msg.IntfId, s.GemPort, s.Onu.Sn(), s.Onu.PortNo, s.EapolState, msg.Packet, s.Stream, nil)
		} else if msg.Type == packetHandlers.DHCP {
			_ = dhcp.HandleNextPacket(s.Onu.PonPort.Olt.ID, s.Onu.ID, s.Onu.PonPortID, s.Name, s.Onu.Sn(), s.Onu.PortNo, s.CTag, s.GemPort, s.HwAddress, s.DHCPState, msg.Packet, s.UsPonCTagPriority, s.Stream)
		}
	}
}

func (s *Service) HandleChannel() {
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.Onu.ID,
		"IntfId":    s.Onu.PonPortID,
		"OnuSn":     s.Onu.Sn(),
		"GemPortId": s.GemPort,
		"Name":      s.Name,
	}).Debug("Listening on Service Channel")

	defer func() {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.Onu.ID,
			"IntfId":    s.Onu.PonPortID,
			"OnuSn":     s.Onu.Sn(),
			"GemPortId": s.GemPort,
			"Name":      s.Name,
		}).Debug("Done Listening on Service Channel")
	}()
	for msg := range s.Channel {
		if msg.Type == StartEAPOL {
			if err := s.handleEapolStart(s.Stream); err != nil {
				serviceLogger.WithFields(log.Fields{
					"OnuId":  s.Onu.ID,
					"IntfId": s.Onu.PonPortID,
					"OnuSn":  s.Onu.Sn(),
					"Name":   s.Name,
					"err":    err,
				}).Error("Error while sending EapolStart packet")
				_ = s.EapolState.Event("auth_failed")
			}
		} else if msg.Type == StartDHCP {
			if err := s.handleDHCPStart(s.Stream); err != nil {
				serviceLogger.WithFields(log.Fields{
					"OnuId":  s.Onu.ID,
					"IntfId": s.Onu.PonPortID,
					"OnuSn":  s.Onu.Sn(),
					"Name":   s.Name,
					"err":    err,
				}).Error("Error while sending DHCPDiscovery packet")
				_ = s.DHCPState.Event("dhcp_failed")
			}
		}
	}
}

func (s *Service) Initialize(stream bbsimTypes.Stream) {
	if err := s.InternalState.Event("initialized", stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
			"Err":    err,
		}).Error("Cannot initialize service")
	}
}

func (s *Service) Disable() {
	if err := s.InternalState.Event("disabled"); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
			"Err":    err,
		}).Error("Cannot disable service")
	}
}

func (s *Service) handleEapolStart(stream bbsimTypes.Stream) error {
	// TODO fail Auth if it does not succeed in 30 seconds
	serviceLogger.WithFields(log.Fields{
		"OnuId":   s.Onu.ID,
		"IntfId":  s.Onu.PonPortID,
		"OnuSn":   s.Onu.Sn(),
		"GemPort": s.GemPort,
		"Name":    s.Name,
	}).Trace("handleEapolStart")

	if err := eapol.SendEapStart(s.Onu.ID, s.Onu.PonPortID, s.Onu.Sn(), s.Onu.PortNo,
		s.HwAddress, s.GemPort, s.EapolState, stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":   s.Onu.ID,
			"IntfId":  s.Onu.PonPortID,
			"OnuSn":   s.Onu.Sn(),
			"GemPort": s.GemPort,
			"Name":    s.Name,
		}).Error("handleEapolStart")
		return err
	}
	return nil
}

func (s *Service) handleDHCPStart(stream bbsimTypes.Stream) error {
	// TODO fail DHCP if it does not succeed in 30 seconds
	serviceLogger.WithFields(log.Fields{
		"OnuId":     s.Onu.ID,
		"IntfId":    s.Onu.PonPortID,
		"OnuSn":     s.Onu.Sn(),
		"Name":      s.Name,
		"GemPortId": s.GemPort,
	}).Debugf("HandleDHCPStart")

	if err := dhcp.SendDHCPDiscovery(s.Onu.PonPort.Olt.ID, s.Onu.PonPortID, s.Onu.ID, s.Name, int(s.CTag), s.GemPort,
		s.Onu.Sn(), s.Onu.PortNo, s.DHCPState, s.HwAddress, s.UsPonCTagPriority, stream); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":     s.Onu.ID,
			"IntfId":    s.Onu.PonPortID,
			"OnuSn":     s.Onu.Sn(),
			"Name":      s.Name,
			"GemPortId": s.GemPort,
		}).Error("HandleDHCPStart")
		return err
	}
	return nil
}

func (s *Service) logStateChange(stateMachine string, src string, dst string) {
	serviceLogger.WithFields(log.Fields{
		"OnuId":  s.Onu.ID,
		"IntfId": s.Onu.PonPortID,
		"OnuSn":  s.Onu.Sn(),
		"Name":   s.Name,
	}).Debugf("Changing Service.%s InternalState from %s to %s", stateMachine, src, dst)
}
