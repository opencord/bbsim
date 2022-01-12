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

package devices

import (
	"testing"

	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/stretchr/testify/assert"
)

func createTestPotsUni() (*PotsUniPort, error) {
	onu := &Onu{
		ID:           0,
		SerialNumber: NewSN(1, 1, 1),
		PonPortID:    0,
		Channel:      make(chan bbsimTypes.Message),
	}

	uni, err := NewPotsUniPort(1, onu)

	return uni, err
}

func Test_PotsUni_Create(t *testing.T) {
	uni, err := createTestPotsUni()

	assert.Nil(t, err)
	assert.NotNil(t, uni)

	// check initial operational state
	assert.Equal(t, uni.OperState.Current(), PotsUniStateDown)
}

func Test_PotsUni_OperationalState(t *testing.T) {
	uni, _ := createTestPotsUni()

	assert.Equal(t, uni.OperState.Current(), PotsUniStateDown)

	go uni.Enable()
	msg, ok := <-uni.Onu.Channel
	assert.True(t, ok)
	assert.Equal(t, uint8(0), msg.Data.(bbsimTypes.UniStatusAlarmMessage).AdminState)

	assert.Equal(t, PotsUniStateUp, uni.OperState.Current())

	go uni.Disable()
	msg, ok = <-uni.Onu.Channel
	assert.True(t, ok)
	assert.Equal(t, uint8(1), msg.Data.(bbsimTypes.UniStatusAlarmMessage).AdminState)

	assert.Equal(t, PotsUniStateDown, uni.OperState.Current())
}
