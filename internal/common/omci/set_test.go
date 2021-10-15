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
	"github.com/opencord/omci-lib-go/v2"
	me "github.com/opencord/omci-lib-go/v2/generated"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetRequest(t *testing.T) {

	meId := GenerateUniPortEntityId(1)

	meParams := me.ParamData{
		EntityID: meId.ToUint16(),
		Attributes: me.AttributeValueMap{
			"AdministrativeState": 1,
		},
	}
	meInstance, omciError := me.NewPhysicalPathTerminationPointEthernetUni(meParams)
	if omciError.GetError() != nil {
		t.Fatal(omciError.GetError())
	}

	pkt, err := CreateSetRequest(meInstance, 1)
	assert.NoError(t, err)

	omciPkt, omciMsg, err := ParseOpenOltOmciPacket(pkt)
	assert.NoError(t, err)
	assert.Equal(t, omciMsg.MessageType, omci.SetRequestType)

	msgObj, _ := ParseSetRequest(omciPkt)

	assert.Equal(t, meId.ToUint16(), msgObj.EntityInstance)
	assert.Equal(t, uint8(1), msgObj.Attributes["AdministrativeState"])

}
