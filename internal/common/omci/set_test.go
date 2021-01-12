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
	"github.com/google/gopacket"
	"github.com/opencord/omci-lib-go"
	"testing"
)

func omciToSetResponse(t *testing.T, omciPkt *gopacket.Packet) *omci.SetResponse {
	msgLayer := (*omciPkt).Layer(omci.LayerTypeSetResponse)
	if msgLayer == nil {
		t.Fatal("omci Msg layer could not be detected for SetResponse - handling of MibSyncChan stopped")
	}
	msgObj, msgOk := msgLayer.(*omci.SetResponse)
	if !msgOk {
		t.Fatal("omci Msg layer could not be assigned for SetResponse - handling of MibSyncChan stopped")
	}
	return msgObj
}
