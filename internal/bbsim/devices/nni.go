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
	"github.com/google/gopacket"
	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/bbsim/packetHandlers"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	log "github.com/sirupsen/logrus"
)

var nniLogger = log.WithFields(log.Fields{"module": "NNI"})

type NniPort struct {
	// BBSIM Internals
	ID  uint32
	Olt *OltDevice

	// PON Attributes
	OperState   *fsm.FSM
	Type        string
	PacketCount uint64 // dummy value for the stats
}

func CreateNNI(olt *OltDevice) (NniPort, error) {
	nniPort := NniPort{
		ID: uint32(0),
		OperState: getOperStateFSM(func(e *fsm.Event) {
			oltLogger.Debugf("Changing NNI OperState from %s to %s", e.Src, e.Dst)
		}),
		Type: "nni",
		Olt:  olt,
	}

	return nniPort, nil
}

// handleNniPacket will send a packet to a fake DHCP server implementation
func (n *NniPort) handleNniPacket(packet gopacket.Packet) error {
	isDhcp := packetHandlers.IsDhcpPacket(packet)
	isLldp := packetHandlers.IsLldpPacket(packet)
	isIcmp := packetHandlers.IsIcmpPacket(packet)

	if !isDhcp && !isLldp && !isIcmp {
		nniLogger.WithFields(log.Fields{
			"packet": packet,
		}).Trace("Dropping NNI packet as it's not DHCP")
		return nil
	}

	if isDhcp {

		// get a response packet from the DHCP server
		pkt, err := n.Olt.dhcpServer.HandleServerPacket(packet)
		if err != nil {
			nniLogger.WithFields(log.Fields{
				"SourcePkt": hex.EncodeToString(packet.Data()),
				"Err":       err,
			}).Error("DHCP Server can't handle packet")
			return err
		}

		// send packetIndication to VOLTHA
		data := &openolt.Indication_PktInd{PktInd: &openolt.PacketIndication{
			IntfType: "nni",
			IntfId:   n.ID,
			Pkt:      pkt.Data()}}
		if err := n.Olt.OpenoltStream.Send(&openolt.Indication{Data: data}); err != nil {
			oltLogger.WithFields(log.Fields{
				"IntfType": data.PktInd.IntfType,
				"IntfId":   n.ID,
				"Pkt":      hex.EncodeToString(pkt.Data()),
			}).Errorf("Fail to send PktInd indication: %v", err)
			return err
		}
	} else if isLldp {
		// TODO rework this when BBSim supports data-plane packets
		nniLogger.Trace("Received LLDP Packet, ignoring it")
	} else if isIcmp {
		nniLogger.Trace("Received ICMP Packet, ignoring it")
	}
	return nil
}
