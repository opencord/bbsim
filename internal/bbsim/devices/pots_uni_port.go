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

	"github.com/looplab/fsm"
	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	log "github.com/sirupsen/logrus"
)

var potsUniLogger = log.WithFields(log.Fields{
	"module": "POTS",
})

const (
	PotsUniStateUp   = "up"
	PotsUniStateDown = "down"

	potsUniTxEnable  = "enable"
	potsUniTxDisable = "disable"
)

type PotsUniPortIf interface {
	Enable() error
	Disable() error
}

type PotsUniPort struct {
	ID        uint32
	MeId      omcilib.EntityID
	PortNo    uint32
	OperState *fsm.FSM
	Onu       *Onu
	logger    *log.Entry
}

func NewPotsUniPort(ID uint32, onu *Onu) (*PotsUniPort, error) {
	uni := PotsUniPort{
		ID:   ID,
		Onu:  onu,
		MeId: omcilib.GenerateUniPortEntityId(ID + 1),
	}

	uni.logger = potsUniLogger.WithFields(log.Fields{
		"PotsUniId": uni.ID,
		"OnuSn":     onu.Sn(),
	})

	uni.OperState = fsm.NewFSM(
		PotsUniStateDown,
		fsm.Events{
			{Name: potsUniTxEnable, Src: []string{PotsUniStateDown}, Dst: PotsUniStateUp},
			{Name: potsUniTxDisable, Src: []string{PotsUniStateUp}, Dst: PotsUniStateDown},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				uni.logger.Debugf("changing-pots-uni-operstate-from-%s-to-%s", e.Src, e.Dst)
			},
			fmt.Sprintf("enter_%s", PotsUniStateUp): func(e *fsm.Event) {
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
			},
			fmt.Sprintf("enter_%s", PotsUniStateDown): func(e *fsm.Event) {
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
			},
		},
	)

	return &uni, nil
}

func (u *PotsUniPort) StorePortNo(portNo uint32) {
	u.PortNo = portNo
	u.logger.WithFields(log.Fields{
		"PortNo": portNo,
	}).Debug("logical-port-number-added-to-uni")
}

func (u *PotsUniPort) Enable() error {
	return u.OperState.Event(uniTxEnable)
}

func (u *PotsUniPort) Disable() error {
	if u.OperState.Is(PotsUniStateDown) {
		return nil
	}
	return u.OperState.Event(potsUniTxDisable)
}
