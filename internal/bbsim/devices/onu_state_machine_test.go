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

	"gotest.tools/assert"
)

func Test_Onu_StateMachine_enable(t *testing.T) {
	onu := createTestOnu()
	assert.Equal(t, onu.InternalState.Current(), "initialized")
	_ = onu.InternalState.Event("discover")
	assert.Equal(t, onu.InternalState.Current(), "discovered")
	_ = onu.InternalState.Event("enable")
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)
}

func Test_Onu_StateMachine_disable(t *testing.T) {
	onu := createTestOnu()
	onu.InternalState.SetState(OnuStateEnabled)
	assert.Equal(t, onu.InternalState.Current(), OnuStateEnabled)

	onu.PortNo = 16
	onu.Flows = []FlowKey{
		{ID: 1, Direction: "upstream"},
		{ID: 2, Direction: "downstream"},
	}

	_ = onu.InternalState.Event(OnuTxDisable)
	assert.Equal(t, onu.InternalState.Current(), OnuStateDisabled)

	assert.Equal(t, onu.PortNo, uint32(0))
	assert.Equal(t, len(onu.Flows), 0)
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
	err := onu.InternalState.Event("start_auth")
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
	err := onu.InternalState.Event("start_auth")
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
	_ = onu.InternalState.Event("start_auth")
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

func Test_Onu_StateMachine_eapol_states(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState("auth_started")

	assert.Equal(t, onu.InternalState.Current(), "auth_started")
	_ = onu.InternalState.Event("eap_start_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_start_sent")
	_ = onu.InternalState.Event("eap_response_identity_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_identity_sent")
	_ = onu.InternalState.Event("eap_response_challenge_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_challenge_sent")
	_ = onu.InternalState.Event("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

	// test that we can retrigger EAPOL
	states := []string{"eap_start_sent", "eap_response_identity_sent", "eap_response_challenge_sent", "eap_response_success_received", "auth_failed", "dhcp_ack_received", "dhcp_failed"}
	for _, state := range states {
		onu.InternalState.SetState(state)
		err := onu.InternalState.Event("start_auth")
		assert.Equal(t, err, nil)
		assert.Equal(t, onu.InternalState.Current(), "auth_started")
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

	onu.InternalState.SetState("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

	err := onu.InternalState.Event("start_dhcp")
	if err == nil {
		t.Fail()
	}
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-dhcp-flow-is-missing")
}

// if the ONU does not have a GemPort we can't start DHCP
func Test_Onu_StateMachine_dhcp_no_gem(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

	err := onu.InternalState.Event("start_dhcp")
	if err == nil {
		t.Fail()
	}
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-gemport-is-missing")
}

func Test_Onu_StateMachine_dhcp_start(t *testing.T) {
	t.Skip("Needs to be moved in the Service struct")
	onu := createTestOnu()

	onu.InternalState.SetState("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

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
	states := []string{"eap_response_success_received", "dhcp_discovery_sent", "dhcp_request_sent", "dhcp_ack_received", "dhcp_failed"}
	for _, state := range states {
		onu.InternalState.SetState(state)
		err := onu.InternalState.Event("start_dhcp")
		assert.Equal(t, err, nil)
		assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	}
}
