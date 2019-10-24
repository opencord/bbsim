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
	"context"
	"errors"
	"github.com/google/gopacket/layers"
	"github.com/looplab/fsm"
	"github.com/opencord/voltha-protos/go/openolt"
	"github.com/opencord/voltha-protos/go/tech_profile"
	"google.golang.org/grpc"
	"gotest.tools/assert"
	"net"
	"testing"
)

type FlowAddSpy struct {
	CallCount int
	Calls     map[int]*openolt.Flow
}

type mockClient struct {
	FlowAddSpy
	fail bool
}

func (s *mockClient) DisableOlt(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) ReenableOlt(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) ActivateOnu(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) DeactivateOnu(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) DeleteOnu(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) OmciMsgOut(ctx context.Context, in *openolt.OmciMsg, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) OnuPacketOut(ctx context.Context, in *openolt.OnuPacket, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) UplinkPacketOut(ctx context.Context, in *openolt.UplinkPacket, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) FlowAdd(ctx context.Context, in *openolt.Flow, opts ...grpc.CallOption) (*openolt.Empty, error) {
	s.FlowAddSpy.CallCount++
	if s.fail {
		return nil, errors.New("fake-error")
	}
	s.FlowAddSpy.Calls[s.FlowAddSpy.CallCount] = in
	return &openolt.Empty{}, nil
}
func (s *mockClient) FlowRemove(ctx context.Context, in *openolt.Flow, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) HeartbeatCheck(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.Heartbeat, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) EnablePonIf(ctx context.Context, in *openolt.Interface, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) DisablePonIf(ctx context.Context, in *openolt.Interface, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) GetDeviceInfo(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.DeviceInfo, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) Reboot(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) CollectStatistics(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) CreateTrafficSchedulers(ctx context.Context, in *tech_profile.TrafficSchedulers, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) RemoveTrafficSchedulers(ctx context.Context, in *tech_profile.TrafficSchedulers, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) CreateTrafficQueues(ctx context.Context, in *tech_profile.TrafficQueues, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) RemoveTrafficQueues(ctx context.Context, in *tech_profile.TrafficQueues, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) EnableIndication(ctx context.Context, in *openolt.Empty, opts ...grpc.CallOption) (openolt.Openolt_EnableIndicationClient, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}

func createMockOnu(id uint32, ponPortId uint32, sTag int, cTag int) Onu {
	o := Onu{
		ID:        id,
		PonPortID: ponPortId,
		STag:      sTag,
		CTag:      cTag,
		HwAddress: net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(ponPortId), byte(id)},
		PortNo:    0,
	}
	o.SerialNumber = o.NewSN(0, ponPortId, o.ID)
	return o
}

func Test_Onu_SendEapolFlow(t *testing.T) {
	onu := createMockOnu(1, 1, 900, 900)

	client := &mockClient{
		FlowAddSpy: FlowAddSpy{
			Calls: make(map[int]*openolt.Flow),
		},
		fail: false,
	}

	onu.sendEapolFlow(client)
	assert.Equal(t, client.FlowAddSpy.CallCount, 1)

	assert.Equal(t, client.FlowAddSpy.Calls[1].AccessIntfId, int32(onu.PonPortID))
	assert.Equal(t, client.FlowAddSpy.Calls[1].OnuId, int32(onu.ID))
	assert.Equal(t, client.FlowAddSpy.Calls[1].UniId, int32(0))
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowId, onu.ID)
	assert.Equal(t, client.FlowAddSpy.Calls[1].FlowType, "downstream")
	assert.Equal(t, client.FlowAddSpy.Calls[1].PortNo, onu.ID)
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// it transition to auth_started state
func Test_HandleFlowUpdateEapolFromGem(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900)

	onu.InternalState = fsm.NewFSM(
		"gem_port_added",
		fsm.Events{
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken
func Test_HandleFlowUpdateEapolFromGemIgnore(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900)

	onu.InternalState = fsm.NewFSM(
		"gem_port_added",
		fsm.Events{
			{Name: "start_auth", Src: []string{"eapol_flow_received", "gem_port_added"}, Dst: "auth_started"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(1),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "gem_port_added")
}

// validates that when an ONU receives an EAPOL flow for UNI 0
// it transition to auth_started state
func Test_HandleFlowUpdateEapolFromEnabled(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(0),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "eapol_flow_received")
}

// validates that when an ONU receives an EAPOL flow for UNI that is not 0
// no action is taken
func Test_HandleFlowUpdateEapolFromEnabledIgnore(t *testing.T) {

	onu := createMockOnu(1, 1, 900, 900)

	onu.InternalState = fsm.NewFSM(
		"enabled",
		fsm.Events{
			{Name: "receive_eapol_flow", Src: []string{"enabled", "gem_port_added"}, Dst: "eapol_flow_received"},
		},
		fsm.Callbacks{},
	)

	flow := openolt.Flow{
		AccessIntfId:  int32(onu.PonPortID),
		OnuId:         int32(onu.ID),
		UniId:         int32(1),
		FlowId:        uint32(onu.ID),
		FlowType:      "downstream",
		AllocId:       int32(0),
		NetworkIntfId: int32(0),
		Classifier: &openolt.Classifier{
			EthType: uint32(layers.EthernetTypeEAPOL),
			OVid:    4091,
		},
		Action:   &openolt.Action{},
		Priority: int32(100),
		PortNo:   uint32(onu.ID), // NOTE we are using this to map an incoming packetIndication to an ONU
	}

	msg := OnuFlowUpdateMessage{
		PonPortID: 1,
		OnuID:     1,
		Flow:      &flow,
	}

	onu.handleFlowUpdate(msg)
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

// TODO add tests for DHCP flow
