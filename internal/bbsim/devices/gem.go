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
	"time"
)

var gemLogger = log.WithFields(log.Fields{
	"module": "GEM",
})

type GemPort struct {
	ID         int32
	Onu        *Onu
	FlowID     uint32
	VlanID     uint32
	NeedsEapol bool
	EapolState *fsm.FSM // Holds the state of EAPOL based authentication for this GemPort
	NeedsDhcp  bool
	DHCPState  *fsm.FSM // Holds the DHCP state
	PacketCh   chan OnuPacketMessage
}

func NewGemPort(onu *Onu, ID int32, flowId uint32, vlanID uint32, eapol bool, dhcp bool) (*GemPort, error) {
	gem := GemPort{
		ID:         ID,
		Onu:        onu,
		FlowID:     flowId,
		VlanID:     vlanID,
		NeedsEapol: eapol,
		NeedsDhcp:  dhcp,
		PacketCh:   make(chan OnuPacketMessage),
	}

	gem.EapolState = fsm.NewFSM(
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
				gem.logStateChange("EapolState", e.Src, e.Dst)
			},
		},
	)

	gem.DHCPState = fsm.NewFSM(
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
				gem.logStateChange("DHCPState", e.Src, e.Dst)
			},
		},
	)

	// is there a better way to get the OpenoltStream gRPC here?
	go gem.handlePackets(gem.Onu.PonPort.Olt.OpenoltStream)

	return &gem, nil
}

func (gem *GemPort) HandleAuth(stream bbsimTypes.Stream) {
	if err := gem.EapolState.Event("start_auth"); err != nil {
		gemLogger.WithFields(log.Fields{
			"OnuId":     gem.Onu.ID,
			"IntfId":    gem.Onu.PonPortID,
			"OnuSn":     gem.Onu.Sn(),
			"GemPortId": gem.ID,
			"err":       err.Error(),
		}).Error("Can't start auth for this Gem")
	} else {
		if err := gem.HandleEapolStart(stream); err != nil {
			gemLogger.WithFields(log.Fields{
				"OnuId":     gem.Onu.ID,
				"IntfId":    gem.Onu.PonPortID,
				"OnuSn":     gem.Onu.Sn(),
				"GemPortId": gem.ID,
				"err":       err,
			}).Error("Error while sending EapolStart packet")
			gem.EapolState.Event("auth_failed")
		}
	}
}

func (gem *GemPort) HandleEapolStart(stream bbsimTypes.Stream) error {
	gemLogger.WithFields(log.Fields{
		"OnuId":     gem.Onu.ID,
		"IntfId":    gem.Onu.PonPortID,
		"OnuSn":     gem.Onu.Sn(),
		"GemPortId": gem.ID,
	}).Debugf("HandleEapolStart")

	if err := eapol.SendEapStart(gem.Onu.ID, gem.Onu.PonPortID, gem.Onu.Sn(), gem.Onu.PortNo,
		gem.Onu.HwAddress, gem.ID, gem.EapolState, stream); err != nil {
		gemLogger.WithFields(log.Fields{
			"OnuId":     gem.Onu.ID,
			"IntfId":    gem.Onu.PonPortID,
			"OnuSn":     gem.Onu.Sn(),
			"GemPortId": gem.ID,
		}).Error("HandleEapolStart")
		return err
	}
	return nil
}

