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

type OnuGAttributes int

const (
	_                                       = iota
	VendorID                 OnuGAttributes = 0x8000
	Version                  OnuGAttributes = 0x4000
	SerialNumber             OnuGAttributes = 0x2000
	TrafficManagementOptions OnuGAttributes = 0x1000
	VpVcCrossConnectOptions  OnuGAttributes = 0x0800
	BatteryBackup            OnuGAttributes = 0x0400
	AdministrativeState      OnuGAttributes = 0x0200
	OperationalState         OnuGAttributes = 0x0100
	OntSurvivalTime          OnuGAttributes = 0x0080
	LogicalOnuID             OnuGAttributes = 0x0040
	LogicalPassword          OnuGAttributes = 0x0020
	CredentialsStatus        OnuGAttributes = 0x0010
	ExtendedTcLayerOptions   OnuGAttributes = 0x0008
)

type OnuGAttributeHandler func(*uint, []byte, OnuKey) ([]byte, error)

var OnuGAttributeHandlers = map[OnuGAttributes]OnuGAttributeHandler{
	VendorID:                 GetVendorID,
	Version:                  GetVersion,
	SerialNumber:             GetSerialNumber,
	TrafficManagementOptions: GetTrafficManagementOptions,
	VpVcCrossConnectOptions:  GetVpVcCrossConnectOptions,
	BatteryBackup:            GetBatteryBackup,
	AdministrativeState:      GetAdministrativeState,
	OperationalState:         GetOperationalState,
	OntSurvivalTime:          GetOntSurvivalTime,
	LogicalOnuID:             GetLogicalOnuID,
	LogicalPassword:          GetLogicalPassword,
	CredentialsStatus:        GetCredentialsStatus,
	ExtendedTcLayerOptions:   GetExtendedTcLayerOptions,
}

func GetOnuGAttributes(pos *uint, pkt []byte, content OmciContent, key OnuKey) ([]byte, error) {
	AttributesMask := getAttributeMask(content)

	for index := uint(16); index >= 1; index-- {
		Attribute := 1 << (index - 1)
		reqAttribute := Attribute & AttributesMask

		if reqAttribute != 0 {
			pkt, _ = OnuGAttributeHandlers[OnuGAttributes(reqAttribute)](pos, pkt, key)
		}
	}

	pkt[8] = 0x00 // Command Processed Successfully
	pkt[9] = uint8(AttributesMask >> 8)
	pkt[10] = uint8(AttributesMask & 0x00FF)

	return pkt, nil

}

func GetVendorID(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 4 bytes
	vendorid := []byte("BBSM")
	for _, ch := range vendorid {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetVersion(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 14 bytes
	for i := 1; i <= 14; i++ {
		b := byte(' ')
		pkt[*pos] = b
		*pos++
	}
	return pkt, nil
}

func GetSerialNumber(pos *uint, pkt []byte, key OnuKey) ([]byte, error) {
	// 8 bytes
	vendorid := []byte("BBSM")
	serialhex := []byte{0x00, byte(key.OltId % 256), byte(key.IntfId), byte(key.OnuId)}
	serialnumber := append(vendorid, serialhex...)
	for _, ch := range serialnumber {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}

func GetTrafficManagementOptions(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetVpVcCrossConnectOptions(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetBatteryBackup(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetAdministrativeState(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetOperationalState(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetOntSurvivalTime(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetLogicalOnuID(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 24 bytes
	for i := 1; i <= 24; i++ {
		b := byte(' ')
		pkt[*pos] = b
		*pos++
	}
	return pkt, nil
}

func GetLogicalPassword(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 24 bytes
	for i := 1; i <= 24; i++ {
		b := byte(' ')
		pkt[*pos] = b
		*pos++
	}
	return pkt, nil
}

func GetCredentialsStatus(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 1 byte
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetExtendedTcLayerOptions(pos *uint, pkt []byte, _ OnuKey) ([]byte, error) {
	// 2 bytes
	tcbits := []byte{0x00, 0x00}
	for _, ch := range tcbits {
		pkt[*pos] = ch
		*pos++
	}
	return pkt, nil
}
