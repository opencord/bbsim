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
	"gotest.tools/assert"
	"testing"
)

func Test_Onu_StateMachine_enable(t *testing.T) {
	onu := createTestOnu()

	assert.Equal(t, onu.InternalState.Current(), "created")
	onu.InternalState.Event("discover")
	assert.Equal(t, onu.InternalState.Current(), "discovered")
	onu.InternalState.Event("enable")
	assert.Equal(t, onu.InternalState.Current(), "enabled")
}

func Test_Onu_StateMachine_eapol_start_eap_flow(t *testing.T) {
	onu := createTestOnu()

	onu.InternalState.SetState("enabled")

	// TODO we need to add a check so that you can't go from eapol_flow_received
	// to auth_started without passing through gem_port_added
	// (see start_dhcp for an example)

	assert.Equal(t, onu.InternalState.Current(), "enabled")
	onu.InternalState.Event("receive_eapol_flow")
	assert.Equal(t, onu.InternalState.Current(), "eapol_flow_received")
	onu.InternalState.Event("add_gem_port")
	assert.Equal(t, onu.InternalState.Current(), "gem_port_added")
	onu.InternalState.Event("start_auth")
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

func Test_Onu_StateMachine_eapol_start_gem_port(t *testing.T) {
	onu := createTestOnu()

	onu.InternalState.SetState("enabled")

	// TODO we need to add a check so that you can't go from gem_port_added
	// to auth_started without passing through eapol_flow_received
	// (see start_dhcp for an example)

	assert.Equal(t, onu.InternalState.Current(), "enabled")
	onu.InternalState.Event("add_gem_port")
	assert.Equal(t, onu.InternalState.Current(), "gem_port_added")
	onu.InternalState.Event("receive_eapol_flow")
	assert.Equal(t, onu.InternalState.Current(), "eapol_flow_received")
	onu.InternalState.Event("start_auth")
	assert.Equal(t, onu.InternalState.Current(), "auth_started")
}

func Test_Onu_StateMachine_eapol_states(t *testing.T) {
	onu := createTestOnu()

	onu.InternalState.SetState("auth_started")

	assert.Equal(t, onu.InternalState.Current(), "auth_started")
	onu.InternalState.Event("eap_start_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_start_sent")
	onu.InternalState.Event("eap_response_identity_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_identity_sent")
	onu.InternalState.Event("eap_response_challenge_sent")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_challenge_sent")
	onu.InternalState.Event("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
}

func Test_Onu_StateMachine_dhcp_start(t *testing.T) {
	onu := createTestOnu()
	onu.DhcpFlowReceived = true

	onu.InternalState.SetState("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

	onu.InternalState.Event("start_dhcp")

	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
}

func Test_Onu_StateMachine_dhcp_start_error(t *testing.T) {
	onu := createTestOnu()

	onu.InternalState.SetState("eap_response_success_received")
	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")

	err := onu.InternalState.Event("start_dhcp")

	assert.Equal(t, onu.InternalState.Current(), "eap_response_success_received")
	assert.Equal(t, err.Error(), "transition canceled with error: cannot-go-to-dhcp-started-as-dhcp-flow-is-missing")
}

func Test_Onu_StateMachine_dhcp_states(t *testing.T) {
	onu := createTestOnu()

	onu.DhcpFlowReceived = false

	onu.InternalState.SetState("dhcp_started")

	assert.Equal(t, onu.InternalState.Current(), "dhcp_started")
	onu.InternalState.Event("dhcp_discovery_sent")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_discovery_sent")
	onu.InternalState.Event("dhcp_request_sent")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_request_sent")
	onu.InternalState.Event("dhcp_ack_received")
	assert.Equal(t, onu.InternalState.Current(), "dhcp_ack_received")
}
