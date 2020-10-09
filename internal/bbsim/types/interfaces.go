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

package types

import (
	"encoding/hex"
	"github.com/opencord/voltha-protos/v3/go/openolt"
	log "github.com/sirupsen/logrus"
	"time"
)

// represent a gRPC stream
type Stream interface {
	Send(*openolt.Indication) error
}

// PacketStream is responsible to send packet-ins to the openolt adapter
// and ensure that no more than a certain number of packets per second are sent
// This is necessary to avoid overloading the radius and dhcp servers
type PacketStream struct {
	grpcStream          Stream
	maxPacketsPerSecond int
	// this channel will hold all the packets that needs to be sent out
	pktCh chan *openolt.Indication
	// this channel will drive the processing
	schedulerCh chan time.Time
}

func (p *PacketStream) Send(indication *openolt.Indication) error {
	switch indication.Data.(type) {
	case *openolt.Indication_PktInd:
		p.pktCh <- indication
		return nil
	default:
		return p.grpcStream.Send(indication)
	}
}

func (p *PacketStream) handlePackets() {
	// every second make sure to refill the available spots in the queue
	go func() {
		for t := range time.Tick(1 * time.Second) {
			for i := 0; i < p.maxPacketsPerSecond; i++ {
				p.schedulerCh <- t
			}
		}
	}()

	for pkt := range p.pktCh {
		// wait to have availability in the scheduler
		<-p.schedulerCh

		pktIndication := pkt.GetPktInd()
		// send the packet
		if err := p.grpcStream.Send(pkt); err != nil {
			log.WithFields(log.Fields{
				"IntfId": pktIndication.IntfId,
				"PortNo": pktIndication.PortNo,
				"Pkt":    hex.EncodeToString(pktIndication.Pkt),
			}).Error("Failed to send packet on the stream")
		}
		log.WithFields(log.Fields{
			"IntfId": pktIndication.IntfId,
			"PortNo": pktIndication.PortNo,
			"Pkt":    hex.EncodeToString(pktIndication.Pkt),
		}).Trace("Successfully send packet on the stream")
	}
}

func NewPacketStream(grpcStream Stream, maxPacketsPerSecond int) *PacketStream {
	s := &PacketStream{
		grpcStream:          grpcStream,
		maxPacketsPerSecond: maxPacketsPerSecond,
		pktCh:               make(chan *openolt.Indication, 2000),
		schedulerCh:         make(chan time.Time, maxPacketsPerSecond),
	}
	go s.handlePackets()
	return s
}
