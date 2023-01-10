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
	"testing"

	bbsimTypes "github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/stretchr/testify/assert"
)

func createTestPots() (*PotsPort, error) {
	onu := &Onu{
		ID:           0,
		SerialNumber: NewSN(1, 1, 1),
		PonPortID:    0,
		Channel:      make(chan bbsimTypes.Message),
	}

	pots, err := NewPotsPort(1, onu)

	return pots, err
}

func Test_Pots_Create(t *testing.T) {
	pots, err := createTestPots()

	assert.Nil(t, err)
	assert.NotNil(t, pots)

	// check initial operational state
	assert.Equal(t, pots.OperState.Current(), PotsStateDown)
}

func Test_Pots_OperationalState(t *testing.T) {
	pots, _ := createTestPots()

	assert.Equal(t, pots.OperState.Current(), PotsStateDown)

	go func() {
		//When it will be enabled
		msg, ok := <-pots.Onu.Channel
		assert.True(t, ok)
		assert.Equal(t, uint8(0), msg.Data.(bbsimTypes.UniStatusAlarmMessage).AdminState)

		//When it will be disabled
		msg, ok = <-pots.Onu.Channel
		assert.True(t, ok)
		assert.Equal(t, uint8(1), msg.Data.(bbsimTypes.UniStatusAlarmMessage).AdminState)
	}()

	err := pots.Enable()
	assert.Nil(t, err)
	assert.Equal(t, PotsStateUp, pots.OperState.Current())

	err = pots.Disable()
	assert.Nil(t, err)
	assert.Equal(t, PotsStateDown, pots.OperState.Current())
}
