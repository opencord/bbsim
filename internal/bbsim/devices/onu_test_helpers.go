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
	"context"
	"errors"
	"time"

	bbsim_common "github.com/opencord/bbsim/internal/common"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	"github.com/opencord/voltha-protos/v5/go/extension"
	log "github.com/sirupsen/logrus"

	"github.com/opencord/bbsim/internal/bbsim/types"
	"github.com/opencord/voltha-protos/v5/go/ext/config"

	"github.com/opencord/voltha-protos/v5/go/openolt"
	"github.com/opencord/voltha-protos/v5/go/tech_profile"
	"google.golang.org/grpc"
)

func init() {
	bbsim_common.SetLogLevel(log.StandardLogger(), "error", false)
}

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
func (s *mockClient) PerformGroupOperation(ctx context.Context, group *openolt.Group, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) DeleteGroup(ctx context.Context, group *openolt.Group, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) GetExtValue(ctx context.Context, in *openolt.ValueParam, opts ...grpc.CallOption) (*extension.ReturnValues, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) OnuItuPonAlarmSet(ctx context.Context, in *config.OnuItuPonAlarm, opts ...grpc.CallOption) (*openolt.Empty, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) GetLogicalOnuDistanceZero(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.OnuLogicalDistance, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}
func (s *mockClient) GetLogicalOnuDistance(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.OnuLogicalDistance, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}

func (s *mockClient) GetGemPortStatistics(ctx context.Context, in *openolt.OnuPacket, opts ...grpc.CallOption) (*openolt.GemPortStatistics, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}

func (s *mockClient) GetOnuStatistics(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.OnuStatistics, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}

func (s *mockClient) GetPonRxPower(ctx context.Context, in *openolt.Onu, opts ...grpc.CallOption) (*openolt.PonRxPowerData, error) {
	return nil, errors.New("unimplemented-in-mock-client")
}

// this method creates a fake ONU used in the tests
func createMockOnu(id uint32, ponPortId uint32) *Onu {
	o := Onu{
		ID:        id,
		PonPortID: ponPortId,
		PonPort: &PonPort{
			AllocatedGemPorts: make(map[uint16]*openolt.SerialNumber),
			AllocatedAllocIds: make(map[AllocIDKey]*AllocIDVal),
			Olt:               &OltDevice{},
		},
		OmciResponseRate: 10,
		OmciMsgCounter:   0,
	}
	o.SerialNumber = NewSN(0, ponPortId, o.ID)
	o.Channel = make(chan types.Message, 10)

	unis := []UniPortIf{
		&UniPort{ID: 0, Onu: &o, PortNo: 16, MeId: omcilib.GenerateUniPortEntityId(1), logger: uniLogger},
		&UniPort{ID: 1, Onu: &o, PortNo: 17, MeId: omcilib.GenerateUniPortEntityId(2), logger: uniLogger},
		&UniPort{ID: 2, Onu: &o, PortNo: 18, MeId: omcilib.GenerateUniPortEntityId(3), logger: uniLogger},
		&UniPort{ID: 3, Onu: &o, PortNo: 19, MeId: omcilib.GenerateUniPortEntityId(4), logger: uniLogger},
	}

	o.UniPorts = unis
	return &o
}

// this method creates a real ONU to be used in the tests
func createTestOnu() *Onu {
	nextCtag := map[string]int{}
	nextStag := map[string]int{}

	olt := OltDevice{
		ID:               0,
		OmciResponseRate: 10,
	}

	pon := CreatePonPort(&olt, 1, bbsim_common.XGSPON)

	onu := CreateONU(&olt, pon, 1, time.Duration(1*time.Millisecond), nextCtag, nextStag, true)
	// NOTE we need this in order to create the OnuChannel
	_ = onu.InternalState.Event(OnuTxInitialize)
	onu.DiscoveryRetryDelay = 100 * time.Millisecond
	return onu
}
