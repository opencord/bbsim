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
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	log "github.com/sirupsen/logrus"
	"net"
)

var uniLogger = log.WithFields(log.Fields{
	"module": "UNI",
})

const (
	maxUniPorts = 4

	UniStateUp   = "up"
	UniStateDown = "down"

	uniTxEnable  = "enable"
	uniTxDisable = "disable"
)

type UniPortIf interface {
	StorePortNo(portNo uint32)
	UpdateStream(stream bbsimTypes.Stream)
	Enable() error
	Disable() error

	HandlePackets()                  // start listening on the PacketCh
	HandleAuth()                     // Sends the EapoStart packet
	HandleDhcp(pbit uint8, cTag int) // Sends the DHCPDiscover packet
}

type UniPort struct {
	ID        uint32
	MeId      omcilib.EntityID
	PortNo    uint32
	OperState *fsm.FSM
	Onu       *Onu
	Services  []ServiceIf
	logger    *log.Entry
	PacketCh  chan bbsimTypes.OnuPacketMessage // handle packets
}

func NewUniPort(ID uint32, onu *Onu, nextCtag map[string]int, nextStag map[string]int) (*UniPort, error) {

	// IDs starts from 0, thus the maximum UNI supported is maxUniPorts - 1
	if ID > (maxUniPorts - 1) {
		return nil, fmt.Errorf("%d-is-higher-than-the-maximum-supported-unis-%d", ID, maxUniPorts)
	}

	uni := UniPort{
		ID:   ID,
		Onu:  onu,
		MeId: omcilib.GenerateUniPortEntityId(ID + 1),
	}

	uni.logger = uniLogger.WithFields(log.Fields{
		"UniId": uni.ID,
		"OnuSn": onu.Sn(),
	})

	uni.OperState = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: uniTxEnable, Src: []string{UniStateDown}, Dst: UniStateUp},
			{Name: uniTxDisable, Src: []string{UniStateUp}, Dst: UniStateDown},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				uni.logger.Debugf("changing-uni-operstate-from-%s-to-%s", e.Src, e.Dst)
			},
			fmt.Sprintf("enter_%s", UniStateUp): func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.UniStatusAlarm,
					Data: bbsimTypes.UniStatusAlarmMessage{
						OnuSN:          uni.Onu.SerialNumber,
						OnuID:          uni.Onu.ID,
						AdminState:     0,
						EntityID:       uni.MeId.ToUint16(),
						RaiseOMCIAlarm: false, // never raise an LOS when enabling a UNI
					},
				}
				uni.Onu.Channel <- msg
				go uni.HandlePackets()
				for _, s := range uni.Services {
					s.Initialize(uni.Onu.PonPort.Olt.OpenoltStream)
				}
			},
			fmt.Sprintf("enter_%s", UniStateDown): func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.UniStatusAlarm,
					Data: bbsimTypes.UniStatusAlarmMessage{
						OnuSN:          uni.Onu.SerialNumber,
						OnuID:          uni.Onu.ID,
						AdminState:     1,
						EntityID:       uni.MeId.ToUint16(),
						RaiseOMCIAlarm: true, // raise an LOS when disabling a UNI
					},
				}
				uni.Onu.Channel <- msg
				for _, s := range uni.Services {
					s.Disable()
				}
			},
		},
	)

	for k, s := range common.Services {

		// find the correct cTag for this service
		if _, ok := nextCtag[s.Name]; !ok {
			// it's the first time we iterate over this service,
			// so we start from the config value
			nextCtag[s.Name] = s.CTag
		} else {
			// we have a previous value, so we check it
			// if Allocation is unique, we increment,
			// otherwise (shared) we do nothing
			if s.CTagAllocation == common.TagAllocationUnique.String() {
				nextCtag[s.Name] = nextCtag[s.Name] + 1

				// the max valid value for a tag is 4096
				// check we're not going over
				if nextCtag[s.Name] > 4096 {
					uni.logger.WithFields(log.Fields{
						"cTag":    nextCtag[s.Name],
						"Service": s.Name,
					}).Fatal("c-tag-limit-reached-too-many-subscribers")
				}
			}
		}

		// find the correct sTag for this service
		if _, ok := nextStag[s.Name]; !ok {
			nextStag[s.Name] = s.STag
		} else {
			if s.STagAllocation == common.TagAllocationUnique.String() {
				nextStag[s.Name] = nextStag[s.Name] + 1

				// the max valid value for a tag is 4096
				// check we're not going over
				if nextStag[s.Name] > 4096 {
					uni.logger.WithFields(log.Fields{
						"cTag":    nextCtag[s.Name],
						"Service": s.Name,
					}).Fatal("s-tag-limit-reached-too-many-subscribers")
				}
			}
		}

		mac := net.HardwareAddr{0x2e, byte(olt.ID), byte(onu.PonPortID), byte(onu.ID), byte(uni.ID), byte(k)}
		service, err := NewService(uint32(k), s.Name, mac, &uni, nextCtag[s.Name], nextStag[s.Name],
			s.NeedsEapol, s.NeedsDhcp, s.NeedsIgmp, s.TechnologyProfileID, s.UniTagMatch,
			s.ConfigureMacAddress, s.UsPonCTagPriority, s.UsPonSTagPriority, s.DsPonCTagPriority, s.DsPonSTagPriority)

		if err != nil {
			oltLogger.WithFields(log.Fields{
				"Err": err.Error(),
			}).Fatal("Can't create Service")
		}

		uni.Services = append(uni.Services, service)
	}

	uni.PacketCh = make(chan bbsimTypes.OnuPacketMessage)

	return &uni, nil
}

