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

type PerformanceMonitoringHistoryData int

const (
	_                                                                = iota
	IntervalEndTime                 PerformanceMonitoringHistoryData = 0x8000
	ThresholdDataId                 PerformanceMonitoringHistoryData = 0x4000
	FCSErrors                       PerformanceMonitoringHistoryData = 0x2000
	ExcessiveCollisionCounter       PerformanceMonitoringHistoryData = 0x1000
	LateCollisionCounter            PerformanceMonitoringHistoryData = 0x0800
	FrameTooLong                    PerformanceMonitoringHistoryData = 0x0400
	BufferOverflowOnReceive         PerformanceMonitoringHistoryData = 0x0200
	BufferOverflowOnTransmit        PerformanceMonitoringHistoryData = 0x0100
	SingleCollisionFrameCounter     PerformanceMonitoringHistoryData = 0x0080
	MultipleCollisionFrameCounter   PerformanceMonitoringHistoryData = 0x0040
	SQECounter                      PerformanceMonitoringHistoryData = 0x0020
	DeferredTransmissionCounter     PerformanceMonitoringHistoryData = 0x0010
	InternalMACTransmitErrorCounter PerformanceMonitoringHistoryData = 0x0008
	CarrierSenseErrorCounter        PerformanceMonitoringHistoryData = 0x0004
	AllignmentErrorCounter          PerformanceMonitoringHistoryData = 0x0002
	InternalMACReceiveErrorCounter  PerformanceMonitoringHistoryData = 0x0001
)

type PMHistoryAttributeHandler func(*uint, []byte) ([]byte, error)

var PMHistoryAttributeHandlers = map[PerformanceMonitoringHistoryData]ANIGAttributeHandler{
	IntervalEndTime :                GetIntervalEndTime,
	ThresholdDataId:                 GetThresholdDataId,
	FCSErrors:                       GetFCSErrors,
	ExcessiveCollisionCounter:       GetExcessiveCollisionCounter,
	LateCollisionCounter:            GetLateCollisionCounter,
	FrameTooLong:                    GetFrameTooLong,
	BufferOverflowOnReceive:         GetBufferOverflowOnReceive,
	BufferOverflowOnTransmit:        GetBufferOverflowOnTransmit,
	SingleCollisionFrameCounter:     GetSingleCollisionFrameCounter,
	MultipleCollisionFrameCounter:   GetMultipleCollisionFrameCounter,
	SQECounter:                      GetSQECounter,
	DeferredTransmissionCounter:     GetDeferredTransmissionCounter,
	InternalMACTransmitErrorCounter: GetInternalMACTransmitErrorCounter,
	CarrierSenseErrorCounter:        GetCarrierSenseErrorCounter,
	AllignmentErrorCounter:          GetAllignmentErrorCounter,
	InternalMACReceiveErrorCounter:  GetInternalMACReceiveErrorCounter,
}

func GetEthernetPMHistoryDataAttributes(pos *uint, pkt []byte, content OmciContent) ([]byte, error) {
	AttributesMask := getAttributeMask(content)

	for index := uint(16); index>=1 ; index-- {
		Attribute := 1 << (index - 1)
		reqAttribute := Attribute & AttributesMask

		if reqAttribute != 0 {
			pkt, _ = PMHistoryAttributeHandlers[PerformanceMonitoringHistoryData(reqAttribute)](pos, pkt)
		}
	}

	pkt[8] = 0x00 // Command Processed Successfully
	pkt[9] = uint8(AttributesMask >> 8)
	pkt[10] = uint8(AttributesMask & 0x00FF)

	return pkt, nil

}

func GetIntervalEndTime(pos *uint, pkt []byte) ([]byte, error) {
	// With the hardware, it is seen that all attributes are 0x00
	// Nevertheless these functions are made to provide specific values in future if required
	*pos++
	return pkt, nil
}

func GetThresholdDataId(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	return pkt, nil
}

func GetFCSErrors(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetExcessiveCollisionCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetLateCollisionCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetFrameTooLong(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetBufferOverflowOnReceive(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetBufferOverflowOnTransmit(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetSingleCollisionFrameCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetMultipleCollisionFrameCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetSQECounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetDeferredTransmissionCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetInternalMACTransmitErrorCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetCarrierSenseErrorCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetAllignmentErrorCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}

func GetInternalMACReceiveErrorCounter(pos *uint, pkt []byte) ([]byte, error) {
	*pos++
	*pos++
	*pos++
	*pos++
	return pkt, nil
}


