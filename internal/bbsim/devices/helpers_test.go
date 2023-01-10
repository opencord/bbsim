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

	"github.com/looplab/fsm"
	"gotest.tools/assert"
)

var (
	originalNewFSM func(initial string, events []fsm.EventDesc, callbacks map[string]fsm.Callback) *fsm.FSM
)

func setUpHelpers() {
	originalNewFSM = newFSM
}

func tearDownHelpers() {
	newFSM = originalNewFSM
}

func Test_Helpers(t *testing.T) {

	setUpHelpers()

	// feedback values for the mock
	called := 0
	args := struct {
		initial   string
		events    []fsm.EventDesc
		callbacks map[string]fsm.Callback
	}{}

	// creating the mock function
	mockFSM := func(initial string, events []fsm.EventDesc, callbacks map[string]fsm.Callback) *fsm.FSM {
		called++
		args.initial = initial
		args.events = events
		args.callbacks = callbacks
		return fsm.NewFSM(initial, events, callbacks)
	}
	newFSM = mockFSM

	// params for the method under test
	cb_called := 0
	cb := func(e *fsm.Event) {
		cb_called++
	}

	// calling the method under test
	sm := getOperStateFSM(cb)

	// verify
	assert.Equal(t, called, 1, "Expected fsm.NewFSM to have been called once, instead it was called %d", called)
	assert.Equal(t, args.initial, "down")

	assert.Equal(t, args.events[0].Name, "enable")
	assert.Equal(t, args.events[0].Src[0], "down")
	assert.Equal(t, args.events[0].Dst, "up")

	assert.Equal(t, args.events[1].Name, "disable")
	assert.Equal(t, args.events[1].Src[0], "up")
	assert.Equal(t, args.events[1].Dst, "down")

	// this is to test that the callback is called when the state change
	_ = sm.Event("enable")
	assert.Equal(t, cb_called, 1)

	tearDownHelpers()

}
