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

package bbr

import (
	"context"
	"github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	//omcisim "github.com/opencord/omci-sim"
	pb "github.com/opencord/voltha-protos/go/openolt"
	log "github.com/sirupsen/logrus"
)

type ONUMock struct {
	OmciChannel  chan *pb.OmciIndication
	SerialNumber *pb.SerialNumber
	IntfId       uint32
	ID           uint32

	// OMCI params
	tid   uint16
	hpTid uint16
}

func GetSN(oltid int, intfid uint32, onuid uint32) *pb.SerialNumber {

	sn := new(pb.SerialNumber)

	//sn = new(openolt.SerialNumber)
	sn.VendorId = []byte("BBSM")
	sn.VendorSpecific = []byte{0, byte(oltid % 256), byte(intfid), byte(onuid)}

	return sn
}

func (onu *ONUMock) getNextTid(highPriority ...bool) uint16 {
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

func (o *ONUMock) startOmci(client pb.OpenoltClient) {
	log.WithFields(log.Fields{
		"IntfId":       o.IntfId,
		"SerialNumber": common.OnuSnToString(o.SerialNumber),
		"OnuId":        o.ID,
	}).Info("sending MibReset")

	mibReset, _ := omcilib.CreateMibResetRequest(o.getNextTid(false))

	omciMsg := pb.OmciMsg{
		IntfId: o.IntfId,
		OnuId:  o.ID,
		Pkt:    mibReset,
	}

	if _, err := client.OmciMsgOut(context.Background(), &omciMsg); err != nil {
		log.WithFields(log.Fields{
			"IntfId":       o.IntfId,
			"SerialNumber": common.OnuSnToString(o.SerialNumber),
		}).Error("Failed to send MIB Reset")
	}

	go func() {
		// TODO close the channel once done
		for msg := range o.OmciChannel {
			log.WithFields(log.Fields{
				"IntfId": msg.IntfId,
				"OnuId":  msg.OnuId,
				"OnuSn":  common.OnuSnToString(o.SerialNumber),
				"Pkt":    msg.Pkt,
			}).Debug("ONU Receveives OMCI Msg")

			// TODO implement an OMCI responder
		}
	}()
}
