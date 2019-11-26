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
package core

import (
	"errors"
	"fmt"
	"sync"
	log "github.com/sirupsen/logrus"
)

type OnuOmciState struct {
	gemPortId         uint16
	mibUploadCtr      uint16
	extraMibUploadCtr uint16 // this is only for debug purposes, will be removed in the future
	uniGInstance      uint8
	tcontInstance     uint8
	pptpInstance      uint8
	priorQInstance    uint8 // To assign incrementing value to PQ instance-Id
	priorQPriority	  uint8 // Priority of the PriorityQueueG (0-7)
	tcontPointer      uint8 // Tcont Pointer for PriorQ
	state             istate
}

type istate int

// TODO - Needs to reflect real ONU/OMCI state
const (
	INCOMPLETE istate = iota
	DONE
)

var OnuOmciStateMap = map[OnuKey]*OnuOmciState{}
var OnuOmciStateMapLock = sync.RWMutex{}

func NewOnuOmciState() *OnuOmciState {
	return &OnuOmciState{gemPortId: 0, mibUploadCtr: 0, uniGInstance: 1, tcontInstance: 0, pptpInstance: 1}
}
func (s *OnuOmciState) ResetOnuOmciState() {
	// Resetting the counters  
	s.mibUploadCtr = 0
	s.extraMibUploadCtr = 0
	s.gemPortId = 0
	s.uniGInstance = 1
	s.tcontInstance = 0
	s.pptpInstance = 1
	s.tcontPointer = 0
	s.priorQPriority = 0
}
func GetOnuOmciState(intfId uint32, onuId uint32) istate {
	key := OnuKey{intfId, onuId}
	OnuOmciStateMapLock.RLock()
	defer OnuOmciStateMapLock.RUnlock()
	if onu, ok := OnuOmciStateMap[key]; ok {
		return onu.state
	} else {
		return INCOMPLETE
	}
}

func GetGemPortId(intfId uint32, onuId uint32) (uint16, error) {
	key := OnuKey{intfId, onuId}
	OnuOmciStateMapLock.RLock()
	defer OnuOmciStateMapLock.RUnlock()
	if OnuOmciState, ok := OnuOmciStateMap[key]; ok {
		if OnuOmciState.state != DONE {
			errmsg := fmt.Sprintf("ONU {intfid:%d, onuid:%d} - Not DONE (GemportID is not set)", intfId, onuId)
			return 0, errors.New(errmsg)
		}
		return OnuOmciState.gemPortId, nil
	}
	errmsg := fmt.Sprintf("ONU {intfid:%d, onuid:%d} - Failed to find a key in OnuOmciStateMap", intfId, onuId)
	return 0, errors.New(errmsg)
}

func CheckIsTeo() string {
	log.Warn("It's TEO!")
	return "It's TEO!"
}
