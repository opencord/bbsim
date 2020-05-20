// Copyright 2018-present Open Networking Foundation
// SPDX-FileCopyrightText: 2019 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package devices

import (
	"gotest.tools/assert"
	"testing"
)

func Test_Onu_CreateOnu(t *testing.T) {

	olt := OltDevice{
		ID: 0,
	}
	pon := PonPort{
		ID:  1,
		Olt: &olt,
	}

	onu := CreateONU(&olt, pon, 1, 900, 900, true, false,0,false)

	assert.Equal(t, onu.Sn(), "BBSM00000101")
	assert.Equal(t, onu.STag, 900)
	assert.Equal(t, onu.CTag, 900)
	assert.Equal(t, onu.Auth, true)
	assert.Equal(t, onu.Dhcp, false)
	assert.Equal(t, onu.HwAddress.String(), "2e:60:70:00:01:01")
}