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
	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	log "github.com/sirupsen/logrus"
)

var gemLogger = log.WithFields(log.Fields{
	"module": "GEM",
})

// TODO add VlanInfos
type GemPort struct {
	ID         int32
	Onu        *Onu
	FlowID     uint32
	EapolState *fsm.FSM // Holds the state of EAPOL based authentication for this GemPort
	DHCPState  *fsm.FSM // Holds the DHCP state
	PacketCh   chan OnuPacketMessage
}

func NewGemPort(onu *Onu, ID int32, flowId uint32) (*GemPort, error) {
	gem := GemPort{
		ID:     ID,
		Onu:    onu,
		FlowID: flowId,
		PacketCh: make(chan OnuPacketMessage),
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
				gem.logStateChange(e.Src, e.Dst)
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
				gem.logStateChange(e.Src, e.Dst)
			},
		},
	)

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
		go gem.handlePackets(stream)

		//time.Sleep(1 * time.Second)

		if err := gem.HandleEapolStart(stream); err != nil {
			gemLogger.WithFields(log.Fields{
				"OnuId":     gem.Onu.ID,
				"IntfId":    gem.Onu.PonPortID,
				"OnuSn":     gem.Onu.Sn(),
				"GemPortId": gem.ID,
				"err":       err,
			}).Error("Error while sending EapolStart packet")
			gem.EapolState.Event("auth_failed")
			close(gem.PacketCh)
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

			eapol.HandleNextPacket(msg.OnuId, msg.IntfId, gem.ID, gem.Onu.Sn(), gem.Onu.PortNo, gem.EapolState, msg.Packet, stream, nil)

		}
	}
}

func (gem *GemPort) logStateChange(src string, dst string) {
	gemLogger.WithFields(log.Fields{
		"OnuId":     gem.Onu.ID,
		"IntfId":    gem.Onu.PonPortID,
		"OnuSn":     gem.Onu.Sn(),
		"GemPortId": gem.ID,
	}).Debugf("Changing GemPort InternalState from %s to %s", src, dst)
}
