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
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"github.com/opencord/voltha-protos/v2/go/tech_profile"
	"google.golang.org/grpc"
	"net"
	"time"
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

// this method creates a fake ONU used in the tests
func createMockOnu(id uint32, ponPortId uint32, sTag int, cTag int, auth bool, dhcp bool) *Onu {
	o := Onu{
		ID:           id,
		PonPortID:    ponPortId,
		STag:         sTag,
		CTag:         cTag,
		HwAddress:    net.HardwareAddr{0x2e, 0x60, 0x70, 0x13, byte(ponPortId), byte(id)},
		PortNo:       0,
		Auth:         auth,
		Dhcp:         dhcp,
		GemPortAdded: true,
	}
	o.SerialNumber = o.NewSN(0, ponPortId, o.ID)
	return &o
}

// this method creates a real ONU to be used in the tests
func createTestOnu() *Onu {
	olt := OltDevice{
		ID: 0,
	}
	pon := PonPort{
		ID:  1,
		Olt: &olt,
	}
	onu := CreateONU(&olt, &pon, 1, 900, 900, false, false, time.Duration(1*time.Millisecond), true)
	// NOTE we need this in order to create the OnuChannel
	onu.InternalState.Event("initialize")
	onu.DiscoveryRetryDelay = 100 * time.Millisecond
	return onu
}
