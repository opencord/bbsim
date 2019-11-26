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

type AniGAttributes int
const (
	_								= iota
	SRIndication  				AniGAttributes	= 0x8000
	TotalTcontNumber 			AniGAttributes	= 0x4000
	GEMBlockLength				AniGAttributes	= 0x2000
	PiggybackDBAReporting 		AniGAttributes	= 0x1000
	WholeONTDBAReporting		AniGAttributes	= 0x0800
	SFThreshold					AniGAttributes	= 0x0400
	SDThreshold					AniGAttributes	= 0x0200
	ARC							AniGAttributes	= 0x0100
	ARCInterval					AniGAttributes	= 0x0080
	OpticalSignalLevel			AniGAttributes	= 0x0040
	LowerOpticalThreshold		AniGAttributes	= 0x0020
	UpperOpticalThreshold		AniGAttributes	= 0x0010
	ONTResponseTime				AniGAttributes	= 0x0008
	TransmitOpticalLeval		AniGAttributes	= 0x0004
	LowerTransmitPowerThreshold	AniGAttributes	= 0x0002
	UpperTransmitPowerThreshold	AniGAttributes	= 0x0001
)

type ANIGAttributeHandler func(*uint, []byte) ([]byte, error)

var ANIGAttributeHandlers = map[AniGAttributes]ANIGAttributeHandler{
	SRIndication: GetSRIndication,
	OpticalSignalLevel: GetOpticalSignalLevel,
	LowerOpticalThreshold: GetLowerOpticalThreshold,
	UpperOpticalThreshold: GetUpperOpticalThreshold,
	TotalTcontNumber: GetTotalTcontNumber,
	GEMBlockLength: GetGEMBlockLength,
	PiggybackDBAReporting: GetPiggybackDBAReporting,
	WholeONTDBAReporting: GetWholeONTDBAReporting,
	SFThreshold: GetSFThreshold,
	SDThreshold: GetSDThreshold,
	ARC: GetARC,
	ARCInterval: GetARCInterval,
	ONTResponseTime: GetONTResponseTime,
	TransmitOpticalLeval: GetTransmitOpticalLeval,
	LowerTransmitPowerThreshold: GetLowerTransmitPowerThreshold,
	UpperTransmitPowerThreshold: GetUpperTransmitPowerThreshold,
}


func GetANIGAttributes(pos *uint, pkt []byte, content OmciContent) ([]byte, error) {
	AttributesMask := getAttributeMask(content)

	for index := uint(16); index>=1 ; index-- {
		Attribute := 1 << (index - 1)
		reqAttribute := Attribute & AttributesMask

		if reqAttribute != 0 {
			pkt, _ = ANIGAttributeHandlers[AniGAttributes(reqAttribute)](pos, pkt)
		}
	}

	pkt[8] = 0x00 // Command Processed Successfully
	pkt[9] = uint8(AttributesMask >> 8)
	pkt[10] = uint8(AttributesMask & 0x00FF)

	return pkt, nil

}


func GetSRIndication(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x01
	*pos++
	return pkt, nil
}

func GetOpticalSignalLevel(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0xd7
	*pos++
	pkt[*pos] = 0xa9
	*pos++
	return pkt, nil
}

func GetTotalTcontNumber(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x08
	*pos++
	return pkt, nil
}

func GetGEMBlockLength(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	pkt[*pos] = 0x30
	return pkt, nil
}

func GetPiggybackDBAReporting (pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetWholeONTDBAReporting(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetUpperOpticalThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0xff
	*pos++
	return pkt, nil
}

func GetSFThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x03
	*pos++
	return pkt, nil
}

func GetSDThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x05
	*pos++
	return pkt, nil
}

func GetARC(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetARCInterval(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetONTResponseTime(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x00
	*pos++
	pkt[*pos] = 0x00
	*pos++
	return pkt, nil
}

func GetLowerOpticalThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0xff
	*pos++
	return pkt, nil
}

func GetTransmitOpticalLeval(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x07
	*pos++
	pkt[*pos] = 0x1e
	*pos++
	return pkt, nil
}

func GetLowerTransmitPowerThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x81
	*pos++
	return pkt, nil
}

func GetUpperTransmitPowerThreshold(pos *uint, pkt []byte) ([]byte, error) {
	pkt[*pos] = 0x81
	*pos++
	return pkt, nil
}
