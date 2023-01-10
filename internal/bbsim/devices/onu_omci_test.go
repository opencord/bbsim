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
	"strconv"
	"testing"

	"github.com/google/gopacket"
	bbsim "github.com/opencord/bbsim/internal/bbsim/types"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	"github.com/opencord/voltha-protos/v5/go/openolt"
	"gotest.tools/assert"
)

var mockAttr = me.AttributeValueMap{
	me.GemPortNetworkCtp_PortId:                                       0,
	me.GemPortNetworkCtp_TContPointer:                                 0,
	me.GemPortNetworkCtp_Direction:                                    0,
	me.GemPortNetworkCtp_TrafficManagementPointerForUpstream:          0,
	me.GemPortNetworkCtp_TrafficDescriptorProfilePointerForUpstream:   0,
	me.GemPortNetworkCtp_PriorityQueuePointerForDownStream:            0,
	me.GemPortNetworkCtp_TrafficDescriptorProfilePointerForDownstream: 0,
	me.GemPortNetworkCtp_EncryptionKeyRing:                            0,
}

func makeOmciCreateRequest(t *testing.T) []byte {
	omciReq := &omci.CreateRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
		Attributes: mockAttr,
	}

	omciPkt, err := omcilib.Serialize(omci.CreateRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciSetRequest(t *testing.T) []byte {
	omciReq := &omci.SetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
		Attributes: mockAttr,
	}
	omciPkt, err := omcilib.Serialize(omci.SetRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciDeleteRequest(t *testing.T) []byte {
	omciReq := &omci.DeleteRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass:    me.GemPortNetworkCtpClassID,
			EntityInstance: 12,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.DeleteRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciMibResetRequest(t *testing.T) []byte {
	omciReq := &omci.MibResetRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.OnuDataClassID,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.MibResetRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciStartSoftwareDownloadRequest(t *testing.T) []byte {
	omciReq := &omci.StartSoftwareDownloadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.SoftwareImageClassID,
		},
		ImageSize:            31,
		NumberOfCircuitPacks: 1,
		WindowSize:           31,
		CircuitPacks:         []uint16{0},
	}
	omciPkt, err := omcilib.Serialize(omci.StartSoftwareDownloadRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciEndSoftwareDownloadRequest(t *testing.T, imageSize uint32, imageCrc uint32) []byte {
	omciReq := &omci.EndSoftwareDownloadRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.SoftwareImageClassID,
		},
		NumberOfInstances: 1,
		ImageInstances:    []uint16{0},
		ImageSize:         imageSize,
		CRC32:             imageCrc,
	}
	omciPkt, err := omcilib.Serialize(omci.EndSoftwareDownloadRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciActivateSoftwareRequest(t *testing.T) []byte {
	omciReq := &omci.ActivateSoftwareRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.SoftwareImageClassID,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.ActivateSoftwareRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciCommitSoftwareRequest(t *testing.T) []byte {
	omciReq := &omci.CommitSoftwareRequest{
		MeBasePacket: omci.MeBasePacket{
			EntityClass: me.SoftwareImageClassID,
		},
	}
	omciPkt, err := omcilib.Serialize(omci.CommitSoftwareRequestType, omciReq, 66)
	if err != nil {
		t.Fatal(err.Error())
	}

	omciPkt, _ = omcilib.HexEncode(omciPkt)

	return omciPkt
}

func makeOmciMessage(t *testing.T, onu *Onu, pkt []byte) bbsim.OmciMessage {
	omciPkt, omciMsg, err := omcilib.ParseOpenOltOmciPacket(pkt)
	if err != nil {
		t.Fatal(err.Error())
	}
	return bbsim.OmciMessage{
		OnuSN:   onu.SerialNumber,
		OnuID:   onu.ID,
		OmciPkt: omciPkt,
		OmciMsg: omciMsg,
	}
}

func omciBytesToMsg(t *testing.T, data []byte) (*omci.OMCI, *gopacket.Packet) {
	packet := gopacket.NewPacket(data, omci.LayerTypeOMCI, gopacket.NoCopy)
	if packet == nil {
		t.Fatal("could not decode rxMsg as OMCI")
	}
	omciLayer := packet.Layer(omci.LayerTypeOMCI)
	if omciLayer == nil {
		t.Fatal("could not decode omci layer")
	}
	omciMsg, ok := omciLayer.(*omci.OMCI)
	if !ok {
		t.Fatal("could not assign omci layer")
	}
	return omciMsg, &packet
}

func omciToCreateResponse(t *testing.T, omciPkt *gopacket.Packet) *omci.CreateResponse {
	msgLayer := (*omciPkt).Layer(omci.LayerTypeCreateResponse)
	if msgLayer == nil {
		t.Fatal("omci Msg layer could not be detected for CreateResponse - handling of MibSyncChan stopped")
	}
	msgObj, msgOk := msgLayer.(*omci.CreateResponse)
	if !msgOk {
		t.Fatal("omci Msg layer could not be assigned for CreateResponse - handling of MibSyncChan stopped")
	}
	return msgObj
}

func Test_MibDataSyncIncrease(t *testing.T) {
	onu := createTestOnu()

	assert.Equal(t, onu.MibDataSync, uint8(0))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// send a Create and check that MDS has been increased
	err := onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// send a Set and check that MDS has been increased
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciSetRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(2))

	// send a Delete and check that MDS has been increased
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciDeleteRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(3))

	// Start software download
	onu.InternalState.SetState(OnuStateEnabled)
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciStartSoftwareDownloadRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(4))

	// End software download
	onu.ImageSoftwareExpectedSections = 31
	onu.ImageSoftwareReceivedSections = 31 // we fake that we have received all the download section we expect
	onu.ImageSectionData = []byte{111, 114, 116, 116, 105, 116, 111, 114, 32, 113, 117, 105, 115, 46, 32, 86, 105, 118, 97, 109, 117, 115, 32, 110, 101, 99, 32, 108, 105, 98, 101}
	onu.InternalState.SetState(OnuStateImageDownloadInProgress)
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciEndSoftwareDownloadRequest(t, 31, 1523894119)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(5))

	// Activate software
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciActivateSoftwareRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(6))

	// Commit software
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCommitSoftwareRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(7))
}

