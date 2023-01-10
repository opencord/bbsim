/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

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

var potsLogger = log.WithFields(log.Fields{
	"module": "POTS",
})

const (
	PotsStateUp   = "up"
	PotsStateDown = "down"

	potsTxEnable  = "enable"
	potsTxDisable = "disable"
)

type PotsPortIf interface {
	GetID() uint32
	Enable() error
	Disable() error
}

type PotsPort struct {
	ID        uint32
	MeId      omcilib.EntityID
	PortNo    uint32
	OperState *fsm.FSM
	Onu       *Onu
	logger    *log.Entry
}

func NewPotsPort(ID uint32, onu *Onu) (*PotsPort, error) {
	pots := PotsPort{
		ID:   ID,
		Onu:  onu,
		MeId: omcilib.GenerateUniPortEntityId(ID + 1),
	}

	pots.logger = potsLogger.WithFields(log.Fields{
		"PotsUniId": pots.ID,
		"OnuSn":     onu.Sn(),
	})

	pots.OperState = fsm.NewFSM(
		PotsStateDown,
		fsm.Events{
			{Name: potsTxEnable, Src: []string{PotsStateDown}, Dst: PotsStateUp},
			{Name: potsTxDisable, Src: []string{PotsStateUp}, Dst: PotsStateDown},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				pots.logger.Debugf("changing-pots-operstate-from-%s-to-%s", e.Src, e.Dst)
			},
			fmt.Sprintf("enter_%s", PotsStateUp): func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.UniStatusAlarm,
					Data: bbsimTypes.UniStatusAlarmMessage{
						OnuSN:          pots.Onu.SerialNumber,
						OnuID:          pots.Onu.ID,
						AdminState:     0,
						EntityID:       pots.MeId.ToUint16(),
						RaiseOMCIAlarm: false, // never raise an LOS when enabling a UNI
					},
				}
				pots.Onu.Channel <- msg
			},
			fmt.Sprintf("enter_%s", PotsStateDown): func(e *fsm.Event) {
				msg := bbsimTypes.Message{
					Type: bbsimTypes.UniStatusAlarm,
					Data: bbsimTypes.UniStatusAlarmMessage{
						OnuSN:          pots.Onu.SerialNumber,
						OnuID:          pots.Onu.ID,
						AdminState:     1,
						EntityID:       pots.MeId.ToUint16(),
						RaiseOMCIAlarm: true, // raise an LOS when disabling a UNI
					},
				}
				pots.Onu.Channel <- msg
			},
		},
	)

	return &pots, nil
}

func (p *PotsPort) GetID() uint32 {
	return p.ID
}

func (p *PotsPort) StorePortNo(portNo uint32) {
	p.PortNo = portNo
	p.logger.WithFields(log.Fields{
		"PortNo": portNo,
	}).Debug("logical-port-number-added-to-pots")
}

func (p *PotsPort) Enable() error {
	return p.OperState.Event(potsTxEnable)
}

func (p *PotsPort) Disable() error {
	if p.OperState.Is(PotsStateDown) {
		return nil
	}
	return p.OperState.Event(potsTxDisable)
}
