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
	"bytes"
	"fmt"
	"sync"

	"github.com/looplab/fsm"
	"github.com/opencord/bbsim/internal/common"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	log "github.com/sirupsen/logrus"
)

var ponLogger = log.WithFields(log.Fields{
	"module": "PON",
})

type AllocIDVal struct {
	OnuSn   *openolt.SerialNumber
	AllocID uint16
}

type AllocIDKey struct {
	PonID    uint32
	OnuID    uint32
	EntityID uint16
}

type PonPort struct {
	// BBSIM Internals
	ID            uint32
	Technology    common.PonTechnology
	NumOnu        int
	Onus          []*Onu
	Olt           *OltDevice
	PacketCount   uint64
	InternalState *fsm.FSM

	// PON Attributes
	OperState *fsm.FSM
	Type      string

	// Allocated resources
	// Some resources (eg: OnuId, AllocId and GemPorts) have to be unique per PON port
	// we are keeping a list so that we can throw an error in cases we receive duplicates
	AllocatedGemPorts     map[uint16]*openolt.SerialNumber
	allocatedGemPortsLock sync.RWMutex
	AllocatedOnuIds       map[uint32]*openolt.SerialNumber
	allocatedOnuIdsLock   sync.RWMutex
	AllocatedAllocIds     map[AllocIDKey]*AllocIDVal // key is AllocIDKey
	allocatedAllocIdsLock sync.RWMutex
}

// CreatePonPort creates pon port object
func CreatePonPort(olt *OltDevice, id uint32, tech common.PonTechnology) *PonPort {
	ponPort := PonPort{
		NumOnu:            olt.NumOnuPerPon,
		ID:                id,
		Technology:        tech,
		Type:              "pon",
		Olt:               olt,
		Onus:              []*Onu{},
		AllocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
		AllocatedOnuIds:   make(map[uint32]*openolt.SerialNumber),
		AllocatedAllocIds: make(map[AllocIDKey]*AllocIDVal),
	}

	ponPort.InternalState = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: "enable", Src: []string{"created", "disabled"}, Dst: "enabled"},
			{Name: "disable", Src: []string{"enabled"}, Dst: "disabled"},
		},
		fsm.Callbacks{
			"enter_enabled": func(e *fsm.Event) {
				ponLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port InternalState from %s to %s", e.Src, e.Dst)

				if e.Src == "created" {
					if olt.ControlledActivation == Default || olt.ControlledActivation == OnlyPON {
						for _, onu := range ponPort.Onus {
							if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
								ponLogger.Errorf("Error initializing ONU: %v", err)
								continue
							}
							if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
								ponLogger.Errorf("Error discover ONU: %v", err)
							}
						}
					}
				} else if e.Src == "disabled" {
					if ponPort.Olt.ControlledActivation == OnlyONU || ponPort.Olt.ControlledActivation == Both {
						// if ONUs are manually activated then only initialize them
						for _, onu := range ponPort.Onus {
							if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
								ponLogger.WithFields(log.Fields{
									"Err":    err,
									"OnuSn":  onu.Sn(),
									"IntfId": onu.PonPortID,
								}).Error("Error initializing ONU")
								continue
							}
						}
					} else {
						for _, onu := range ponPort.Onus {
							if onu.InternalState.Current() == OnuStatePonDisabled {
								if err := onu.InternalState.Event(OnuTxEnable); err != nil {
									ponLogger.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error enabling ONU")
								}
							} else if onu.InternalState.Current() == OnuStateDisabled {
								if err := onu.InternalState.Event(OnuTxInitialize); err != nil {
									ponLogger.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error initializing ONU")
									continue
								}
								if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
									ponLogger.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error discovering ONU")
								}
							} else if onu.InternalState.Current() == OnuStateInitialized {
								if err := onu.InternalState.Event(OnuTxDiscover); err != nil {
									ponLogger.WithFields(log.Fields{
										"Err":    err,
										"OnuSn":  onu.Sn(),
										"IntfId": onu.PonPortID,
									}).Error("Error discovering ONU")
								}
							} else {
								// this is to loudly report unexpected states in order to address them
								ponLogger.WithFields(log.Fields{
									"OnuSn":         onu.Sn(),
									"IntfId":        onu.PonPortID,
									"InternalState": onu.InternalState.Current(),
								}).Error("Unexpected ONU state in PON enabling")
							}
						}
					}
				}
			},
			"enter_disabled": func(e *fsm.Event) {
				for _, onu := range ponPort.Onus {
					if onu.InternalState.Current() == OnuStateInitialized || onu.InternalState.Current() == OnuStateDisabled {
						continue
					}
					if err := onu.InternalState.Event(OnuTxPonDisable); err != nil {
						ponLogger.Errorf("Failed to move ONU in %s states: %v", OnuStatePonDisabled, err)
					}
				}
			},
		},
	)

	ponPort.OperState = fsm.NewFSM(
		"down",
		fsm.Events{
			{Name: "enable", Src: []string{"down"}, Dst: "up"},
			{Name: "disable", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"enter_up": func(e *fsm.Event) {
				ponLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
				olt.sendPonIndication(ponPort.ID)
			},
			"enter_down": func(e *fsm.Event) {
				ponLogger.WithFields(log.Fields{
					"ID": ponPort.ID,
				}).Debugf("Changing PON Port OperState from %s to %s", e.Src, e.Dst)
				olt.sendPonIndication(ponPort.ID)
			},
		},
	)
	return &ponPort
}

