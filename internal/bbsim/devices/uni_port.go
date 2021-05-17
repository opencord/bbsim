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
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	log "github.com/sirupsen/logrus"
)

const maxUniPorts = 4

type UniPort struct {
	ID        uint32
	MeId      omcilib.EntityID
	OperState *fsm.FSM
	Onu       *Onu
}

func NewUniPort(ID uint32, onu *Onu) (*UniPort, error) {

	// IDs starts from 0, thus the maximum UNI supported is maxUniPorts - 1
	if ID > (maxUniPorts - 1) {
		return nil, fmt.Errorf("%d-is-higher-than-the-maximum-supported-unis-%d", ID, maxUniPorts)
	}

	uni := UniPort{
		ID:   ID,
		Onu:  onu,
		MeId: omcilib.GenerateUniPortEntityId(ID + 1),
	}

	uni.OperState = getOperStateFSM(func(e *fsm.Event) {
		onuLogger.WithFields(log.Fields{
			"ID":    uni.ID,
			"OnuSn": onu.Sn(),
		}).Debugf("changing-uni-operstate-from-%s-to-%s", e.Src, e.Dst)
	})

	return &uni, nil
}
