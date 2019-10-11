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

//https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking
package devices

import (
	"errors"
	"gotest.tools/assert"
	"testing"
)

func TestSetVethUpSuccess(t *testing.T) {
	spy := &SetVethUpSpy{}
	err := setVethUp(spy, "test_veth")
	assert.Equal(t, spy.CommandCallCount, 1)
	assert.Equal(t, err, nil)
}

func TestSetVethUpFail(t *testing.T) {
	spy := &SetVethUpSpy{failRun: true}
	err := setVethUp(spy, "test_veth")
	assert.Equal(t, spy.CommandCallCount, 1)
	assert.Equal(t, err.Error(), "fake-error")
}

type SetVethUpSpy struct {
	failRun bool

	CommandCallCount int
	RunCallCount     int
	//Calls
}

func (s *SetVethUpSpy) Command(name string, arg ...string) Runnable {
	s.CommandCallCount++
	return s
}

func (s *SetVethUpSpy) Run() error {
	s.RunCallCount++
	if s.failRun {
		return errors.New("fake-error")
	}
	return nil
}