func Test_MibDataSyncReset(t *testing.T) {
	onu := createMockOnu(1, 1)
	onu.MibDataSync = 192
	assert.Equal(t, onu.MibDataSync, uint8(192))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// create a GemPort and an AllocId for this ONU
	onu.PonPort.storeGemPort(1024, onu.SerialNumber)
	onu.PonPort.storeAllocId(1024, 1024, 0x8001, 1024, onu.SerialNumber)

	// send a MibReset
	err := onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciMibResetRequest(t)), stream)
	assert.NilError(t, err)

	// check that MDS has reset to 0
	assert.Equal(t, onu.MibDataSync, uint8(0))

	// check that GemPort and AllocId have been removed
	assert.Equal(t, len(onu.PonPort.AllocatedGemPorts), 0)
	assert.Equal(t, len(onu.PonPort.AllocatedAllocIds), 0)
}

func Test_MibDataSyncRotation(t *testing.T) {
	onu := createMockOnu(1, 1)
	onu.MibDataSync = 255
	assert.Equal(t, onu.MibDataSync, uint8(255))

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// send a request that increases the MDS, but once we're at 255 we should go back to 0 (8bit)
	err := onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciDeleteRequest(t)), stream)
	assert.NilError(t, err)
	assert.Equal(t, onu.MibDataSync, uint8(0))
}

func Test_MibDataSyncInvalidation(t *testing.T) {
	onu := createMockOnu(1, 1)
	onu.MibDataSync = 250
	assert.Equal(t, onu.MibDataSync, uint8(250))

	onu.InvalidateMibDataSync()

	// check if the MDS has been changed
	assert.Assert(t, onu.MibDataSync != uint8(250))

	// the MDS has to be between 1 and 255, since 0 is valid for a reset
	assert.Assert(t, onu.MibDataSync > 0)
	// assert.Assert(t, onu.MibDataSync <= 255) // This is always true since 'MibDataSync' is uint8
}

