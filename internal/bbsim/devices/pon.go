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
	"bytes"
	"errors"
	"fmt"
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
)

type PonPort struct {
	// BBSIM Internals
	ID     uint32
	NumOnu int
	Onus   []*Onu
	Olt    OltDevice

	// PON Attributes
	OperState *fsm.FSM
	Type      string

	// NOTE do we need a state machine for the PON Ports?
}

func (p PonPort) GetOnuBySn(sn *openolt.SerialNumber) (*Onu, error) {
	for _, onu := range p.Onus {
		if bytes.Equal(onu.SerialNumber.VendorSpecific, sn.VendorSpecific) {
			return onu, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find Onu with serial number %d in PonPort %d", sn, p.ID))
}

func (p PonPort) GetOnuById(id uint32) (*Onu, error) {
	for _, onu := range p.Onus {
		if onu.ID == id {
			return onu, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Cannot find Onu with id %d in PonPort %d", id, p.ID))
}
