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

	"github.com/opencord/bbsim/internal/bbsim/responders/eapol"
	omcilib "github.com/opencord/bbsim/internal/common/omci"
	me "github.com/opencord/omci-lib-go/v2/generated"

	"gotest.tools/assert"
)

func Test_Onu_StateMachine_enable(t *testing.T) {
	onu := createTestOnu()
	assert.Equal(t, onu.InternalState.Current(), OnuStateInitialized)
	_ = onu.InternalState.Event(OnuTxDiscover)
	assert.Equal(t, onu.InternalState.Current(), OnuStateDiscovered)
	_ = onu.InternalState.Event(OnuTxEnable)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)
}

func Test_Onu_StateMachine_disable(t *testing.T) {
	onu := createTestOnu()
	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	onu.Flows = []FlowKey{
		{ID: 1},
		{ID: 2},
	}
	key := omcilib.OnuAlarmInfoMapKey{
		MeInstance: 257,
		MeClassID:  me.PhysicalPathTerminationPointEthernetUniClassID,
	}
	onu.onuAlarmsInfo[key] = omcilib.OnuAlarmInfo{SequenceNo: 1, AlarmBitMap: [28]byte{}}
	onu.PonPort.storeOnuId(onu.ID, onu.SerialNumber)
	onu.PonPort.storeAllocId(1, 1024, 0x8001, 1024, onu.SerialNumber)
	onu.PonPort.storeGemPort(1, onu.SerialNumber)

	_ = onu.InternalState.Event(OnuTxDisable)
	assert.Equal(t, onu.InternalState.Current(), OnuStateDisabled)

	assert.Equal(t, len(onu.onuAlarmsInfo), 0)
	assert.Equal(t, len(onu.Flows), 0)
	assert.Equal(t, len(onu.PonPort.AllocatedOnuIds), 0)
	assert.Equal(t, len(onu.PonPort.AllocatedAllocIds), 0)
	assert.Equal(t, len(onu.PonPort.AllocatedGemPorts), 0)
}

func Test_Onu_StateMachine_pon_disable(t *testing.T) {
	onu := createTestOnu()
	var err error

	onu.InternalState.SetState(OnuStateEnabled)
	err = onu.InternalState.Event(OnuTxPonDisable)
	assert.NilError(t, err)
	assert.Equal(t, OnuStatePonDisabled, onu.InternalState.Current())

	onu.InternalState.SetState(OnuStateImageDownloadComplete)
	err = onu.InternalState.Event(OnuTxPonDisable)
	assert.NilError(t, err)
	assert.Equal(t, OnuStatePonDisabled, onu.InternalState.Current())
}

func Test_Onu_StateMachine_software_image(t *testing.T) {
	onu := createTestOnu()
	var err error

	// happy path
	onu.InternalState.SetState(OnuStateEnabled)
	err = onu.InternalState.Event(OnuTxStartImageDownload)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageDownloadStarted, onu.InternalState.Current())

	err = onu.InternalState.Event(OnuTxProgressImageDownload)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageDownloadInProgress, onu.InternalState.Current())

	err = onu.InternalState.Event(OnuTxCompleteImageDownload)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageDownloadComplete, onu.InternalState.Current())

	err = onu.InternalState.Event(OnuTxActivateImage)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageActivated, onu.InternalState.Current())

	// after image activate we get an ONU reboot, thus the state is back to Enabled before committing
	onu.InternalState.SetState(OnuStateEnabled)
	err = onu.InternalState.Event(OnuTxCommitImage)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageCommitted, onu.InternalState.Current())

	// but we should be able to start a new download
	err = onu.InternalState.Event(OnuTxStartImageDownload)
	assert.NilError(t, err)
	assert.Equal(t, OnuStateImageDownloadStarted, onu.InternalState.Current())
}

// check that I can go to auth_started only if
// - the GemPort is set
// - the eapolFlow is received
func Test_Onu_StateMachine_eapol_no_flow(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	// fail as no EapolFlow has been received
	err := onu.InternalState.Event(eapol.EventStartAuth)
	if err == nil {
		t.Fatal("can't start EAPOL without EapolFlow")
	}
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-auth-started-as-eapol-flow-is-missing")
}

func Test_Onu_StateMachine_eapol_no_gem(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	// fail has no GemPort has been set
	err := onu.InternalState.Event(eapol.EventStartAuth)
	if err == nil {
		t.Fatal("can't start EAPOL without GemPort")
	}
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-auth-started-as-gemport-is-missing")

}

func Test_Onu_StateMachine_eapol_start(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	// succeed
	_ = onu.InternalState.Event(eapol.EventStartAuth)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateAuthStarted)
}

func Test_Onu_StateMachine_eapol_states(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(eapol.StateAuthStarted)

	assert.Equal(t, onu.InternalState.Current(), eapol.StateAuthStarted)
	_ = onu.InternalState.Event(eapol.EventStartSent)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateStartSent)
	_ = onu.InternalState.Event(eapol.EventResponseIdentitySent)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseIdentitySent)
	_ = onu.InternalState.Event(eapol.EventResponseChallengeSent)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseChallengeSent)
	_ = onu.InternalState.Event(eapol.EventResponseSuccessReceived)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)

	// test that we can retrigger EAPOL
	states := []string{eapol.StateStartSent, eapol.StateResponseIdentitySent, eapol.StateResponseChallengeSent, eapol.StateResponseSuccessReceived, eapol.StateAuthFailed, "dhcp_ack_received", "dhcp_failed"}
	for _, state := range states {
		onu.InternalState.SetState(state)
		err := onu.InternalState.Event(eapol.EventStartAuth)
		assert.Equal(t, err, nil)
		assert.Equal(t, onu.InternalState.Current(), eapol.StateAuthStarted)
	}
}

// if auth is set to true we can't go from enabled to dhcp_started
func Test_Onu_StateMachine_dhcp_no_auth(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	err := onu.InternalState.Event("start_dhcp")
	if err == nil {
		t.Fail()
	}
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-authentication-is-required")
}

// if the DHCP flow has not been received we can't start authentication
func Test_Onu_StateMachine_dhcp_no_flow(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(eapol.StateResponseSuccessReceived)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)

	err := onu.InternalState.Event("start_dhcp")
	if err == nil {
		t.Fail()
	}
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-dhcp-flow-is-missing")
}

// if the ONU does not have a GemPort we can't start DHCP
func Test_Onu_StateMachine_dhcp_no_gem(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(eapol.StateResponseSuccessReceived)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)

	err := onu.InternalState.Event("start_dhcp")
	if err == nil {
		t.Fail()
	}
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-gemport-is-missing")
}

func Test_Onu_StateMachine_dhcp_start(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState(eapol.StateResponseSuccessReceived)
	assert.Equal(t, onu.InternalState.Current(), eapol.StateResponseSuccessReceived)

	// default transition
	_ = onu.InternalState.Event("start_dhcp")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
}

func Test_Onu_StateMachine_dhcp_states(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState("dhcp_started")

	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	_ = onu.InternalState.Event("dhcp_discovery_sent")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_discovery_sent")
	_ = onu.InternalState.Event("dhcp_request_sent")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_request_sent")
	_ = onu.InternalState.Event("dhcp_ack_received")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_ack_received")

	// test that we can retrigger DHCP
	states := []string{eapol.StateResponseSuccessReceived, "dhcp_discovery_sent", "dhcp_request_sent", "dhcp_ack_received", "dhcp_failed"}
	for _, state := range states {
		onu.InternalState.SetState(state)
		err := onu.InternalState.Event("start_dhcp")
		assert.Equal(t, err, nil)
		assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	}
}
