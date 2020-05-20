// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
// SPDX-FileCopyrightText: 2020 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

//https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking
package devices

import (
	"errors"
	"github.com/google/gopacket/pcap"
	"testing"

	"github.com/opencord/bbsim/internal/bbsim/types"
	"gotest.tools/assert"
)

func TestSetVethUpSuccess(t *testing.T) {
	spy := &ExecutorSpy{
		Calls: make(map[int][]string),
	}
	err := setVethUp(spy, "test_veth")
	assert.Equal(t, spy.CommandCallCount, 1)
	assert.Equal(t, spy.Calls[1][2], "test_veth")
	assert.Equal(t, err, nil)
}

func TestSetVethUpFail(t *testing.T) {
	spy := &ExecutorSpy{
		failRun: true,
		Calls:   make(map[int][]string),
	}
	err := setVethUp(spy, "test_veth")
	assert.Equal(t, spy.CommandCallCount, 1)
	assert.Equal(t, err.Error(), "fake-error")
}

func TestCreateNNIPair(t *testing.T) {

	startDHCPServerCalled := false
	_startDHCPServer := startDHCPServer
	defer func() { startDHCPServer = _startDHCPServer }()
	startDHCPServer = func(upstreamVeth string, dhcpServerIp string) error {
		startDHCPServerCalled = true
		return nil
	}

	listenOnVethCalled := false
	_listenOnVeth := listenOnVeth
	defer func() { listenOnVeth = _listenOnVeth }()
	listenOnVeth = func(vethName string) (chan *types.PacketMsg, *pcap.Handle, error) {
		listenOnVethCalled = true
		return make(chan *types.PacketMsg, 1), nil, nil
	}
	spy := &ExecutorSpy{
		failRun: false,
		Calls:   make(map[int][]string),
	}

	olt := OltDevice{}
	nni := NniPort{}

	err := createNNIPair(spy, &olt, &nni)
	olt.nniPktInChannel, olt.nniHandle, _ = nni.NewVethChan()

	assert.Equal(t, spy.CommandCallCount, 3)
	assert.Equal(t, startDHCPServerCalled, true)
	assert.Equal(t, listenOnVethCalled, true)
	assert.Equal(t, err, nil)
	assert.Assert(t, olt.nniPktInChannel != nil)
}

type ExecutorSpy struct {
	failRun bool

	CommandCallCount int
	RunCallCount     int
	Calls            map[int][]string
}

func (s *ExecutorSpy) Command(name string, arg ...string) Runnable {
	s.CommandCallCount++

	s.Calls[s.CommandCallCount] = arg

	return s
}

func (s *ExecutorSpy) Run() error {
	s.RunCallCount++
	if s.failRun {
		return errors.New("fake-error")
	}
	return nil
}