func (p *PonPort) GetOnuBySn(sn *openolt.SerialNumber) (*Onu, error) {
	for _, onu := range p.Onus {
		if bytes.Equal(onu.SerialNumber.VendorSpecific, sn.VendorSpecific) {
			return onu, nil
		}
	}
	return nil, fmt.Errorf("Cannot find Onu with serial number %d in PonPort %d", sn, p.ID)
}

func (p *PonPort) GetOnuById(id uint32) (*Onu, error) {
	for _, onu := range p.Onus {
		if onu.ID == id {
			return onu, nil
		}
	}
	return nil, fmt.Errorf("Cannot find Onu with id %d in PonPort %d", id, p.ID)
}

// GetNumOfActiveOnus returns number of active ONUs for PON port
func (p *PonPort) GetNumOfActiveOnus() uint32 {
	var count uint32 = 0
	for _, onu := range p.Onus {
		if onu.InternalState.Current() == OnuStateInitialized || onu.InternalState.Current() == OnuStateCreated || onu.InternalState.Current() == OnuStateDisabled {
			continue
		}
		count++
	}
	return count
}

// storeOnuId adds the Id to the ONU Ids already allocated to this PON port
func (p *PonPort) storeOnuId(onuId uint32, onuSn *openolt.SerialNumber) {
	p.allocatedOnuIdsLock.Lock()
	defer p.allocatedOnuIdsLock.Unlock()
	p.AllocatedOnuIds[onuId] = onuSn
}

// removeOnuId removes the OnuId from the allocated resources
func (p *PonPort) removeOnuId(onuId uint32) {
	p.allocatedOnuIdsLock.Lock()
	defer p.allocatedOnuIdsLock.Unlock()
	delete(p.AllocatedOnuIds, onuId)
}

func (p *PonPort) removeAllOnuIds() {
	p.allocatedOnuIdsLock.Lock()
	defer p.allocatedOnuIdsLock.Unlock()
	p.AllocatedOnuIds = make(map[uint32]*openolt.SerialNumber)
}

// isOnuIdAllocated returns whether this OnuId is already in use on this PON
func (p *PonPort) isOnuIdAllocated(onuId uint32) (bool, *openolt.SerialNumber) {
	p.allocatedOnuIdsLock.RLock()
	defer p.allocatedOnuIdsLock.RUnlock()

	if _, ok := p.AllocatedOnuIds[onuId]; ok {
		return true, p.AllocatedOnuIds[onuId]
	}
	return false, nil
}

