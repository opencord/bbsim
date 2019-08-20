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

package responders

import (
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	omci "github.com/opencord/omci-sim"
	"github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
	"net"
)

var eapolLogger = log.WithFields(log.Fields{
	"module": "EAPOL",
})

func StartWPASupplicant(onuId uint32, ponPortId uint32, serialNumber *openolt.SerialNumber, stream openolt.Openolt_EnableIndicationServer, pktOutCh chan *bbsim.ByteMsg)  {
	// NOTE pckOutCh is channel to listen on for packets received by VOLTHA
	// the ONU device will publish messages on that channel

	eapolLogger.WithFields(log.Fields{
		"OnuId": onuId,
		"IntfId": ponPortId,
		"OnuSn": serialNumber,
	}).Infof("EAPOL State Machine starting")

	// send the packet (hacked together)
	gemid, err := omci.GetGemPortId(ponPortId, onuId)
	if err != nil {
		eapolLogger.WithFields(log.Fields{
			"OnuId": onuId,
			"IntfId": ponPortId,
			"OnuSn": serialNumber,
		}).Errorf("Can't retrieve GemPortId: %s", err)
		return
	}

	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, 0x07, byte(onuId)},
		DstMAC:       net.HardwareAddr{0x01, 0x80, 0xC2, 0x00, 0x00, 0x03},
		EthernetType: layers.EthernetTypeEAPOL,
	}

	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		&layers.EAPOL{Version: 1, Type: 1, Length: 0},
	)

	msg := buffer.Bytes()

	data := &openolt.Indication_PktInd{
		PktInd: &openolt.PacketIndication{
			IntfType: "pon",
			IntfId: ponPortId,
			GemportId: uint32(gemid),
			Pkt: msg,
		},
	}
	// end of hacked (move in an EAPOL state machine)
	if err := stream.Send(&openolt.Indication{Data: data}); err != nil {
		eapolLogger.Error("Fail to send EAPOL PktInd indication. %v", err)
	}
}