func Test_GemPortValidation(t *testing.T) {

	// setup
	onu := createMockOnu(1, 1)

	stream := &mockStream{
		Calls: make(map[int]*openolt.Indication),
	}

	// create a gem port via OMCI (gemPortId 12)
	err := onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)
	assert.NilError(t, err)

	// the first time we created the gemPort
	// the MDS should be incremented
	assert.Equal(t, stream.CallCount, 1)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// and the OMCI response status should be me.Success
	indication := stream.Calls[1].GetOmciInd()
	_, omciPkt := omciBytesToMsg(t, indication.Pkt)
	responseLayer := omciToCreateResponse(t, omciPkt)
	assert.Equal(t, responseLayer.Result, me.Success)

	// send a request to create the same gem port via OMCI (gemPortId 12)
	err = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciCreateRequest(t)), stream)
	assert.NilError(t, err)

	// this time the MDS should not be incremented
	assert.Equal(t, stream.CallCount, 2)
	assert.Equal(t, onu.MibDataSync, uint8(1))

	// and the OMCI response status should be me.ProcessingError
	_, omciPkt = omciBytesToMsg(t, stream.Calls[2].GetOmciInd().Pkt)
	responseLayer = omciToCreateResponse(t, omciPkt)
	assert.Equal(t, responseLayer.Result, me.ProcessingError)
}

func Test_OmciResponseRate(t *testing.T) {

	onu := createMockOnu(1, 1)

	for onu.OmciResponseRate = 0; onu.OmciResponseRate <= maxOmciMsgCounter; onu.OmciResponseRate++ {
		stream := &mockStream{
			Calls: make(map[int]*openolt.Indication),
		}
		//send ten OMCI requests and check if number of responses is only equal to onu.OmciResponseRate
		for i := 0; i < 10; i++ {
			// we are not checking the error as we're expecting them (some messages should be skipped)
			_ = onu.handleOmciRequest(makeOmciMessage(t, onu, makeOmciSetRequest(t)), stream)
		}
		assert.Equal(t, stream.CallCount, int(onu.OmciResponseRate))
	}
}

func Test_EndSoftwareDownloadRequestHandling(t *testing.T) {
	onu := createTestOnu()

	// test EndSoftwareDownloadRequest in case of abort
	onu.ImageSoftwareReceivedSections = 2
	imageCrc, _ := strconv.ParseInt("FFFFFFFF", 16, 64)
	msg := makeOmciMessage(t, onu, makeOmciEndSoftwareDownloadRequest(t, 0, uint32(imageCrc)))
	res := onu.handleEndSoftwareDownloadRequest(msg)
	assert.Equal(t, res, true)
	assert.Equal(t, onu.ImageSoftwareReceivedSections, uint32(0))

	// test EndSoftwareDownloadRequest if we received less sections than expected
	onu.ImageSoftwareExpectedSections = 2
	onu.ImageSoftwareReceivedSections = 1
	msg = makeOmciMessage(t, onu, makeOmciEndSoftwareDownloadRequest(t, 2, 2))
	res = onu.handleEndSoftwareDownloadRequest(msg)
	assert.Equal(t, res, false)

	// test CRC Mismatch
	onu.ImageSectionData = []byte{111, 114, 116, 116, 105, 116, 111, 114, 32, 113, 117, 105, 115, 46, 32, 86, 105, 118, 97, 109, 117, 115, 32, 110, 101, 99, 32, 108, 105, 98, 101}
	onu.ImageSoftwareExpectedSections = 2
	onu.ImageSoftwareReceivedSections = 2
	msg = makeOmciMessage(t, onu, makeOmciEndSoftwareDownloadRequest(t, 31, 12))
	res = onu.handleEndSoftwareDownloadRequest(msg)
	assert.Equal(t, res, false)

	// if it's a valid case then set the StandbyImageVersion
	onu.InDownloadImageVersion = "DownloadedImage"
	msg = makeOmciMessage(t, onu, makeOmciEndSoftwareDownloadRequest(t, 31, 1523894119))
	res = onu.handleEndSoftwareDownloadRequest(msg)
	assert.Equal(t, res, true)
	assert.Equal(t, onu.StandbyImageVersion, "DownloadedImage")
}
