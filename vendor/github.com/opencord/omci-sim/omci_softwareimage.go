/*
 * Copyright 2020-present Open Networking Foundation

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

type SoftwareImageAttributes int

const (
	_                                       = iota
	SoftwareVersion SoftwareImageAttributes = 0x8000
	IsCommited      SoftwareImageAttributes = 0x4000
	IsActive        SoftwareImageAttributes = 0x2000
	IsValid         SoftwareImageAttributes = 0x1000
	ProductCode     SoftwareImageAttributes = 0x0800
	ImageHash       SoftwareImageAttributes = 0x0400
)

type SoftwareImageAttributeHandler func(*uint, []byte) ([]byte, error)

var SoftwareImageAttributeHandlers = map[SoftwareImageAttributes]SoftwareImageAttributeHandler{
	SoftwareVersion: GetSoftwareVersion,
	IsCommited:      GetIsCommited,
	IsActive:        GetIsActive,
	IsValid:         GetIsValid,
	ProductCode:     GetProductCode,
	ImageHash:       GetImageHash,
}

func GetSoftwareImageAttributes(pos *uint, pkt []byte, content OmciContent) ([]byte, error) {
	AttributesMask := getAttributeMask(content)

	for index := uint(16); index >= 1; index-- {
		Attribute := 1 << (index - 1)
		reqAttribute := Attribute & AttributesMask

		if reqAttribute != 0 {
			pkt, _ = SoftwareImageAttributeHandlers[SoftwareImageAttributes(reqAttribute)](pos, pkt)
		}
	}

	pkt[8] = 0x00 // Command Processed Successfully
	pkt[9] = uint8(AttributesMask >> 8)
	pkt[10] = uint8(AttributesMask & 0x00FF)

	return pkt, nil

}

func GetSoftwareVersion(pos *uint, pkt []byte) ([]byte, error) {
	// 14 bytes
	version := []byte("00000000000001")
	for _, ch := range version {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetIsCommited(pos *uint, pkt []byte) ([]byte, error) {
	// 1 bytes
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetIsActive(pos *uint, pkt []byte) ([]byte, error) {
	// 1 bytes
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetIsValid(pos *uint, pkt []byte) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetProductCode(pos *uint, pkt []byte) ([]byte, error) {
	// 25 bytes
	// BRCM has 25 nulls
	for i := 1; i <= 25; i++ {
		pkt[*pos] = 0x00
		*pos++
	}
	return pkt, nil
}

func GetImageHash(pos *uint, pkt []byte) ([]byte, error) {
	// 16 bytes
	// BRCM has 16 nulls
	for i := 1; i <= 16; i++ {
		pkt[*pos] = 0x00
		*pos++
	}
	return pkt, nil
}
