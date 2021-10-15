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

func TestEntityID_ToUint16(t *testing.T) {
	var id EntityID
	var res uint16

	id = EntityID{0x01, 0x01}
	res = id.ToUint16()
	assert.Equal(t, uint16(257), res)

	id = EntityID{0x00, 0x00}
	res = id.ToUint16()
	assert.Equal(t, uint16(0), res)
}

func TestEntityID_FromUint16(t *testing.T) {
	id := uint16(257)
	e := EntityID{}.FromUint16(id)

	assert.Equal(t, e.ToString(), "0101")
	assert.Equal(t, e.ToUint16(), id)
}

func TestEntityID_Equals(t *testing.T) {
	a := EntityID{0x01, 0x01}
	b := EntityID{0x01, 0x01}
	c := EntityID{0x01, 0x02}

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}

func Test_GenerateMibDatabase(t *testing.T) {
	const uniPortCount = 4
	mibDb, err := GenerateMibDatabase(uniPortCount)

	expectedItems := 9                     //ONU-G + 2 Circuit Packs (4 messages each)
	expectedItems += 2 * uniPortCount      // 1 PPTP and 1 UniG per UNI
	expectedItems += 1                     // ANI-G
	expectedItems += 2 * tconts            // T-CONT and traffic schedulers
	expectedItems += 1                     // ONU-2g
	expectedItems += 2 * 8 * tconts        // 8 upstream queues for each T-CONT, and we report each queue twice
	expectedItems += 2 * 16 * uniPortCount // 16 downstream queues for each T-CONT, and we report each queue twice

	assert.NoError(t, err)
	assert.NotNil(t, mibDb)
	assert.Equal(t, expectedItems, int(mibDb.NumberOfCommands))

	// now try to serialize all messages to check on the attributes
	for _, entry := range mibDb.items {
		reportedMe, meErr := me.LoadManagedEntityDefinition(entry.classId, me.ParamData{
			EntityID:   entry.entityId.ToUint16(),
			Attributes: entry.params,
		})
		assert.NoError(t, meErr.GetError())

		response := &omci.MibUploadNextResponse{
			MeBasePacket: omci.MeBasePacket{
				EntityClass: me.OnuDataClassID,
			},
			ReportedME: *reportedMe,
		}

		_, err := Serialize(omci.MibUploadNextResponseType, response, uint16(10))
		assert.NoError(t, err)
	}

}
