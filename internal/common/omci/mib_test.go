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

package omci

import (
	"github.com/opencord/omci-lib-go"
	me "github.com/opencord/omci-lib-go/generated"
	"gotest.tools/assert"
	"testing"
)

func TestCreateMibResetResponse(t *testing.T) {
	data, _ := CreateMibResetResponse(1)

	omciMsg, omciPkt := omciBytesToMsg(t, data)

	assert.Equal(t, omciMsg.MessageType, omci.MibResetResponseType)

	msgLayer := (*omciPkt).Layer(omci.LayerTypeMibResetResponse)
	msgObj, msgOk := msgLayer.(*omci.MibResetResponse)
	if !msgOk {
		t.Fail()
	}

	assert.Equal(t, msgObj.Result, me.Success)
}

func TestCreateMibUploadNextResponse(t *testing.T) {
	//type args struct {
	//	omciPkt gopacket.Packet
	//	omciMsg *omci.OMCI
	//}
	//
	//tests := []struct {
	//	name    string
	//	args    args
	//	want    string
	//}{
	//	{"mibUploadNext-1", args{omciPkt: "localhost", omciMsg: ""}, "00:00:7f:00:00:01"},
	//}

	mibUploadNext, _ := CreateMibUploadNextRequest(1, 0)
	mibUploadNext = hexDecode(mibUploadNext)
	mibUploadNextMsg, mibUploadNextPkt := omciBytesToMsg(t, mibUploadNext)

	data, _ := CreateMibUploadNextResponse(*mibUploadNextPkt, mibUploadNextMsg)
	omciMsg, omciPkt := omciBytesToMsg(t, data)

	assert.Equal(t, omciMsg.MessageType, omci.MibUploadNextResponseType)

	msgLayer := (*omciPkt).Layer(omci.LayerTypeMibUploadNextResponse)
	msgObj, msgOk := msgLayer.(*omci.MibUploadNextResponse)
	if !msgOk {
		t.Fail()
	}

	assert.Equal(t, msgObj.ReportedME.GetName(), "OnuData")
	assert.Equal(t, msgObj.EntityClass, me.OnuDataClassID)
	mds, _ := msgObj.ReportedME.GetAttribute("MibDataSync")
	assert.Equal(t, mds, uint8(0))

}
