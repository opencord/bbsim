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
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
	"net"
)

var serviceLogger = log.WithFields(log.Fields{
	"module": "SERVICE",
})

type ServiceIf interface {
	HandleAuth(stream bbsimTypes.Stream)
}


type Service struct {
	Name                string
	HwAddress           net.HardwareAddr
	Onu                 *Onu
	CTag                int
	STag                int
	NeedsEapol          bool
	NeedsDchp           bool
	NeedsIgmp           bool
	TechnologyProfileID int
	UniTagMatch         int

	// state
	GemPort    uint32
	EapolState *fsm.FSM
	DHCPState  *fsm.FSM
	PacketCh   chan OnuPacketMessage
}

func NewService(name string, hwAddress net.HardwareAddr, onu *Onu, cTag int, sTag int, needsEapol bool, needsDchp bool, needsIgmp bool) (*Service, error) {

	service := Service{
		Name:                name,
		HwAddress:           hwAddress,
		Onu:                 onu,
		CTag:                cTag,
		STag:                sTag,
		NeedsEapol:          needsEapol,
		NeedsDchp:           needsDchp,
		NeedsIgmp:           needsIgmp,
		PacketCh:            make(chan OnuPacketMessage),
		TechnologyProfileID: 64, // FIXME extend config to pass this in the service definition
		UniTagMatch:         0,  // FIXME extend config to pass this in the service definition
	}

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
		},
	)

	service.DHCPState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "start_dhcp", Src: []string{"created", "eap_response_success_received", "dhcp_discovery_sent", "dhcp_request_sent", "dhcp_ack_received", "dhcp_failed"}, Dst: "dhcp_started"},
			{Name: "dhcp_discovery_sent", Src: []string{"dhcp_started"}, Dst: "dhcp_discovery_sent"},
			{Name: "dhcp_request_sent", Src: []string{"dhcp_discovery_sent"}, Dst: "dhcp_request_sent"},
			{Name: "dhcp_ack_received", Src: []string{"dhcp_request_sent"}, Dst: "dhcp_ack_received"},
			{Name: "dhcp_failed", Src: []string{"dhcp_started", "dhcp_discovery_sent", "dhcp_request_sent"}, Dst: "dhcp_failed"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				service.logStateChange("DHCPState", e.Src, e.Dst)
			},
		},
	)

	// is there a better way to get the OpenoltStream gRPC here?
	// TODO listen for packets once the ONU is active
	//go service.handlePackets(*service.Onu.PonPort.Olt.OpenoltStream)

	return &service, nil
}

func (s *Service) HandleAuth(stream bbsimTypes.Stream) {
	if err := s.EapolState.Event("start_auth"); err != nil {
		serviceLogger.WithFields(log.Fields{
			"OnuId":  s.Onu.ID,
			"IntfId": s.Onu.PonPortID,
			"OnuSn":  s.Onu.Sn(),
			"Name":   s.Name,
			"err":    err.Error(),
		}).Error("Can't start auth for this Gem")
	} else {
		if err := s.handleEapolStart(stream); err != nil {
			serviceLogger.WithFields(log.Fields{
				"OnuId":  s.Onu.ID,
				"IntfId": s.Onu.PonPortID,
				"OnuSn":  s.Onu.Sn(),
				"Name":   s.Name,
				"err":    err,
			}).Error("Error while sending EapolStart packet")
			s.EapolState.Event("auth_failed")
		}
	}
}

func (s *Service) handleEapolStart(stream bbsimTypes.Stream) error {
	serviceLogger.WithFields(log.Fields{
		"OnuId":   s.Onu.ID,
		"IntfId":  s.Onu.PonPortID,
		"OnuSn":   s.Onu.Sn(),
		"GemPort": s.GemPort,
		"Name":    s.Name,
	}).Debugf("handleEapolStart")

	go s.handlePackets(stream)

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

func (s *Service) handlePackets(stream bbsimTypes.Stream) {
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
loop:
	for {
		select {
		case msg, ok := <-s.PacketCh:
			if !ok {
				onuLogger.WithFields(log.Fields{
					"OnuId":  s.Onu.ID,
					"IntfId": s.Onu.PonPortID,
					"OnuSn":  s.Onu.Sn(),
					"Name":   s.Name,
					"ok":     ok,
					"msg":    msg,
				}).Warn("Service Packet Channel handling canceled via channel close")
				break loop
			}
			serviceLogger.WithFields(log.Fields{
				"OnuId":       s.Onu.ID,
				"IntfId":      s.Onu.PonPortID,
				"OnuSn":       s.Onu.Sn(),
				"Name":        s.Name,
				"messageType": msg.Type,
			}).Debug("Received message on Service Packet Channel")

			if msg.Type == packetHandlers.EAPOL {
				eapol.HandleNextPacket(msg.OnuId, msg.IntfId, s.GemPort, s.Onu.Sn(), s.Onu.PortNo, s.EapolState, msg.Packet, stream, nil)
			} else if msg.Type == packetHandlers.DHCP {
				serviceLogger.WithFields(log.Fields{
					"OnuId":       s.Onu.ID,
					"IntfId":      s.Onu.PonPortID,
					"OnuSn":       s.Onu.Sn(),
					"Name":        s.Name,
					"messageType": msg.Type,
				}).Warn("I'm not ready to handle DHCP packets yet!")
			}
		}
	}
}

func (s *Service) logStateChange(stateMachine string, src string, dst string) {
	serviceLogger.WithFields(log.Fields{
		"OnuId":  s.Onu.ID,
		"IntfId": s.Onu.PonPortID,
		"OnuSn":  s.Onu.Sn(),
		"Name":   s.Name,
	}).Debugf("Changing Service.%s InternalState from %s to %s", stateMachine, src, dst)
}