// storeGemPort adds the gemPortId to the gemports already allocated to this PON port
func (p *PonPort) storeGemPort(gemPortId uint16, onuSn *openolt.SerialNumber) {
	p.allocatedGemPortsLock.Lock()
	defer p.allocatedGemPortsLock.Unlock()
	p.AllocatedGemPorts[gemPortId] = onuSn
}

// removeGemPort removes the gemPortId from the allocated resources
func (p *PonPort) removeGemPort(gemPortId uint16) {
	p.allocatedGemPortsLock.Lock()
	defer p.allocatedGemPortsLock.Unlock()
	delete(p.AllocatedGemPorts, gemPortId)
}

func (p *PonPort) removeGemPortBySn(onuSn *openolt.SerialNumber) {
	p.allocatedGemPortsLock.Lock()
	defer p.allocatedGemPortsLock.Unlock()
	for gemPort, sn := range p.AllocatedGemPorts {
		if sn == onuSn {
			delete(p.AllocatedGemPorts, gemPort)
		}
	}
}

func (p *PonPort) removeAllGemPorts() {
	p.allocatedGemPortsLock.Lock()
	defer p.allocatedGemPortsLock.Unlock()
	p.AllocatedGemPorts = make(map[uint16]*openolt.SerialNumber)
}

// isGemPortAllocated returns whether this gemPort is already in use on this PON
func (p *PonPort) isGemPortAllocated(gemPortId uint16) (bool, *openolt.SerialNumber) {
	p.allocatedGemPortsLock.RLock()
	defer p.allocatedGemPortsLock.RUnlock()

	if _, ok := p.AllocatedGemPorts[gemPortId]; ok {
		return true, p.AllocatedGemPorts[gemPortId]
	}
	return false, nil
}

// storeAllocId adds the Id to the ONU Ids already allocated to this PON port
func (p *PonPort) storeAllocId(ponID uint32, onuID uint32, entityID uint16, allocId uint16, onuSn *openolt.SerialNumber) {
	p.allocatedAllocIdsLock.Lock()
	defer p.allocatedAllocIdsLock.Unlock()
	p.AllocatedAllocIds[AllocIDKey{ponID, onuID, entityID}] = &AllocIDVal{onuSn, allocId}
}

// removeAllocId removes the AllocId from the allocated resources
func (p *PonPort) removeAllocId(ponID uint32, onuID uint32, entityID uint16) {
	p.allocatedAllocIdsLock.Lock()
	defer p.allocatedAllocIdsLock.Unlock()
	allocKey := AllocIDKey{ponID, onuID, entityID}
	delete(p.AllocatedAllocIds, allocKey)
}

// removeAllocIdsForOnuSn removes the all AllocIds for the given onu serial number
func (p *PonPort) removeAllocIdsForOnuSn(onuSn *openolt.SerialNumber) {
	p.allocatedAllocIdsLock.Lock()
	defer p.allocatedAllocIdsLock.Unlock()
	for id, allocObj := range p.AllocatedAllocIds {
		if onuSn == allocObj.OnuSn {
			delete(p.AllocatedAllocIds, id)
		}
	}
}

func (p *PonPort) removeAllAllocIds() {
	p.allocatedAllocIdsLock.Lock()
	defer p.allocatedAllocIdsLock.Unlock()
	p.AllocatedAllocIds = make(map[AllocIDKey]*AllocIDVal)
}

// isAllocIdAllocated returns whether this AllocId is already in use on this PON
func (p *PonPort) isAllocIdAllocated(ponID uint32, onuID uint32, entityID uint16) (bool, *AllocIDVal) {
	p.allocatedAllocIdsLock.RLock()
	defer p.allocatedAllocIdsLock.RUnlock()
	allocKey := AllocIDKey{ponID, onuID, entityID}
	if _, ok := p.AllocatedAllocIds[allocKey]; ok {
		return true, p.AllocatedAllocIds[allocKey]
	}
	return false, nil
}
