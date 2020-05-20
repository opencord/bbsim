// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package devices

import (
	"github.com/looplab/fsm"
	"gotest.tools/assert"
	"testing"
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
		return
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
	sm.Event("enable")
	assert.Equal(t, cb_called, 1)

	tearDownHelpers()

}
