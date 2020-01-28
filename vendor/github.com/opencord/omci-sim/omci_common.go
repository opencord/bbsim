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
	"fmt"
	log "github.com/sirupsen/logrus"
)

type OmciError struct {
	Msg string
}

func (e *OmciError) Error() string {
	return fmt.Sprintf("%s", e.Msg)
}

type OnuKey struct {
	IntfId, OnuId uint32
}

func (k OnuKey) String() string {
	return fmt.Sprintf("Onu {intfid:%d, onuid:%d}", k.IntfId, k.OnuId)
}

func GetAttributes(class OmciClass, content OmciContent, key OnuKey, pkt []byte) []byte {
	log.WithFields(log.Fields{
		"IntfId": key.IntfId,
		"OnuId": key.OnuId,
	}).Tracef("GetAttributes() invoked")

	switch class {
	case ANIG:
		pos := uint(11)
		pkt, _ = GetANIGAttributes(&pos, pkt, content)
		return pkt

	case SoftwareImage:
		pos := uint(11)
		pkt, _ = GetSoftwareImageAttributes(&pos, pkt, content)
		return pkt

	case ONUG:
		pos := uint(11)
		pkt, _ = GetOnuGAttributes(&pos, pkt, content)
		return pkt

	case ONU2G:
		pos := uint(11)
		pkt, _ = GetOnu2GAttributes(&pos, pkt, content)
		return pkt

	case EthernetPMHistoryData:
		pos := uint(11)
		pkt, _ = GetEthernetPMHistoryDataAttributes(&pos, pkt, content)
		return pkt

	default:
		// For unimplemented MEs, just fill in the attribute mask and return 0 values for the requested attributes
		// TODO implement Get for unimplemented MEs as well
		log.WithFields(log.Fields{
			"IntfId": key.IntfId,
			"OnuId": key.OnuId,
			"class": class,
		}).Tracef("Unimplemeted GetAttributes for ME Class: %v " +
		    "Filling with zero value for the requested attributes", class)
		AttributesMask := getAttributeMask(content)
		pkt[8] = 0x00 // Command Processed Successfully
		pkt[9] = uint8(AttributesMask >> 8)
		pkt[10] = uint8(AttributesMask & 0xFF)

		return pkt
	}
}

func getAttributeMask(content OmciContent) int {
	// mask is present in pkt[8] and pkt[9]
	log.WithFields(log.Fields{
		"content[0]": content[0],
		"content[1]": content[1],
	}).Tracef("GetAttributeMask() invoked")
	return (int(content[0]) << 8) | int(content[1])
}