func (u *UniPort) StorePortNo(portNo uint32) {
	u.PortNo = portNo
	u.logger.WithFields(log.Fields{
		"PortNo": portNo,
	}).Debug("logical-port-number-added-to-uni")
}

func (u *UniPort) UpdateStream(stream bbsimTypes.Stream) {
	for _, service := range u.Services {
		service.UpdateStream(stream)
	}
}

func (u *UniPort) Enable() error {
	return u.OperState.Event(uniTxEnable)
}

func (u *UniPort) Disable() error {
	if u.OperState.Is(UniStateDown) {
		return nil
	}
	return u.OperState.Event(uniTxDisable)
}

// this method simply forwards the packet to the correct service
func (u *UniPort) HandlePackets() {
	u.logger.Debug("listening-on-uni-packet-channel")

	defer func() {
		u.logger.Debug("done-listening-on-uni-packet-channel")
	}()

	for msg := range u.PacketCh {
		u.logger.WithFields(log.Fields{
			"messageType": msg.Type,
		}).Trace("received-message-on-uni-packet-channel")

		if msg.Type == packetHandlers.EAPOL || msg.Type == packetHandlers.DHCP {
			service, err := u.findServiceByMacAddress(msg.MacAddress)
			if err != nil {
				u.logger.WithFields(log.Fields{"err": err}).Error("cannot-process-uni-pkt")
				continue
			}
			service.PacketCh <- msg
		} else if msg.Type == packetHandlers.IGMP {
			//IGMP packets don't refer to any Mac Address, thus
			//if it's an IGMP packet we assume we have a single IGMP service
			for _, s := range u.Services {
				service := s.(*Service)

				if service.NeedsIgmp {
					service.PacketCh <- msg
				}
			}
		}
	}
}

func (u *UniPort) HandleAuth() {
	for _, s := range u.Services {
		s.HandleAuth()
	}
}

func (u *UniPort) HandleDhcp(pbit uint8, cTag int) {
	for _, s := range u.Services {
		s.HandleDhcp(pbit, cTag)
	}
}

func (u *UniPort) addGemPortToService(gemport uint32, ethType uint32, oVlan uint32, iVlan uint32) {
	for _, s := range u.Services {
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

			// for loggin purpose only
			if service.GemPort == gemport {
				// TODO move to Trace level
				u.logger.WithFields(log.Fields{
					"OnuId":  service.UniPort.Onu.ID,
					"IntfId": service.UniPort.Onu.PonPortID,
					"OnuSn":  service.UniPort.Onu.Sn(),
					"Name":   service.Name,
					"PortNo": service.UniPort.PortNo,
					"UniId":  service.UniPort.ID,
				}).Debug("gem-port-added-to-service")
			}
		}
	}
}

func (u *UniPort) findServiceByMacAddress(macAddress net.HardwareAddr) (*Service, error) {
	for _, s := range u.Services {
		service := s.(*Service)
		if service.HwAddress.String() == macAddress.String() {
			return service, nil
		}
	}
	return nil, fmt.Errorf("cannot-find-service-with-mac-address-%s", macAddress.String())
}