func (gem *GemPort) handlePackets(stream bbsimTypes.Stream) {
	gemLogger.WithFields(log.Fields{
		"OnuId":     gem.Onu.ID,
		"IntfId":    gem.Onu.PonPortID,
		"OnuSn":     gem.Onu.Sn(),
		"GemPortId": gem.ID,
	}).Debug("Listening on GemPort Packet Channel")

	defer func() {
		gemLogger.WithFields(log.Fields{
			"OnuId":     gem.Onu.ID,
			"IntfId":    gem.Onu.PonPortID,
			"OnuSn":     gem.Onu.Sn(),
			"GemPortId": gem.ID,
		}).Debug("Done Listening on GemPort Packet Channel")
	}()

loop:
	for {
		select {
		case msg, ok := <-gem.PacketCh:
			if !ok {
				onuLogger.WithFields(log.Fields{
					"OnuId":     gem.Onu.ID,
					"IntfId":    gem.Onu.PonPortID,
					"OnuSn":     gem.Onu.Sn(),
					"GemPortId": gem.ID,
					"ok":        ok,
					"msg":       msg,
				}).Warn("GemPort Packet Channel handling canceled via channel close")
				break loop
			}
			gemLogger.WithFields(log.Fields{
				"OnuId":       gem.Onu.ID,
				"IntfId":      gem.Onu.PonPortID,
				"OnuSn":       gem.Onu.Sn(),
				"GemPortId":   gem.ID,
				"messageType": msg.Type,
			}).Debug("Received message on GemPort Packet Channel")

			if msg.Type == packetHandlers.EAPOL {
				eapol.HandleNextPacket(msg.OnuId, msg.IntfId, gem.ID, gem.Onu.Sn(), gem.Onu.PortNo, gem.EapolState, msg.Packet, stream, nil)
			} else if msg.Type == packetHandlers.DHCP {
				gemLogger.WithFields(log.Fields{
					"OnuId":       gem.Onu.ID,
					"IntfId":      gem.Onu.PonPortID,
					"OnuSn":       gem.Onu.Sn(),
					"GemPortId":   gem.ID,
					"messageType": msg.Type,
				}).Debug("Received DHCP packet on GemPort Packet Channel")
				dhcp.HandleNextPacket(gem.Onu.PonPort.Olt.ID, gem.Onu.ID, gem.Onu.PonPortID, gem.Onu.Sn(), gem.Onu.PortNo, int(gem.VlanID), int(gem.ID), gem.Onu.HwAddress, gem.DHCPState, msg.Packet, stream)
			}
		}
	}
}

func (gem *GemPort) HandleDHCP(stream bbsimTypes.Stream) {

	if gem.DHCPState.Current() != "created" {
		// if DHCP is in process do not restart it
		return
	}

	defer func() {
		// if DHCP does not complete in a reasonable amount of time, set the state to dhcp_failed
		time.Sleep(30 * time.Second)
		gem.DHCPState.Event("dhcp_failed")
	}()

	if err := gem.DHCPState.Event("start_dhcp"); err != nil {
		gemLogger.WithFields(log.Fields{
			"OnuId":     gem.Onu.ID,
			"IntfId":    gem.Onu.PonPortID,
			"OnuSn":     gem.Onu.Sn(),
			"GemPortId": gem.ID,
			"err":       err.Error(),
		}).Error("Can't start DHCP for this Gem")
	} else {
		if err := gem.HandleDHCPStart(stream); err != nil {
			gemLogger.WithFields(log.Fields{
				"OnuId":     gem.Onu.ID,
				"IntfId":    gem.Onu.PonPortID,
				"OnuSn":     gem.Onu.Sn(),
				"GemPortId": gem.ID,
				"err":       err,
			}).Error("Error while sending DHCPDiscovery packet")
			gem.EapolState.Event("dhcp_failed")
		}
	}
}

func (gem *GemPort) HandleDHCPStart(stream bbsimTypes.Stream) error {
	gemLogger.WithFields(log.Fields{
		"OnuId":     gem.Onu.ID,
		"IntfId":    gem.Onu.PonPortID,
		"OnuSn":     gem.Onu.Sn(),
		"GemPortId": gem.ID,
	}).Debugf("HandleDHCPStart")

	if err := dhcp.SendDHCPDiscovery(gem.Onu.PonPort.Olt.ID, gem.Onu.PonPortID, gem.Onu.ID, int(gem.VlanID), int(gem.ID),
		gem.Onu.Sn(), gem.Onu.PortNo, gem.DHCPState, gem.Onu.HwAddress, stream); err != nil {
		gemLogger.WithFields(log.Fields{
			"OnuId":     gem.Onu.ID,
			"IntfId":    gem.Onu.PonPortID,
			"OnuSn":     gem.Onu.Sn(),
			"GemPortId": gem.ID,
		}).Error("HandleDHCPStart")
		return err
	}
	return nil
}

func (gem *GemPort) logStateChange(stateMachine string, src string, dst string) {
	gemLogger.WithFields(log.Fields{
		"OnuId":     gem.Onu.ID,
		"IntfId":    gem.Onu.PonPortID,
		"OnuSn":     gem.Onu.Sn(),
		"GemPortId": gem.ID,
	}).Debugf("Changing GemPort.%s InternalState from %s to %s", stateMachine, src, dst)
}
